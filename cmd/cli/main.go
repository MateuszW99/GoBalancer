package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/MateuszW99/GoBalancer/internal/config"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"github.com/MateuszW99/GoBalancer/internal/strategy"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	lbPort := flag.Int("lb-port", 3000, "Port for load balancer to listen on")
	adminPort := flag.Int("admin-port", 3001, "Port for admin api to listen on")
	serverConfig := flag.String("server-config", "servers.json", "Servers to which traffic will be distributed")
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

	errCh := make(chan error, 2)

	adminServer := configureAdminServer(*adminPort)
	go func() {
		sugar.Infow("starting admin server on port", "port", *adminPort)
		errCh <- adminServer.ListenAndServe()
	}()

	server.StartHealthChecking(pool, server.DefaultHealthCheckConfig, sugar)

	lbServer := configureLoadBalancerServer(*lbPort, loadBalancer)
	go func() {
		sugar.Infow("starting load balancer on port", "port", *lbPort)
		errCh <- lbServer.ListenAndServe()
	}()

	if err != nil {
		sugar.Fatalf("failed to start load balancer server: %v", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		sugar.Fatalw("load balancer failed", "port", *lbPort, "err", err)
	case <-sig:
		sugar.Info("shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = adminServer.Shutdown(ctx)
		_ = lbServer.Shutdown(ctx)
	}
}

func configureLoadBalancerServer(port int, loadBalancer *strategy.LoadBalancer) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", loadBalancer.Serve)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return srv
}

func configureAdminServer(port int) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/admin", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "hello, world") })

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return srv
}
