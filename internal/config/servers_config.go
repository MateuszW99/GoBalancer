package config

import (
	"encoding/json"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"os"
)

func LoadServersFromFile(path string) ([]server.Server, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var servers []server.Server
	if err := json.NewDecoder(file).Decode(&servers); err != nil {
		return nil, err
	}

	return servers, nil
}
