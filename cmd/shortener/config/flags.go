package config

import (
	"flag"
	"os"
)

var Args struct {
	ServerAddress string
	BaseURL       string
}

func ParseFlags() {
	flag.StringVar(&Args.ServerAddress, "a", ":8080", "server address host:port")
	flag.StringVar(&Args.BaseURL, "b", "http://localhost:8080", "Result net address host:port")
	flag.Parse()

	setArgsFromEnv()
}

func setArgsFromEnv() {
	if serverAddr := os.Getenv("SERVER_ADDRESS"); serverAddr != "" {
		Args.ServerAddress = serverAddr
	}

	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		Args.BaseURL = baseURL
	}
}
