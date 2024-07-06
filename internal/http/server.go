package http

import (
	"fmt"
	"net/http"
)

type Server struct {
	listenDomain string
	listenPort   int
	urlHandler   URLHandler
}

func NewServer(listenDomain string, listenPort int, handler URLHandler) *Server {
	return &Server{
		listenDomain: listenDomain,
		listenPort:   listenPort,
		urlHandler:   handler,
	}
}

func (s *Server) Start() error {
	listenHost := fmt.Sprintf("%s:%v", s.listenDomain, s.listenPort)

	return http.ListenAndServe(listenHost, s.useRoutes())
}
