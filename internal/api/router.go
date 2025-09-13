package api

import (
	"github.com/go-chi/chi/v5"
)

func NewRouter(admin *AdminApi) *chi.Mux {
	router := chi.NewRouter()

	router.Route("/admin", func(r chi.Router) {
		r.Get("/serverPools", admin.GetAllServerPools)
		r.Get("/serverPools/{name}", admin.GetServerPool)
	})

	return router
}
