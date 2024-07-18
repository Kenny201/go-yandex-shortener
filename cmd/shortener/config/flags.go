package config

import (
	"flag"
	"os"
)

type Args struct {
	ServerAddress string
	BaseURL       string
}

func NewArgs() *Args {
	return &Args{}
}

// ParseFlags Парсинг переменных из коман
func (a *Args) ParseFlags() {
	flag.StringVar(&a.ServerAddress, "a", ":8080", "Server address host:port")
	flag.StringVar(&a.BaseURL, "b", "http://localhost:8080", "Result net address host:port")
	flag.Parse()

	a.setArgsFromEnv()
}

// Установить аргументы из переменной окружения
func (a *Args) setArgsFromEnv() {
	if serverAddr := os.Getenv("SHORTENER_SERVER_ADDRESS"); serverAddr != "" {
		a.ServerAddress = serverAddr
	}

	if baseURL := os.Getenv("SHORTENER_BASE_URL"); baseURL != "" {
		a.BaseURL = baseURL
	}
}

// SetArgs Установить аргументы
func (a *Args) SetArgs(serverAddress string, BaseURL string) {
	a.ServerAddress = serverAddress
	a.BaseURL = BaseURL
}
