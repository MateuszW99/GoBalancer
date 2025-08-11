package main

import (
	"flag"
	"fmt"
	"github.com/MateuszW99/GoBalancer/internal/config"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"github.com/MateuszW99/GoBalancer/internal/strategy"
	"log"
	"net/http"
	"time"
)

func main() {
	port := flag.Int("port", 3000, "Port to listen on")
	serverConfig := flag.String("server-config", "servers.yaml", "Servers to which traffic will be distributed")
	flag.Parse()

	serverPools, err := config.LoadServersFromFile(*serverConfig)
	if err != nil {
		log.Fatalf("failed to load server config: %v", err)
	}
	if len(serverPools) == 0 {
		log.Fatalf("no servers found in %v", *serverConfig)
	}

	pool := serverPools[0] // TODO: run all pools concurrently
	loadBalancer, err := strategy.SelectLoadBalancerWithStrategy(strategy.ParseStrategyType(pool.Strategy), pool)
	if err != nil {
		log.Fatalf("failed to select strategy: %v", err)
	}
	server.StartHealthChecking(pool, 5*time.Second)
	distributeLoad(*port, loadBalancer)

	select {}
}

func distributeLoad(port int, loadBalancer *strategy.LoadBalancer) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", loadBalancer.Serve)

	trafficDistributor := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Printf("starting load balancer on port %d", port)

	if err := trafficDistributor.ListenAndServe(); err != nil {
		log.Fatalf("load balancer on port %d failed: %v", port, err)
	}
}
