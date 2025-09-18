package api

import (
	"encoding/json"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

type AdminApi struct {
	serverPool []*server.ServerPool
	logger     *zap.SugaredLogger
}

func NewAdminApi(serverPool []*server.ServerPool, logger *zap.SugaredLogger) *AdminApi {
	return &AdminApi{
		serverPool: serverPool,
		logger:     logger,
	}
}

func (admin *AdminApi) GetAllServerPools(w http.ResponseWriter, r *http.Request) {
	serverPools := make([]ServerPool, len(admin.serverPool))
	for i, srvPool := range admin.serverPool {
		serverPools[i] = mapServerPool(srvPool)
	}

	if err := json.NewEncoder(w).Encode(serverPools); err != nil {
		admin.logger.Errorw("failed to encode server pool")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (admin *AdminApi) GetServerPool(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var found *server.ServerPool
	for _, srvPool := range admin.serverPool {
		if srvPool.Name == name {
			found = srvPool
			break
		}
	}

	if found == nil {
		writeErrorResponse(w, http.StatusNotFound, "server pool not found")
		return
	}

	resp := mapServerPool(found)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		admin.logger.Errorw("failed to encode server pool")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func mapServerPool(srvPool *server.ServerPool) ServerPool {
	servers := make([]Server, len(srvPool.Servers))
	for i, srv := range srvPool.Servers {
		servers[i] = Server{
			ID:              srv.ID,
			Name:            srv.Name,
			Url:             srv.Url,
			IsHealthy:       srv.IsHealthy,
			LastHealthCheck: srv.LastHealthCheck,
			HealthcheckUrl:  srv.HealthcheckUrl,
		}
	}

	return ServerPool{
		Name:     srvPool.Name,
		Strategy: srvPool.Strategy,
		Servers:  servers,
	}
}
