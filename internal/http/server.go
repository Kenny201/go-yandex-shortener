package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Kenny201/go-yandex-shortener.git/internal/http/handler"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/closer"
)

const (
	shutdownTimeout = 5 * time.Second
)

type Server struct {
	server *http.Server
	ctx    context.Context
}

func NewServer(ctx context.Context, serverAddress string, handler handler.Handler) *Server {
	server := &http.Server{
		Addr:         serverAddress,
		Handler:      useRoutes(handler),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return &Server{server, ctx}
}

func (s *Server) Start() {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Could not listen on ", slog.String("address", s.server.Addr), slog.String("error", err.Error()))
		}
	}()

	closer.CL.Add(func(ctx context.Context) error {
		return s.server.Shutdown(s.ctx)
	})

	closer.CL.Add(func(ctx context.Context) error {
		time.Sleep(6 * time.Second)
		return nil
	})

	slog.Info("Listening on", slog.String("address", s.server.Addr))

	<-s.ctx.Done()

	slog.Info("Shutting down server gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)

	defer cancel()

	if err := closer.CL.Close(shutdownCtx); err != nil {
		slog.Error("Could not close", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
