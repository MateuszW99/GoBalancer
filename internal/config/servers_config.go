package config

import (
	"encoding/json"
	"fmt"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"os"
	"time"
)

type ServerConfig struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Protocol       string `json:"protocol"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	HealthcheckUrl string `json:"healthcheckUrl"`
}

func LoadServersFromFile(path string) ([]*server.Server, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var serverConfigs []ServerConfig
	if err := json.NewDecoder(file).Decode(&serverConfigs); err != nil {
		return nil, err
	}

	servers := make([]*server.Server, 0, len(serverConfigs))
	for _, cfg := range serverConfigs {
		servers = append(servers, newServerFromConfig(cfg))
	}

	return servers, nil
}

func newServerFromConfig(cfg ServerConfig) *server.Server {
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
