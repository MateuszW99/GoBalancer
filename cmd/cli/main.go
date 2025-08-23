package main

import (
	"flag"
	"fmt"
	"github.com/MateuszW99/GoBalancer/internal/config"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"github.com/MateuszW99/GoBalancer/internal/strategy"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()
	sugar := logger.Sugar()

	port := flag.Int("port", 3000, "Port to listen on")
	serverConfig := flag.String("server-config", "servers.yaml", "Servers to which traffic will be distributed")
	flag.Parse()

	serverPools, err := config.LoadServersFromFile(*serverConfig, sugar)
	if err != nil {
		sugar.Fatalf("failed to load server config: %v", err)
	}
	if len(serverPools) == 0 {
		sugar.Fatalf("no servers found in %v", *serverConfig)
	}

	pool := serverPools[0] // TODO: run all pools concurrently
	loadBalancer, err := strategy.SelectLoadBalancerWithStrategy(strategy.ParseStrategyType(pool.Strategy), pool, sugar)
	if err != nil {
		sugar.Fatalw("failed to select strategy", "error", err)
	}
	server.StartHealthChecking(pool, 5*time.Second, sugar)
	distributeLoad(*port, loadBalancer, sugar)

	select {}
}

func distributeLoad(port int, loadBalancer *strategy.LoadBalancer, logger *zap.SugaredLogger) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", loadBalancer.Serve)

	trafficDistributor := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	logger.Info("starting load balancer on port", "port", port)

	if err := trafficDistributor.ListenAndServe(); err != nil {
		logger.Fatalw("load balancer failed", "port", port, "err", err)
	}
}
