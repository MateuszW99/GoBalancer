package server

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetHealthyServers_ReturnsOnlyHealthyServers(t *testing.T) {
	s1 := &Server{ID: "1", Name: "S1", IsHealthy: true}
	s2 := &Server{ID: "2", Name: "S2", IsHealthy: true}
	s3 := &Server{ID: "3", Name: "S3", IsHealthy: false}

	pool := &ServerPool{
		Servers: []*Server{s1, s2},
	}

	candidate := GetHealthyServers(pool)

	assert.Len(t, candidate, 2)
	assert.Contains(t, candidate, s1)
	assert.Contains(t, candidate, s2)
	assert.NotContains(t, candidate, s3)
}

func TestGetHealthyServers_WhenAllServersUnhealthy_ReturnsEmptyArray(t *testing.T) {
	s1 := &Server{ID: "1", Name: "S1", IsHealthy: false}
	s2 := &Server{ID: "2", Name: "S2", IsHealthy: false}

	pool := &ServerPool{
		Servers: []*Server{s1, s2},
	}

	candidate := GetHealthyServers(pool)

	assert.Len(t, candidate, 0)
}

func TestStartHealthChecking_MarksServerHealthy(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	lbServer := &Server{
		Url:            mockServer.URL,
		HealthcheckUrl: "",
		IsHealthy:      false,
	}

	pool := &ServerPool{
		Servers: []*Server{lbServer},
	}

	StartHealthChecking(pool, 100*time.Millisecond)
	time.Sleep(300 * time.Millisecond)

	lbServer.mu.RLock()
	defer lbServer.mu.RUnlock()
	assert.True(t, lbServer.IsHealthy, "server should be marked healthy")
}

func TestStartHealthChecking_MarksServerUnhealthyAfterRetries(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer mockServer.Close()

	lbServer := &Server{
		Url:            mockServer.URL,
		HealthcheckUrl: "",
		IsHealthy:      false,
	}

	pool := &ServerPool{
		Servers: []*Server{lbServer},
	}

	StartHealthChecking(pool, 100*time.Millisecond)
	time.Sleep((retries + 1) * time.Millisecond)

	lbServer.mu.RLock()
	defer lbServer.mu.RUnlock()
	assert.False(t, lbServer.IsHealthy, "server should be marked unhealthy after %d retries", retries)
}
