package main

import (
	"log"

	"github.com/Kenny201/go-yandex-shortener.git/cmd/shortener/config"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener/strategy"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http/handler"
)

func main() {
	args := config.NewArgs()
	args.ParseFlags()

	fileStrategy := strategy.NewFile(args.BaseURL, args.FileStoragePath)
	ss := shortener.NewService().SetStrategy(fileStrategy)
	urlHandler := handler.New(ss)

	server := http.NewServer(args.ServerAddress, urlHandler)

	log.Fatal(server.Start())
}
