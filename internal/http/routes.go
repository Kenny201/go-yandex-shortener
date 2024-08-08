package http

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/http/handler"
	"github.com/go-chi/chi/v5"

	"github.com/Kenny201/go-yandex-shortener.git/internal/http/middleware"
)

func useRoutes(handler handler.Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(
		middleware.Gzip,
		middleware.Logger,
	)

	r.Post("/", handler.Post)
	r.Get("/{id}", handler.Get)
	r.Get("/ping", handler.Ping)

	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", handler.PostAPI)
	})

	return r
}
