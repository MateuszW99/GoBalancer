package strategy

import (
	"errors"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"sync"
)

type RoundRobinLoadBalancer struct {
	serverPool      *server.ServerPool
	mu              sync.Mutex
	lastServerIndex int
}

func NewRoundRobinLoadBalancer(serverPool *server.ServerPool) *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{
		serverPool:      serverPool,
		lastServerIndex: -1,
	}
}

func (lb *RoundRobinLoadBalancer) GetNextServer() (*server.Server, error) {
	servers := lb.serverPool.GetAllServers()

	if len(servers) == 0 {
		return nil, errors.New("no servers found")
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.lastServerIndex = (lb.lastServerIndex + 1) % len(servers)

	selectedServer := servers[lb.lastServerIndex]
	return selectedServer, nil
}
