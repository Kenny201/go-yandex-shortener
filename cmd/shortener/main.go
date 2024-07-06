package main

import (
	"fmt"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/url"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http"
	"log"
)

func main() {
	us, err := url.NewUrlService(url.WithMemoryUrlRepository())

	if err != nil {
		fmt.Printf("%v", err)
	}

	urlHandler := http.NewUrlHandler(us)

	server := http.NewServer(":8080", urlHandler)
	log.Fatal(server.Start())
}
