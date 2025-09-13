package api

import (
	"encoding/json"
	"fmt"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAllServerPools_ReturnsAllConfiguredServerPools(t *testing.T) {
	pool := &server.ServerPool{
		Name:     "serverPool",
		Strategy: "rb",
		Servers: []*server.Server{
			{ID: "1", Name: "srv1", Url: "http://localhost:8080", IsHealthy: true, HealthcheckUrl: "/health"},
		},
	}

	api := NewAdminApi([]*server.ServerPool{pool}, zap.NewNop().Sugar())
	r := NewRouter(api)

	request := httptest.NewRequest(http.MethodGet, "/admin/serverPools", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, request)

	require.Equal(t, http.StatusOK, rec.Code)

	var response []ServerPool
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
	assert.Equal(t, 1, len(response))
	assert.Equal(t, "serverPool", response[0].Name)
	assert.Equal(t, "rb", response[0].Strategy)
	assert.Equal(t, 1, len(response[0].Servers))
	assert.Equal(t, "1", response[0].Servers[0].ID)
	assert.Equal(t, "srv1", response[0].Servers[0].Name)
	assert.True(t, response[0].Servers[0].IsHealthy)
	assert.Equal(t, "/health", response[0].Servers[0].HealthcheckUrl)
}

func TestGetServerPool_WhenServerPoolExists_ReturnsOKWithResponse(t *testing.T) {
	poolName := "pool"
	pool := &server.ServerPool{
		Name:     poolName,
		Strategy: "rb",
		Servers: []*server.Server{
			{ID: "1", Name: "srv1", Url: "http://localhost:8080", IsHealthy: true, HealthcheckUrl: "/health"},
		},
	}

	api := NewAdminApi([]*server.ServerPool{pool}, zap.NewNop().Sugar())
	r := NewRouter(api)

	request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/serverPools/%s", poolName), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, request)

	require.Equal(t, http.StatusOK, rec.Code)

	var responseBody ServerPool
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&responseBody))
	assert.Equal(t, "pool", responseBody.Name)
	assert.Equal(t, "rb", responseBody.Strategy)
	require.Len(t, responseBody.Servers, 1)
	assert.Equal(t, "1", responseBody.Servers[0].ID)
	assert.Equal(t, "srv1", responseBody.Servers[0].Name)
	assert.True(t, responseBody.Servers[0].IsHealthy)
	assert.Equal(t, "/health", responseBody.Servers[0].HealthcheckUrl)
}

func TestGetServerPool_WhenServerPoolExists_ReturnsNotFound(t *testing.T) {
	pool := &server.ServerPool{
		Name: "poolA",
	}

	api := NewAdminApi([]*server.ServerPool{pool}, zap.NewNop().Sugar())
	r := NewRouter(api)

	request := httptest.NewRequest(http.MethodGet, "/admin/serverPools/poolB", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, request)

	require.Equal(t, http.StatusNotFound, rec.Code)

	var errResponse ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&errResponse))
	require.Equal(t, "server pool not found", errResponse.Message)
}
