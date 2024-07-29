package http

import (
	"net/http"

	"github.com/Kenny201/go-yandex-shortener.git/internal/http/handler"
)

type Server struct {
	handler       handler.Handler
	serverAddress string
}

func NewServer(serverAddress string, handler handler.Handler) *Server {
	return &Server{
		handler:       handler,
		serverAddress: serverAddress,

	}
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.serverAddress, s.useRoutes())
}
