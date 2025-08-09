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
	require.EqualError(t, err, "no healthy servers available")
}

func TestGetNextServer_LoopsThroughServersCollection(t *testing.T) {
	s1 := &server.Server{ID: "1", IsHealthy: true}
	s2 := &server.Server{ID: "2", IsHealthy: true}
	s3 := &server.Server{ID: "3", IsHealthy: true}

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

func TestGetNextServer_ConcurrentRequests(t *testing.T) {
	s1 := &server.Server{ID: "1", IsHealthy: true}
	s2 := &server.Server{ID: "2", IsHealthy: true}
	s3 := &server.Server{ID: "3", IsHealthy: true}

	serverPool := &server.ServerPool{
		Servers: []*server.Server{s1, s2, s3},
	}

	lb := &RoundRobinLoadBalancer{
		serverPool:      serverPool,
		lastServerIndex: -1,
	}

	const numGoroutines = 99
	results := make(chan string, numGoroutines)

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			srv, err := lb.GetNextServer()
			require.NoError(t, err)
			results <- srv.ID
		}()
	}

	wg.Wait()
	close(results)

	counts := map[string]int{}
	for id := range results {
		counts[id]++
	}

	assert.Len(t, counts, 3, "all 3 servers should be used in round-robin balancer")
	assert.Equal(t, counts["1"], 33)
	assert.Equal(t, counts["2"], 33)
	assert.Equal(t, counts["3"], 33)
}
