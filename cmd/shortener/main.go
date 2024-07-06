package main

import (
	"fmt"
	"github.com/Kenny201/go-yandex-shortener.git/cmd/shortener/config"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/url"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http"
	"log"
)

func main() {
	config.ParseFlags()

	us, err := url.NewService(url.WithMemoryRepository())

	if err != nil {
		fmt.Printf("%v", err)

		return
	}

	urlHandler := http.NewURLHandler(us)

	server := http.NewServer(config.Args.NetAddressEntrance.Host, config.Args.NetAddressEntrance.Port, urlHandler)
	log.Fatal(server.Start())
}
