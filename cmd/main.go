package main

import (
	"fmt"
	"github.com/MateuszW99/GoBalancer/internal/config"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"github.com/MateuszW99/GoBalancer/internal/strategy"
	"log"
	"net/http"
)

func main() {
	servers, err := config.LoadServersFromFile("servers.json")
	if err != nil {
		log.Fatalf("failed to load server config: %v", err)
	}

	pool := server.NewServerPool()
	for i := range servers {
		pool.AddServer(&servers[i])
	}

	roundRobin := strategy.NewRoundRobinLoadBalancer(pool)
	loadBalancer := strategy.NewLoadBalancer(roundRobin)
	distributeLoad(3000, loadBalancer, pool)

	select {}
}

func distributeLoad(port int, loadBalancer strategy.LoadBalancer, pool *server.ServerPool) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", loadBalancer.Serve)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Printf("starting load balancer on port %d", port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("load balancer on port %d failed: %v", port, err)
	}
}
