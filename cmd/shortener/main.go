package main

import (
	"fmt"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/url"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http"
	"log"
)

func main() {
	us, err := url.NewService(url.WithMemoryRepository())

	if err != nil {
		fmt.Printf("%v", err)

		return
	}

	urlHandler := http.NewURLHandler(us)

	server := http.NewServer(":8080", urlHandler)
	log.Fatal(server.Start())
}
