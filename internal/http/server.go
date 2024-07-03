package http

import "net/http"

type Server struct {
	urlHandler URLHandler
	listenPort string
}

func NewServer(listenPort string, handler URLHandler) *Server {
	return &Server{
		urlHandler: handler,
		listenPort: listenPort,
	}
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.listenPort, s.useRoutes())
}
