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
}

func NewServerPool(name string) *ServerPool {
	return &ServerPool{
		Name:    name,
		Servers: make([]*Server, 0),
	}
}

func (serverPool *ServerPool) AddServer(server *Server) error {
	serverPool.Servers = append(serverPool.Servers, server)
	return nil
}

func (serverPool *ServerPool) GetAllServers() []*Server {
	return serverPool.Servers
}
