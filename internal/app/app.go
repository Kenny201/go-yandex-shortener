package app

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/handlers"
	"net/http"
)

type Server struct {
	listenPort string
}

func NewServer(listenPort string) *Server {
	return &Server{
		listenPort: listenPort,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.Handler)
	mux.HandleFunc("/{id}", handlers.Handler)

	return http.ListenAndServe(s.listenPort, mux)
}
