package main

import (
	"fmt"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"github.com/MateuszW99/GoBalancer/internal/strategy"
	"log"
	"net/http"
	"time"
)

var servers = []server.Server{
	{
		ID:              "1",
		Name:            "Server 2137",
		Protocol:        "http",
		Host:            "localhost",
		Port:            2137,
		Url:             "http://localhost:2137",
		IsHealthy:       true,
		LastHealthCheck: time.Now(),
	},
	{
		ID:              "2",
		Name:            "Server 2138",
		Protocol:        "http",
		Host:            "localhost",
		Port:            2138,
		Url:             "http://localhost:2138",
		IsHealthy:       true,
		LastHealthCheck: time.Now(),
	},
}

func main() {
	pool := server.NewServerPool()
	for _, srv := range servers {
		pool.AddServer(&srv)
	}

	roundRobin := strategy.NewRoundRobinLoadBalancer(pool)
	loadBalancer := strategy.NewLoadBalancer(roundRobin)
	distributeLoad(3000, loadBalancer, pool)

	select {}
}

func distributeLoad(port int, loadBalancer strategy.LoadBalancer, pool *server.ServerPool) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", loadBalancer.Serve)
	//mux.HandleFunc("/loadbalancer/get-all-servers")

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Printf("starting load balancer on port %d", port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("load balancer on por t%d failed: %v", port, err)
	}
}
