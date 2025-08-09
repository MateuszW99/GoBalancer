package config

import (
	"encoding/json"
	"fmt"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"os"
	"time"
)

type LoadBalancerConfig struct {
	ServerPools []ServerPoolConfig `json:"serverPools"`
}

type ServerPoolConfig struct {
	Name     string         `json:"name"`
	Strategy string         `json:"strategy"`
	Servers  []ServerConfig `json:"servers"`
}

type ServerConfig struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Protocol       string `json:"protocol"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	HealthcheckUrl string `json:"healthcheckUrl"`
}

func LoadServersFromFile(path string) ([]*server.ServerPool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg LoadBalancerConfig
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config JSON %w", err)
	}

	var serverPools []*server.ServerPool
	for _, serverPoolConfig := range cfg.ServerPools {
		pool := server.NewServerPool(serverPoolConfig.Name)
		for _, serverConfig := range serverPoolConfig.Servers {
			_ = pool.AddServer(serverConfig.newServerFromConfig(serverConfig))
		}
		serverPools = append(serverPools, pool)
	}

	return serverPools, nil
}

func (ServerConfig) newServerFromConfig(cfg ServerConfig) *server.Server {
	return &server.Server{
		ID:              cfg.ID,
		Name:            cfg.Name,
		Protocol:        cfg.Protocol,
		Host:            cfg.Host,
		Port:            cfg.Port,
		HealthcheckUrl:  cfg.HealthcheckUrl,
		Url:             fmt.Sprintf("%s://%s:%d", cfg.Protocol, cfg.Host, cfg.Port),
		IsHealthy:       true,
		LastHealthCheck: time.Now(),
	}
}
