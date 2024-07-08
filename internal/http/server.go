package http

import "net/http"

type Server struct {
	shortenerHandler ShortenerHandler
	listenPort       string
}

func NewServer(listenPort string, handler ShortenerHandler) *Server {
	return &Server{
		shortenerHandler: handler,
		listenPort:       listenPort,
	}
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.listenPort, s.useRoutes())
}
