package main

import (
	"github.com/Kenny201/go-yandex-shortener.git/cmd/shortener/config"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http"
	"log"
)

func main() {
	ss := shortener.NewService(shortener.WithMemoryRepository())
	config.ParseFlags()

	urlHandler := http.NewShortenerHandler(ss)
	server := http.NewServer(config.Args.ServerAddress, urlHandler)

	log.Fatal(server.Start())
}
