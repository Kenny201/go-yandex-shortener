package http

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/http/middlewares"
	"github.com/go-chi/chi/v5"
)

// Routes returns the initialized router
func (s *Server) useRoutes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middlewares.Logger)

	r.Post("/", s.shortenerHandler.PostHandler)
	r.Get("/{id}", s.shortenerHandler.GetByIDHandler)

	return r
}
