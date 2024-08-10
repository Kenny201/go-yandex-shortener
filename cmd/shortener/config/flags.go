package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

const (
	infoServerAddress   = "Server address host:port"
	infoBaseURL         = "Result net address host:port"
	infoFileStoragePath = "File storage path"
	infoDatabaseDNS     = "Database DNS format: postgres://username:password@host:port/dbname?sslmode=disable"
)

type Args struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDNS     string
	config          *Config
}

func NewArgs(config *Config) *Args {
	args := &Args{}

	args.config = config

	return args
}

// ParseFlags Парсинг переменных из командной строки
func (a *Args) ParseFlags(args []string) {
	defaultServerAddress := fmt.Sprintf(":%s", a.config.Port)
	defaultBaseURL := fmt.Sprintf("http://localhost:%s", a.config.Port)

	fs := flag.NewFlagSet("args", flag.ContinueOnError)
	fs.StringVar(&a.ServerAddress, "a", defaultServerAddress, infoServerAddress)
	fs.StringVar(&a.BaseURL, "b", defaultBaseURL, infoBaseURL)
	fs.StringVar(&a.FileStoragePath, "f", "", infoFileStoragePath)
	fs.StringVar(&a.DatabaseDNS, "d", "", infoDatabaseDNS)
	fs.Parse(args)

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

	if databaseDNS := os.Getenv("DATABASE_DSN"); databaseDNS != "" {
		a.DatabaseDNS = databaseDNS
	}
}

func (a *Args) InitRepository() (shortener.Repository, error) {
	if a.DatabaseDNS != "" {
		repository, err := storage.NewDatabaseShortenerRepository(a.BaseURL, a.DatabaseDNS)

		if err != nil {
			return nil, fmt.Errorf("repository database initialization error: %v", err)
		}

		err = repository.Migrate()

		if err != nil {
			return nil, err
		}

		return repository, nil
	} else if a.FileStoragePath != "" {
		repository, err := storage.NewFileShortenerRepository(a.BaseURL, a.FileStoragePath)

		if err != nil {
			return nil, fmt.Errorf("repository file initialization error: %v", err)
		}

		return repository, nil
	} else {
		repository := storage.NewMemoryShortenerRepository(a.BaseURL)
		return repository, nil
	}
}
