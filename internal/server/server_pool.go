package server

import "time"

type Server struct {
	ID              string
	Name            string
	Protocol        string
	Host            string
	Port            int
	Url             string
	IsHealthy       bool
	LastHealthCheck time.Time
}

type ServerPool struct {
	Servers []*Server
}

func NewServerPool() *ServerPool {
	return &ServerPool{
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
