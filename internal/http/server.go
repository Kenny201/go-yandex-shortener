package http

import "net/http"

type Server struct {
	urlHandler UrlHandler
	listenPort string
}

func NewServer(listenPort string, handler UrlHandler) *Server {
	return &Server{
		urlHandler: handler,
		listenPort: listenPort,
	}
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.listenPort, s.useRoutes())
}
