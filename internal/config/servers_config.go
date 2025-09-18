package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"time"
)

type AppConfig struct {
	ServerPools []*server.ServerPool
	LbPort      int
	AdminPort   int
}

type LoadBalancerConfig struct {
	ServerPools []serverPoolConfig `json:"serverPools" yaml:"serverPools"`
}

type serverPoolConfig struct {
	Name     string         `json:"name" yaml:"name"`
	Strategy string         `json:"strategy" yaml:"strategy"`
	Servers  []serverConfig `json:"servers" yaml:"servers"`
}

type serverConfig struct {
	ID             string `json:"id" yaml:"id"`
	Name           string `json:"name" yaml:"name"`
	Protocol       string `json:"protocol" yaml:"protocol"`
	Host           string `json:"host" yaml:"host"`
	Port           int    `json:"port" yaml:"port"`
	HealthcheckUrl string `json:"healthcheckUrl" yaml:"healthcheckUrl"`
}

func Load(logger *zap.SugaredLogger) (*AppConfig, error) {
	lbPort, adminPort, serverConfig := readConfigFlags()

	lbConfig, err := loadLBConfig(serverConfig, logger)
	if err != nil {
		return nil, err
	}
	if len(lbConfig.ServerPools) == 0 {
		return nil, fmt.Errorf("no servers found in %v", serverConfig)
	}

	serverPools := buildServerPools(lbConfig)

	return &AppConfig{
		ServerPools: serverPools,
		LbPort:      lbPort,
		AdminPort:   adminPort,
	}, nil
}

func loadLBConfig(path string, logger *zap.SugaredLogger) (*LoadBalancerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	ext := filepath.Ext(path)
	cfg := &LoadBalancerConfig{}

	logger.Infof("reading server configuration from %s", path)
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

	logger.Infof("found %d server configs", len(cfg.ServerPools))

	return cfg, nil
}

func readConfigFlags() (lbPort int, adminPort int, serverConfigPath string) {
	lbPortFlag := flag.Int("lb-port", 3000, "Port for load balancer to listen on")
	adminPortFlag := flag.Int("admin-port", 3001, "Port for admin api to listen on")
	serverConfigPathFlag := flag.String("server-config", "servers.json", "Servers to which traffic will be distributed")
	flag.Parse()
	return *lbPortFlag, *adminPortFlag, *serverConfigPathFlag
}

func buildServerPools(cfg *LoadBalancerConfig) []*server.ServerPool {
	var serverPools []*server.ServerPool
	for _, serverPoolConfig := range cfg.ServerPools {
		pool := server.NewServerPool(serverPoolConfig.Name, serverPoolConfig.Strategy)
		for _, serverConfig := range serverPoolConfig.Servers {
			_ = pool.AddServer(serverConfig.newServerFromConfig(serverConfig))
		}
		serverPools = append(serverPools, pool)
	}
	return serverPools
}

func (serverConfig) newServerFromConfig(cfg serverConfig) *server.Server {
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
