package server

import (
	"errors"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type HealthCheckConfig struct {
	Retries  int
	Delay    time.Duration
	Timeout  time.Duration
	Interval time.Duration
}

var DefaultHealthCheckConfig = HealthCheckConfig{
	Retries:  3,
	Delay:    500 * time.Millisecond,
	Timeout:  2 * time.Second,
	Interval: 5 * time.Second,
}

var ErrNoHealthyServers = errors.New("no healthy servers available")

func GetHealthyServers(pool *ServerPool) []*Server {
	all := pool.GetAllServers()
	healthyServers := make([]*Server, 0, len(all))
	for _, srv := range all {
		srv.Mu.RLock()
		if srv.IsHealthy {
			healthyServers = append(healthyServers, srv)
		}
		srv.Mu.RUnlock()
	}
	return healthyServers
}

func StartHealthChecking(pool *ServerPool, cfg HealthCheckConfig, logger *zap.SugaredLogger) {
	go func() {
		for {
			for _, srv := range pool.GetAllServers() {
				checkServerHealth(srv, cfg, logger)
			}
			time.Sleep(cfg.Interval)
		}
	}()
}

func checkServerHealth(server *Server, cfg HealthCheckConfig, logger *zap.SugaredLogger) {
	client := http.Client{
		Timeout: cfg.Timeout,
	}

	success := false

	for i := 0; i < cfg.Retries; i++ {
		response, err := client.Get(server.Url + server.HealthcheckUrl)
		if err == nil && response.StatusCode == http.StatusOK {
			success = true
			break
		}

		logger.Warnw("server returned error response for healthcheck", "serverName", server.Name, "error", err)
		logger.Warnw("retrying healthcheck", "retry", i+1, "serverName", server.Name)
		if i < cfg.Retries-1 {
			time.Sleep(cfg.Delay)
		}
	}

	server.Mu.Lock()
	defer server.Mu.Unlock()

	server.LastHealthCheck = time.Now()

	if success {
		//logger.Infow("server is healthy", "serverName", server.Name)
		server.IsHealthy = true
	} else {
		logger.Infow("server is unhealthy", "serverName", server.Name)
		server.IsHealthy = false
	}
}
