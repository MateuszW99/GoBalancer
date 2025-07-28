package server

import (
	"log"
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

func StartHealthChecking(pool *ServerPool, interval time.Duration) {
	go func() {
		for {
			for _, srv := range pool.GetAllServers() {
				checkServerHealth(srv)
			}
			time.Sleep(interval)
		}
	}()
}

func checkServerHealth(server *Server) {
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

		log.Printf("%s returned error response for healthcheck %v", server.Name, err)
		log.Printf("running %d retry", i+1)

		if i < retries-1 {
			time.Sleep(delay)
		}
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	server.LastHealthCheck = time.Now()

	if success {
		log.Printf("%s is healthy", server.Name)
		server.IsHealthy = true
	} else {
		log.Printf("%s is unhealthy", server.Name)
		server.IsHealthy = false
	}
}
