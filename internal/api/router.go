package api

import (
	"github.com/MateuszW99/GoBalancer/internal/strategy"
	"github.com/go-chi/chi/v5"
)

func NewAdminRouter(admin *AdminApi) *chi.Mux {
	router := chi.NewRouter()

	router.Route("/admin", func(r chi.Router) {
		r.Get("/serverPools", admin.GetAllServerPools)
		r.Get("/serverPools/{name}", admin.GetServerPool)
	})

	return router
}

func NewLoadBalancerRouter(lb *strategy.LoadBalancer) *chi.Mux {
	router := chi.NewRouter()
	router.HandleFunc("/*", lb.Serve)
	return router
}
