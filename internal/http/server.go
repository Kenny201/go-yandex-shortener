package http

import (
	"net/http"
)

type Server struct {
	shortenerHandler ShortenerHandler
	serverAddress    string
}

func NewServer(serverAddress string, handler ShortenerHandler) *Server {
	return &Server{
		shortenerHandler: handler,
		serverAddress:    serverAddress,
	}
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.serverAddress, s.useRoutes())
}
