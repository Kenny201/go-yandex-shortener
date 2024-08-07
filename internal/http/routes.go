package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/Kenny201/go-yandex-shortener.git/internal/http/middleware"
)

func (s *Server) useRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(
		middleware.Gzip,
		middleware.Logger,
	)

	r.Post("/", s.handler.Post)
	r.Get("/{id}", s.handler.Get)

	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", s.handler.PostAPI)
	})

	return r
}
