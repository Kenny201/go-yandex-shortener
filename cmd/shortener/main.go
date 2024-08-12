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
	"github.com/Kenny201/go-yandex-shortener.git/internal/utils/closer"
)

func main() {
	closer.New()
	conf, err := config.LoadConfig("./")
	var linkShortener *shortener.Shortener

	if err != nil {
		slog.Error("error read config %v", slog.String("error", err.Error()))
		os.Exit(1)
	}

	args := config.NewArgs(conf)
	args.ParseFlags(os.Args[1:])

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	repository, err := args.InitRepository()

	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	linkShortener = shortener.New(repository)

	urlHandler := handler.New(linkShortener)

	http.NewServer(ctx, args.ServerAddress, urlHandler).Start()
}
