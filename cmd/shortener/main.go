package main

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http"
	"log"
)

func main() {
	ss := shortener.NewService(shortener.WithMemoryRepository())

	urlHandler := http.NewShortenerHandler(ss)

	server := http.NewServer(":8080", urlHandler)
	log.Fatal(server.Start())
}
