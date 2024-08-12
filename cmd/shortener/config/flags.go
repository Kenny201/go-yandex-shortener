package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
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
	config          *Config
}

func NewArgs(config *Config) *Args {
	return &Args{
		config: config,
	}
}

// ParseFlags анализирует аргументы командной строки.
func (a *Args) ParseFlags(args []string) {
	fs := flag.NewFlagSet("args", flag.ContinueOnError)

	fs.StringVar(&a.ServerAddress, "a", fmt.Sprintf(":%s", a.config.Port), infoServerAddress)
	fs.StringVar(&a.BaseURL, "b", fmt.Sprintf("http://localhost:%s", a.config.Port), infoBaseURL)
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

// InitRepository инициализирует соответствующий репозиторий на основе предоставленных аргументов.
func (a *Args) InitRepository() (shortener.Repository, error) {
	switch {
	case a.DatabaseDNS != "":
		repo, err := storage.NewDatabaseShortenerRepository(a.BaseURL, a.DatabaseDNS)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database repository: %w", err)
		}

		if err := repo.Migrate(); err != nil {
			return nil, fmt.Errorf("failed to migrate database: %w", err)
		}

		return repo, nil

	case a.FileStoragePath != "":
		repo, err := storage.NewFileShortenerRepository(a.BaseURL, a.FileStoragePath)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize file repository: %w", err)
		}

		return repo, nil

	default:
		return storage.NewMemoryShortenerRepository(a.BaseURL), nil
	}
}
