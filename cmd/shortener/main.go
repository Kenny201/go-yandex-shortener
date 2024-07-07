package main

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/url"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http"
	"log"
)

func main() {
	us := url.NewService(url.WithMemoryRepository())

	urlHandler := http.NewURLHandler(us)

	server := http.NewServer(":8080", urlHandler)
	log.Fatal(server.Start())
}
