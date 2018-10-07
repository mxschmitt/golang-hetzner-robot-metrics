package main

import (
	"flag"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/mxschmitt/golang-hetzner-robot-metrics/pkg/api"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	hetznerServersCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "hetzner_robot_servers_price",
		Help: "Hetzner Robot Server",
	}, []string{"key"})
)

func init() {
	prometheus.MustRegister(hetznerServersCounter)
}

func main() {
	addr := flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	flag.Parse()
	go func() {
		for {
			api, err := api.GetLiveData()
			if err != nil {
				log.Printf("could not get live data: %v", err)
			}
			for _, server := range api.Server {
				price, err := strconv.ParseFloat(server.Price, 64)
				if err != nil {
					log.Printf("could not parse price: %v", err)
					continue
				}
				price = math.Round(price * 1.19)
				hetznerServersCounter.WithLabelValues(strconv.Itoa(server.Key)).Add(price)
			}
			log.Printf("Crawled %d servers with hash %s", len(api.Server), api.Hash)
			// Sleep 10 minutes
			time.Sleep(10 * 60 * time.Second)
		}
	}()
	log.Printf("Listening on %s", *addr)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
