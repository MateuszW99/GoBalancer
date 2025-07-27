package main

import (
	"flag"
	"fmt"
	"github.com/MateuszW99/GoBalancer/internal/config"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"github.com/MateuszW99/GoBalancer/internal/strategy"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("port", 3000, "Port to listen on")
	serverConfig := flag.String("server-config", "servers.json", "Servers to which traffic will be distributed")
	flag.Parse()

	servers, err := config.LoadServersFromFile(*serverConfig)
	if err != nil {
		log.Fatalf("failed to load server config: %v", err)
	}
	if len(servers) == 0 {
		log.Fatalf("no servers found in %v", *serverConfig)
	}

	pool := server.NewServerPool()
	for i := range servers {
		pool.AddServer(&servers[i])
	}

	roundRobin := strategy.NewRoundRobinLoadBalancer(pool)
	loadBalancer := strategy.NewLoadBalancer(roundRobin)
	distributeLoad(*port, loadBalancer, pool)

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
