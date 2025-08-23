package strategy

import (
	"github.com/MateuszW99/GoBalancer/internal/server"
	"math"
	"sync"
)

type LeastConnectionLoadBalancer struct {
	serverPool        *server.ServerPool
	activeConnections map[*server.Server]int // server : connectionCount
	mu                sync.Mutex
}

func NewLeastConnectionLoadBalancer(serverPool *server.ServerPool) *LeastConnectionLoadBalancer {
	return &LeastConnectionLoadBalancer{
		serverPool:        serverPool,
		activeConnections: make(map[*server.Server]int),
	}
}

func (lb *LeastConnectionLoadBalancer) GetNextServer() (*server.Server, error) {
	healthy := server.GetHealthyServers(lb.serverPool)
	if len(healthy) == 0 {
		return nil, server.ErrNoHealthyServers
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	var selected *server.Server
	leastConnections := math.MaxInt

	for _, s := range healthy {
		conn := lb.activeConnections[s]
		if conn < leastConnections {
			leastConnections = conn
			selected = s
		}
	}

	lb.activeConnections[selected] = leastConnections + 1
	return selected, nil
}

func (lb *LeastConnectionLoadBalancer) Done(srv *server.Server) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	if conn := lb.activeConnections[srv]; conn > 0 {
		lb.activeConnections[srv] = conn - 1
	}
}
