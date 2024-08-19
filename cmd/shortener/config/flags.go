package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

const (
	infoServerAddress   = "Server address (host:port)"
	infoBaseURL         = "Base URL (host:port)"
	infoFileStoragePath = "Path to file storage"
	infoDatabaseDNS     = "Database DNS (e.g., postgres://username:password@host:port/dbname?sslmode=disable)"
)

type Args struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDNS     string
}

func NewArgs() *Args {
	return &Args{}
}

// ParseFlags анализирует аргументы командной строки.
func (a *Args) ParseFlags(args []string) {
	fs := flag.NewFlagSet("args", flag.ContinueOnError)

	fs.StringVar(&a.ServerAddress, "a", fmt.Sprintf(":%s", viper.GetString("APP_PORT")), infoServerAddress)
	fs.StringVar(&a.BaseURL, "b", fmt.Sprintf("http://localhost:%s", viper.GetString("APP_PORT")), infoBaseURL)
	fs.StringVar(&a.FileStoragePath, "f", "", infoFileStoragePath)
	fs.StringVar(&a.DatabaseDNS, "d", "", infoDatabaseDNS)

	_ = fs.Parse(args) // Игнорировать ошибку, поскольку она обрабатывается флагом flag.ContinueOnError

	a.overrideWithEnvVars()
}

// overrideWithEnvVars переопределяет аргументы переменными среды, если они установлены.
func (a *Args) overrideWithEnvVars() {
	if v := os.Getenv("SHORTENER_SERVER_ADDRESS"); v != "" {
		a.ServerAddress = v
	}
	if v := os.Getenv("SHORTENER_BASE_URL"); v != "" {
		a.BaseURL = v
	}
	if v := os.Getenv("FILE_STORAGE_PATH"); v != "" {
		a.FileStoragePath = v
	}
	if v := os.Getenv("DATABASE_DSN"); v != "" {
		a.DatabaseDNS = v
	}
}
