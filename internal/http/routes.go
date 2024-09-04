package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/Kenny201/go-yandex-shortener.git/internal/http/handler"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http/middleware"
)

func useRoutes(handler handler.Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(
		middleware.Gzip,
		middleware.Logger,
	)

	r.With(middleware.AuthMiddleware()).Post("/", handler.Post)
	r.Get("/{id}", handler.Get)
	r.Get("/ping", handler.Ping)

	r.Route("/api", func(r chi.Router) {
		r.With(middleware.AuthMiddleware()).Route("/shorten", func(r chi.Router) {
			r.Post("/", handler.PostAPI)
			r.Post("/batch", handler.PostBatch)
		})
		r.With(middleware.AuthCheckMiddleware()).Route("/user", func(r chi.Router) {
			r.Get("/urls", handler.GetAll)
			r.Delete("/urls", handler.Delete)
		})
	})

	return r
}
