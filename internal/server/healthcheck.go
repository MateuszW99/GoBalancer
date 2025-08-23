package server

import (
	"errors"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const (
	retries = 3
	delay   = 500 * time.Millisecond
	timeout = 2 * time.Second
)

var ErrNoHealthyServers = errors.New("no healthy servers available")

func GetHealthyServers(pool *ServerPool) []*Server {
	all := pool.GetAllServers()
	healthyServers := make([]*Server, 0, len(all))
	for _, srv := range all {
		srv.mu.RLock()
		if srv.IsHealthy {
			healthyServers = append(healthyServers, srv)
		}
		srv.mu.RUnlock()
	}
	return healthyServers
}

func StartHealthChecking(pool *ServerPool, interval time.Duration, logger *zap.SugaredLogger) {
	go func() {
		for {
			for _, srv := range pool.GetAllServers() {
				checkServerHealth(srv, logger)
			}
			time.Sleep(interval)
		}
	}()
}

func checkServerHealth(server *Server, logger *zap.SugaredLogger) {
	client := http.Client{
		Timeout: timeout,
	}

	success := false

	for i := 0; i < retries; i++ {
		response, err := client.Get(server.Url + server.HealthcheckUrl)
		if err == nil && response.StatusCode == http.StatusOK {
			success = true
			break
		}

		logger.Warnw("server returned error response for healthcheck", "serverName", server.Name, "error", err)
		logger.Warnw("retrying healthcheck", "retry", i+1, "serverName", server.Name)
		if i < retries-1 {
			time.Sleep(delay)
		}
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	server.LastHealthCheck = time.Now()

	if success {
		logger.Infow("server is healthy", "serverName", server.Name)
		server.IsHealthy = true
	} else {
		logger.Infow("server is unhealthy", "serverName", server.Name)
		server.IsHealthy = false
	}
}
