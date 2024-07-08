package http

import (
	"fmt"
	"net/http"
)

type Server struct {
	shortenerHandler ShortenerHandler
	listenPort       string
	listenDomain     string
}

func NewServer(listenDomain string, listenPort int, handler ShortenerHandler) *Server {
	return &Server{
		shortenerHandler: handler,
		listenPort:       listenPort,
		listenDomain:     listenDomain,
	}
}

func (s *Server) Start() error {
	listenHost := fmt.Sprintf("%s:%v", s.listenDomain, s.listenPort)

	return http.ListenAndServe(listenHost, s.useRoutes())
}
