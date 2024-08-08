package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Kenny201/go-yandex-shortener.git/cmd/shortener/config"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http/handler"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/closer"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

func main() {
	var repository shortener.Repository

	conf, err := config.LoadConfig("./")

	if err != nil {
		slog.Error("error read config %v", slog.String("error", err.Error()))
		os.Exit(1)
	}

	args := config.NewArgs(conf)
	args.ParseFlags()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cl := closer.New()

	repository, err = storage.NewDatabaseShortenerRepository(args.BaseURL, args.DatabaseDNS, cl)

	if err != nil {
		slog.Error(err.Error(), slog.String("databaseDNS", args.DatabaseDNS))
		os.Exit(1)
	}

	linkShortener := shortener.New(repository)
	urlHandler := handler.New(linkShortener)

	http.NewServer(ctx, args.ServerAddress, urlHandler, cl).Start()
}
