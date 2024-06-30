package main

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/app"
	"log"
)

func main() {
	server := app.NewServer(":8080")
	log.Fatal(server.Start())
}
