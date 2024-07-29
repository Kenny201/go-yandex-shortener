package config

import (
	"flag"
	"os"
)

type Args struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
}

func NewArgs() *Args {
	return &Args{}
}

// ParseFlags Парсинг переменных из командной строки
func (a *Args) ParseFlags() {
	flag.StringVar(&a.ServerAddress, "a", ":8080", "Server address host:port")
	flag.StringVar(&a.BaseURL, "b", "http://localhost:8080", "Result net address host:port")
	flag.StringVar(&a.FileStoragePath, "f", "/tmp/index.txt", "File storage path")
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

	if fileStoragePath := os.Getenv("SHORTENER_FILE_STORAGE_PATH"); fileStoragePath != "" {
		a.FileStoragePath = fileStoragePath
	}
}

// SetArgs Установить аргументы
func (a *Args) SetArgs(serverAddress string, baseURL string, fileStoragePath string) {
	a.ServerAddress = serverAddress
	a.BaseURL = baseURL
	a.FileStoragePath = fileStoragePath
}
