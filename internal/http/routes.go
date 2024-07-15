package http

import (
	"github.com/go-chi/chi/v5"
)

// Routes returns the initialized router
func (s *Server) useRoutes() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", s.shortenerHandler.PostHandler)
	r.Get("/{id}", s.shortenerHandler.GetByIDHandler)

	return r
}
