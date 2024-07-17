package main

import (
	"github.com/Kenny201/go-yandex-shortener.git/cmd/shortener/config"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
	"log"
)

func main() {
	args := config.NewArgs()
	args.ParseFlags()

	ss := shortener.NewService(args.BaseURL, storage.NewRepositoryMemory())

	urlHandler := http.NewShortenerHandler(ss)
	server := http.NewServer(args.ServerAddress, urlHandler)

	log.Fatal(server.Start())
}
