package strategy

import (
	"github.com/MateuszW99/GoBalancer/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestGetNextServer_NoServers(t *testing.T) {
	serverPool := &server.ServerPool{Servers: []*server.Server{}}
	lb := &RoundRobinLoadBalancer{
		mu:              sync.Mutex{},
		serverPool:      serverPool,
		lastServerIndex: -1,
	}

	srv, err := lb.GetNextServer()

	require.Nil(t, srv)
	require.EqualError(t, err, "no servers found")
}

func TestGetNextServer_LoopsThroughServersCollection(t *testing.T) {
	s1 := &server.Server{ID: "1", Name: "A", Host: "127.0.0.1", Port: 80, IsHealthy: true}
	s2 := &server.Server{ID: "2", Name: "B", Host: "127.0.0.2", Port: 80, IsHealthy: true}
	s3 := &server.Server{ID: "3", Name: "C", Host: "127.0.0.3", Port: 80, IsHealthy: true}

	serverPool := &server.ServerPool{
		Servers: []*server.Server{s1, s2, s3},
	}

	lb := &RoundRobinLoadBalancer{
		serverPool:      serverPool,
		lastServerIndex: -1,
		mu:              sync.Mutex{},
	}

	expectedServers := []*server.Server{s1, s2, s3, s1, s2}

	for i, expectedServer := range expectedServers {
		srv, err := lb.GetNextServer()
		require.NoError(t, err)
		assert.Equal(t, expectedServer.ID, srv.ID, "step %d", i)
	}
}
