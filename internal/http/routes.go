package http

import (
	"github.com/go-chi/chi/v5"
)

// Routes returns the initialized router
func (s *Server) useRoutes() *chi.Mux {
	r := chi.NewRouter()
	r.HandleFunc("POST /", s.shortenerHandler.PostHandler)
	r.HandleFunc("GET /{id}", s.shortenerHandler.GetByIDHandler)

	return r
}
