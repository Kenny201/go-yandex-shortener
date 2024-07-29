package main

import (
	"log"

	"github.com/Kenny201/go-yandex-shortener.git/cmd/shortener/config"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http/handler"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

func main() {
	args := config.NewArgs()
	args.ParseFlags()

	ss := shortener.NewService(args.BaseURL, storage.NewRepositoryMemory())

	urlHandler := handler.New(ss)

	server := http.NewServer(args.ServerAddress, urlHandler)

	log.Fatal(server.Start())
}
