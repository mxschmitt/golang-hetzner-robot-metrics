package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mxschmitt/golang-hetzner-robot-metrics/pkg/api"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	hetznerServersHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "hetzner_robot_servers_price",
		Help:    "Hetzner Robot Server",
		Buckets: prometheus.LinearBuckets(10, 5, 100),
	}, []string{"key"})
)

func init() {
	prometheus.MustRegister(hetznerServersHistogram)
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
				hetznerServersHistogram.WithLabelValues(strconv.Itoa(server.Key)).Observe(price)
			}
			// Sleep 30 minutes
			time.Sleep(30 * 60 * time.Second)
		}
	}()
	log.Printf("Listening on %s", *addr)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
