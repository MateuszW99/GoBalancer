package strategy

import (
	"github.com/MateuszW99/GoBalancer/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestGetNextServer_NoServer(t *testing.T) {
	testCases := []struct {
		name    string
		servers []*server.Server
	}{
		{
			name:    "no servers",
			servers: []*server.Server{},
		},
		{
			name:    "no healthy servers",
			servers: []*server.Server{{ID: "1", IsHealthy: false}},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serverPool := &server.ServerPool{Servers: tc.servers}
			lb := &LeastConnectionLoadBalancer{
				mu:                sync.Mutex{},
				serverPool:        serverPool,
				activeConnections: make(map[*server.Server]int),
			}

			srv, err := lb.GetNextServer()

			require.Nil(t, srv)
			require.EqualError(t, err, "no healthy servers available")
		})
	}
}

func TestGetNextServer_ReturnsNextServerAndIncreasesActiveConnectionsCounter(t *testing.T) {
	s1 := &server.Server{ID: "1", IsHealthy: true}
	s2 := &server.Server{ID: "2", IsHealthy: true}
	s3 := &server.Server{ID: "3", IsHealthy: true}

	serverPool := &server.ServerPool{
		Servers: []*server.Server{s1, s2, s3},
	}

	lb := NewLeastConnectionLoadBalancer(serverPool)
	lb.activeConnections[s1] = 10
	lb.activeConnections[s2] = 100
	lb.activeConnections[s3] = 1000

	candidate, err := lb.GetNextServer()

	require.NoError(t, err)
	assert.NotNil(t, candidate)
	assert.Equal(t, candidate, s1)
	assert.Equal(t, 11, lb.activeConnections[s1])
	assert.Equal(t, 100, lb.activeConnections[s2])
	assert.Equal(t, 1000, lb.activeConnections[s3])
}

func TestLeastConnectionLoadBalancer_ConcurrentRequests(t *testing.T) {
	s1 := &server.Server{ID: "1", IsHealthy: true}
	s2 := &server.Server{ID: "2", IsHealthy: false}
	s3 := &server.Server{ID: "3", IsHealthy: true}

	serverPool := &server.ServerPool{
		Servers: []*server.Server{s1, s2, s3},
	}

	lb := NewLeastConnectionLoadBalancer(serverPool)

	const concurrentRequests = 100
	results := make(chan string, concurrentRequests)
	errs := make(chan error, concurrentRequests)
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		go func() {
			defer wg.Done()
			srv, err := lb.GetNextServer()
			errs <- err
			if err == nil {
				results <- srv.ID
			}
		}()
	}

	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}

	counts := map[string]int{}
	for id := range results {
		counts[id]++
	}

	assert.Len(t, counts, 2, "only healthy server should be used in least connections load balancer")
	assert.Equal(t, 50, counts[s1.ID])
	assert.Equal(t, 50, counts[s3.ID])
}

func TestDone_ReducesActiveConnections(t *testing.T) {
	s1 := &server.Server{ID: "1", IsHealthy: true}
	s2 := &server.Server{ID: "2", IsHealthy: true}

	serverPool := &server.ServerPool{
		Servers: []*server.Server{s1, s2},
	}

	lb := NewLeastConnectionLoadBalancer(serverPool)

	// simulate traffic by hardcoding active connections
	lb.activeConnections[s1] = 11
	lb.activeConnections[s2] = 9

	lb.Done(s1)
	assert.Equal(t, 10, lb.activeConnections[s1])
	assert.Equal(t, 9, lb.activeConnections[s2])
}
