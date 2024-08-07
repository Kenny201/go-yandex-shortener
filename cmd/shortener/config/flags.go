package config

import (
	"flag"
	"os"
)

const (
	defaultServerAddress   = ":8080"
	defaultBaseURL         = "http://localhost:8080"
	defaultFileStoragePath = "tmp/Rquxc"

	infoServerAddress   = "Server address host:port"
	infoBaseURL         = "Result net address host:port"
	infoFileStoragePath = "File storage path"
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
	flag.StringVar(&a.ServerAddress, "a", defaultServerAddress, infoServerAddress)
	flag.StringVar(&a.BaseURL, "b", defaultBaseURL, infoBaseURL)
	flag.StringVar(&a.FileStoragePath, "f", defaultFileStoragePath, infoFileStoragePath)
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

	if fileStoragePath := os.Getenv("FILE_STORAGE_PATH"); fileStoragePath != "" {
		a.FileStoragePath = fileStoragePath
	}
}

// SetArgs Установить аргументы
func (a *Args) SetArgs(serverAddress, baseURL, fileStoragePath string) {
	a.ServerAddress = serverAddress
	a.BaseURL = baseURL
	a.FileStoragePath = fileStoragePath
}
