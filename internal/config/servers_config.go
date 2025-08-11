package config

import (
	"encoding/json"
	"fmt"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"time"
)

type LoadBalancerConfig struct {
	ServerPools []ServerPoolConfig `json:"serverPools" yaml:"serverPools"`
}

type ServerPoolConfig struct {
	Name     string         `json:"name" yaml:"name"`
	Strategy string         `json:"strategy" yaml:"strategy"`
	Servers  []ServerConfig `json:"servers" yaml:"servers"`
}

type ServerConfig struct {
	ID             string `json:"id" yaml:"id"`
	Name           string `json:"name" yaml:"name"`
	Protocol       string `json:"protocol" yaml:"protocol"`
	Host           string `json:"host" yaml:"host"`
	Port           int    `json:"port" yaml:"port"`
	HealthcheckUrl string `json:"healthcheckUrl" yaml:"healthcheckUrl"`
}

func LoadServersFromFile(path string) ([]*server.ServerPool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	ext := filepath.Ext(path)
	cfg := &LoadBalancerConfig{}

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config format: %s", ext)
	}

	var serverPools []*server.ServerPool
	for _, serverPoolConfig := range cfg.ServerPools {
		pool := server.NewServerPool(serverPoolConfig.Name, serverPoolConfig.Strategy)
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
