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
	var err error
	var repository shortener.Repository

	args := config.NewArgs()
	args.ParseFlags()

	repository, err = storage.NewFile(args.BaseURL, args.FileStoragePath)
	ss := shortener.NewService(repository)
	urlHandler := handler.New(ss)

	server := http.NewServer(args.ServerAddress, urlHandler)
	err = server.Start()

	if err != nil {
		log.Fatal(err)
	}
}
