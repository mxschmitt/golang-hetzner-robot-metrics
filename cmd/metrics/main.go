package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/api"

	hetznerapi "github.com/mxschmitt/golang-hetzner-robot-metrics/pkg/api"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var store = &Store{}

type Store struct {
	Data *hetznerapi.LiveData
	sync.RWMutex
}

func handleGetServer(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	store.RLock()
	found := false
	if store.Data != nil {
		for _, server := range store.Data.Server {
			if strconv.Itoa(server.Key) == ps.ByName("id") {
				json.NewEncoder(w).Encode(server)
				found = true
				break
			}
		}
	}
	if !found {
		json.NewEncoder(w).Encode(nil)
	}
	store.RUnlock()
}

func main() {
	addr := flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	flag.Parse()

	hetznerServersGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hetzner_robot_servers_price",
		Help: "Hetzner Robot Server",
	}, []string{"key"})
	prometheus.MustRegister(hetznerServersGauge)

	removeOldServers := func(data *hetznerapi.LiveData) error {
		client, err := api.NewClient(api.Config{
			Address: "http://prometheus:9090",
		})
		if err != nil {
			return errors.Wrap(err, "could not create new prometheus api client")
		}
		prometheusAPI := v1.NewAPI(client)
		resp, err := prometheusAPI.LabelValues(context.Background(), "key")
		if err != nil {
			return errors.Wrap(err, "could not query to prometheus")
		}
		soldServersMatches := []string{}
		for _, label := range resp {
			contains := false
			for _, server := range data.Server {
				if strconv.Itoa(server.Key) == string(label) {
					contains = true
				}
			}
			if !contains {
				soldServersMatches = append(soldServersMatches, fmt.Sprintf(`hetzner_robot_servers_price{key="%s"}`, label))
			}
		}
		if len(soldServersMatches) > 0 {
			fmt.Printf("Deleting sold servers: %d\n", len(soldServersMatches))
			if err := prometheusAPI.DeleteSeries(context.Background(), soldServersMatches, time.Unix(0, 0), time.Now()); err != nil {
				return errors.Wrap(err, "could not delete sold servers")
			}
			fmt.Println("Deleted sold servers successfully")
			if err := prometheusAPI.CleanTombstones(context.Background()); err != nil {
				return errors.Wrap(err, "could not clear Tombstones")
			}
			fmt.Println("Cleaned Tombstones")
		}
		return nil
	}

	go func() {
		for {
			data, err := hetznerapi.GetLiveData()
			if err != nil {
				log.Printf("could not get live data: %v", err)
				continue
			}
			store.Lock()
			store.Data = data
			store.Unlock()
			for _, server := range data.Server {
				price, err := strconv.ParseFloat(server.Price, 64)
				if err != nil {
					log.Printf("could not parse price: %v", err)
					continue
				}
				price = math.Round(price * 1.19)
				hetznerServersGauge.WithLabelValues(strconv.Itoa(server.Key)).Set(price)
			}
			if err := removeOldServers(data); err != nil {
				log.Printf("could not remove old servers: %v", err)
			}
			log.Printf("Crawled %d servers with hash %s", len(data.Server), data.Hash)
			// Sleep 1 minute
			time.Sleep(60 * time.Second)
		}
	}()
	log.Printf("Listening on %s", *addr)
	router := httprouter.New()
	router.Handler("GET", "/metrics", promhttp.Handler())
	router.GET("/hetzner/server/:id", handleGetServer)
	log.Fatal(http.ListenAndServe(*addr, router))
}
