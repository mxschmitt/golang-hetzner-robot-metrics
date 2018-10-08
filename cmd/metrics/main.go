package main

import (
	"encoding/json"
	"flag"
	"log"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/mxschmitt/golang-hetzner-robot-metrics/pkg/api"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var store = &Store{}

type Store struct {
	Data *api.LiveData
	sync.RWMutex
}

func handleGetServer(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	store.RLock()
	found := false
	for _, server := range store.Data.Server {
		if strconv.Itoa(server.Key) == ps.ByName("id") {
			json.NewEncoder(w).Encode(server)
			found = true
			break
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

	go func() {
		for {
			data, err := api.GetLiveData()
			if err != nil {
				log.Printf("could not get live data: %v", err)
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
			log.Printf("Crawled %d servers with hash %s", len(data.Server), data.Hash)
			// Sleep 10 minutes
			time.Sleep(10 * 60 * time.Second)
		}
	}()
	log.Printf("Listening on %s", *addr)
	router := httprouter.New()
	router.Handler("GET", "/metrics", promhttp.Handler())
	router.GET("/hetzner/server/:id", handleGetServer)
	log.Fatal(http.ListenAndServe(*addr, router))
}
