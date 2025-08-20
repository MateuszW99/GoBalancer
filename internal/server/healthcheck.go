package server

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

const (
	retries = 3
	delay   = 500 * time.Millisecond
	timeout = 2 * time.Second
)

func GetHealthyServers(pool *ServerPool) []*Server {
	all := pool.GetAllServers()
	healthyServers := make([]*Server, 0, len(all))
	for _, server := range all {
		server.mu.RLock()
		if server.IsHealthy {
			healthyServers = append(healthyServers, server)
		}
		server.mu.RUnlock()
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

		logger.Warn("server returned error response for healthcheck", zap.String("serverName", server.Name), zap.Error(err))
		logger.Warn("retrying healthcheck", zap.Int("retry", i+1), zap.String("serverName", server.Name))

		if i < retries-1 {
			time.Sleep(delay)
		}
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	server.LastHealthCheck = time.Now()

	if success {
		logger.Info("server is healthy", zap.String("serverName", server.Name))
		server.IsHealthy = true
	} else {
		logger.Info("server is unhealthy", zap.String("serverName", server.Name))
		server.IsHealthy = false
	}
}
