package http

import "net/http"

// Routes returns the initialized router
func (s *Server) useRoutes() *http.ServeMux {
	r := http.NewServeMux()
	r.HandleFunc("POST /", s.shortenerHandler.PostHandler)
	r.HandleFunc("GET /{id}", s.shortenerHandler.GetByIDHandler)

	return r
}
