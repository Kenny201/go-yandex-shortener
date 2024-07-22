package http

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

// Routes returns the initialized router
func (s *Server) useRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(
		middleware.Json,
		middleware.Gzip,
		middleware.Logger,
	)

	r.Post("/", s.handler.Post)
	r.Get("/{id}", s.handler.Get)

	r.Route("/api", func(r chi.Router) {
		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", s.handler.PostAPI)
		})
	})

	return r
}
