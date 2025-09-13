package api

import (
	"time"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type Server struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Url             string    `json:"url"`
	IsHealthy       bool      `json:"isHealthy"`
	LastHealthCheck time.Time `json:"lastHealthCheck"`
	HealthcheckUrl  string    `json:"healthcheckUrl"`
}

type ServerPool struct {
	Name     string   `json:"name"`
	Strategy string   `json:"strategy"`
	Servers  []Server `json:"servers"`
}
