package main

import (
	"github.com/Kenny201/go-yandex-shortener.git/cmd/shortener/config"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/url"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http"
	"log"
)

func main() {
	err := config.ParseFlags()

	if err != nil {
		log.Fatal(err)
	}

	us := url.NewService(url.WithMemoryRepository())

	urlHandler := http.NewURLHandler(us)

	server := http.NewServer(config.Args.NetAddressEntrance.Host, config.Args.NetAddressEntrance.Port, urlHandler)
	log.Fatal(server.Start())
}
