package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Kenny201/go-yandex-shortener.git/internal/http/handler"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http/middleware"
)

func useRoutes(handler handler.Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(
		middleware.Gzip,
		middleware.Logger,
	)
	userID := uuid.New().String()

	r.With(middleware.NewTokenMiddleware(userID)).Post("/", handler.Post)
	r.Get("/{id}", handler.Get)
	r.Get("/ping", handler.Ping)

	r.With(middleware.NewTokenMiddleware(userID)).Route("/api", func(r chi.Router) {
		r.Post("/shorten", handler.PostAPI)
		r.Post("/shorten/batch", handler.PostBatch)
		r.With(middleware.AuthCheckMiddleware()).Get("/user/urls", handler.GetAll)
	})

	return r
}
