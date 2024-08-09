package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/closer"
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
func (a *Args) ParseFlags() {
	defaultServerAddress := fmt.Sprintf(":%s", a.config.Port)
	defaultBaseURL := fmt.Sprintf("http://localhost:%s", a.config.Port)

	flag.StringVar(&a.ServerAddress, "a", defaultServerAddress, infoServerAddress)
	flag.StringVar(&a.BaseURL, "b", defaultBaseURL, infoBaseURL)
	flag.StringVar(&a.FileStoragePath, "f", "", infoFileStoragePath)
	flag.StringVar(&a.DatabaseDNS, "d", "", infoDatabaseDNS)
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

	if databaseDNS := os.Getenv("DATABASE_DSN"); databaseDNS != "" {
		a.DatabaseDNS = databaseDNS
	}
}

// SetArgs Установить аргументы
func (a *Args) SetArgs(serverAddress, baseURL, fileStoragePath string) {
	a.ServerAddress = serverAddress
	a.BaseURL = baseURL
	a.FileStoragePath = fileStoragePath
}

func (a *Args) InitArgs() {
	a.ServerAddress = fmt.Sprintf(":%s", a.config.Port)
	a.BaseURL = fmt.Sprintf("http://localhost:%s", a.config.Port)
	a.FileStoragePath = "tmp/Rquxc"
	a.DatabaseDNS = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", a.config.DBUsername, a.config.DBPassword, a.config.DBHost, a.config.DBPort, a.config.DBDatabase)
}

func (a *Args) InitRepository(cl *closer.Closer) (shortener.Repository, error) {
	if a.DatabaseDNS != "" {
		repository, err := storage.NewDatabaseShortenerRepository(a.BaseURL, a.DatabaseDNS, cl)

		if err != nil {
			return nil, err
		}

		err = repository.Migrate()

		if err != nil {
			return nil, err
		}

		return repository, nil

	} else if a.FileStoragePath != "" {
		repository, err := storage.NewFileShortenerRepository(a.BaseURL, a.FileStoragePath)

		if err != nil {
			return nil, err
		}

		return repository, nil
	} else {
		repository := storage.NewMemoryShortenerRepository(a.BaseURL)
		return repository, nil
	}
}
