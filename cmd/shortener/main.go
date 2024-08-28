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
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
	"github.com/Kenny201/go-yandex-shortener.git/internal/utils/closer"
)

func main() {
	closer.New()

	err := config.LoadConfig("./")
	if err != nil {
		slog.Error("error read config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	args := config.NewArgs()
	args.ParseFlags(os.Args[1:])

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	repository, err := initializeRepository(args)

	if err != nil {
		slog.Error("failed to initialize repository", slog.String("error", err.Error()))
		os.Exit(1)
	}

	shortenerService := shortener.New(repository, args.BaseURL)

	urlHandler := handler.New(shortenerService)

	http.NewServer(ctx, args.ServerAddress, urlHandler).Start()
}

func initializeRepository(args *config.Args) (shortener.Repository, error) {
	switch {
	case args.DatabaseDNS != "":
		repo, err := storage.NewShortenerDatabase(args.BaseURL, args.DatabaseDNS)
		if err != nil {
			return nil, err
		}
		if err := repo.Migrate(); err != nil {
			return nil, err
		}
		return repo, nil

	case args.FileStoragePath != "":
		return storage.NewShortenerFile(args.BaseURL, args.FileStoragePath)

	default:
		return storage.NewShortenerMemory(args.BaseURL), nil
	}
}
