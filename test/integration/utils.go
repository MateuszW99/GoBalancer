package integration

import (
	"github.com/MateuszW99/GoBalancer/internal/server"
	"github.com/MateuszW99/GoBalancer/internal/strategy"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var e2eTestCfg = server.HealthCheckConfig{
	Retries:  1,
	Delay:    10 * time.Millisecond,
	Timeout:  50 * time.Millisecond,
	Interval: 20 * time.Millisecond,
}

func startLoadBalancer(t *testing.T, servers []*server.Server, strat string) *httptest.Server {
	t.Helper()

	pool := server.NewServerPool("testPool", strat)
	for _, s := range servers {
		err := pool.AddServer(s)
		if err != nil {
			return nil
		}
	}

	logger := zap.NewNop().Sugar()
	lb, err := strategy.SelectLoadBalancerWithStrategy(strategy.ParseStrategyType(strat), pool, logger)
	if err != nil {
		t.Fatalf("failed to select load balancer strategy: %v", err)
	}

	server.StartHealthChecking(pool, e2eTestCfg, logger)

	testServer := httptest.NewServer(http.HandlerFunc(lb.Serve))
	t.Cleanup(testServer.Close)
	return testServer
}

func waitUntilServerUnhealthy(t *testing.T, srv *server.Server, cfg server.HealthCheckConfig) {
	t.Helper()

	retryBackoff := time.Duration(0)
	if cfg.Retries > 1 {
		retryBackoff = time.Duration(cfg.Retries-1) * cfg.Delay
	}
	deadline := time.Now().Add(cfg.Timeout + retryBackoff + cfg.Interval + 100*time.Millisecond)

	pollEvery := cfg.Delay / 5
	if pollEvery < 5*time.Millisecond {
		pollEvery = 5 * time.Millisecond
	}
	if pollEvery > 50*time.Millisecond {
		pollEvery = 50 * time.Millisecond
	}

	ticker := time.NewTimer(pollEvery)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		srv.Mu.RLock()
		unhealthy := !srv.IsHealthy
		srv.Mu.RUnlock()
		if unhealthy {
			return
		}
		<-ticker.C
	}

	t.Fatalf("server %s was not marked unhealthy within %v", srv.Name, time.Until(deadline)*-1)
}
