package http

import "net/http"

// Routes returns the initialized router
func (s *Server) useRoutes() *http.ServeMux {
	r := http.NewServeMux()
	r.HandleFunc("POST /", s.urlHandler.PostHandler)
	r.HandleFunc("GET /{id}", s.urlHandler.GetByIDHandler)

	return r
}
