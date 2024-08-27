package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"sync"
	"time"

	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage/repository"
	"github.com/Kenny201/go-yandex-shortener.git/internal/utils/closer"
	"github.com/golang-migrate/migrate/v4"
	"log/slog"
)

var (
	ErrOpenDatabaseFailed = errors.New("unable to open database connection")
	Err                   = errors.New("error")
)

// singletonDBPool - пул подключений к базе данных.
var singletonDBPool *pgxpool.Pool
var mu sync.Mutex

type DatabaseRepositories struct {
	shortener   shortener.ShortenerRepository
	db          *pgxpool.Pool
	databaseDNS string
}

func NewDatabaseRepositories(baseURL string, databaseDNS string) (*DatabaseRepositories, error) {
	mu.Lock()
	defer mu.Unlock()

	// Проверяем, существует ли уже подключение
	if singletonDBPool == nil {
		config, err := pgxpool.ParseConfig(databaseDNS)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", Err, err)
		}

		// Настройка параметров пула
		config.MaxConns = 100                    // Максимальное количество соединений
		config.MinConns = 10                     // Минимальное количество соединений
		config.MaxConnIdleTime = 5 * time.Minute // Время ожидания неактивного соединения
		config.MaxConnLifetime = 1 * time.Hour   // Время жизни соединения

		dbPool, err := pgxpool.NewWithConfig(context.Background(), config)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrOpenDatabaseFailed, err)
		}

		// Если подключение успешно, сохраняем его в singleton
		singletonDBPool = dbPool
	}

	// Создаем репозитории с уже существующим пулом
	d := &DatabaseRepositories{
		shortener:   repository.NewShortenerDatabase(baseURL, singletonDBPool),
		db:          singletonDBPool,
		databaseDNS: databaseDNS,
	}

	// Добавляем закрытие соединения в closer
	closer.CL.Add(d.Close)

	return d, nil
}

// Close закрывает соединение с базой данных.
func (d *DatabaseRepositories) Close(ctx context.Context) error {
	// Закрытие пула подключений, ошибки не возвращаются
	singletonDBPool.Close()
	slog.Info("Database connection gracefully closed")
	return nil
}

// Migrate This migrate all tables
func (d *DatabaseRepositories) Migrate() error {
	m, err := migrate.New("file://internal/migrations", d.databaseDNS)
	if err != nil {
		return fmt.Errorf("%w: %v", repository.ErrOpenMigrateFailed, err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %v", err)
	}
	slog.Info("Database migration successful")
	return nil
}

func (d *DatabaseRepositories) GetShortenerRepository() shortener.ShortenerRepository {
	return d.shortener
}
