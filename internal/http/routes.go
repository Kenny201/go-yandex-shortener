package http

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/http/handler"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http/middleware"
	"github.com/go-chi/chi/v5"
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
		r.With(middleware.AuthMiddleware()).Post("/shorten", handler.PostAPI)
		r.With(middleware.AuthMiddleware()).Post("/shorten/batch", handler.PostBatch)
		r.With(middleware.AuthCheckMiddleware()).Get("/user/urls", handler.GetAll)
		r.With(middleware.AuthCheckMiddleware()).Delete("/user/urls", handler.Delete)
	})

	return r
}
