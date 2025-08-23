package server

import (
	"sync"
	"time"
)

type Server struct {
	ID              string
	Name            string
	Protocol        string
	Host            string
	Port            int
	Url             string
	IsHealthy       bool
	LastHealthCheck time.Time
	HealthcheckUrl  string
	mu              sync.RWMutex
}

type ServerPool struct {
	Name     string
	Strategy string
	Servers  []*Server
	mu       sync.RWMutex
}

func NewServerPool(name string, strategy string) *ServerPool {
	return &ServerPool{
		Name:     name,
		Strategy: strategy,
		Servers:  make([]*Server, 0),
	}
}

func (serverPool *ServerPool) AddServer(server *Server) error {
	serverPool.mu.Lock()
	serverPool.Servers = append(serverPool.Servers, server)
	serverPool.mu.Unlock()
	return nil
}

func (serverPool *ServerPool) GetAllServers() []*Server {
	serverPool.mu.RLock()
	servers := make([]*Server, len(serverPool.Servers))
	copy(servers, serverPool.Servers)
	serverPool.mu.RUnlock()
	return servers
}
