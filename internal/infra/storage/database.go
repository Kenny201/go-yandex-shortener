package storage

import (
	"context"
	"errors"
	"fmt"

	"log/slog"
	"sync"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5"

	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage/repository"
	"github.com/Kenny201/go-yandex-shortener.git/internal/utils/closer"
)

var (
	ErrCloseDatabaseFailed = errors.New("unable to close database connection")
)

type DatabaseRepositories struct {
	shortener   shortener.ShortenerRepository
	db          *pgx.Conn
	databaseDNS string
}

// Singleton для хранения единственного экземпляра подключения
var (
	singletonDBConn *pgx.Conn
	mu              sync.Mutex
)

func NewDatabaseRepositories(baseURL string, databaseDNS string) (*DatabaseRepositories, error) {
	mu.Lock()
	defer mu.Unlock()

	// Проверяем, существует ли уже подключение
	if singletonDBConn == nil {
		db, err := pgx.Connect(context.Background(), databaseDNS)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", repository.ErrOpenDatabaseFailed, err)
		}

		// Если подключение успешно, сохраняем его в singleton
		singletonDBConn = db
	}

	// Создаем репозитории с уже существующим подключением
	d := &DatabaseRepositories{
		shortener:   repository.NewShortenerDatabase(baseURL, singletonDBConn),
		db:          singletonDBConn,
		databaseDNS: databaseDNS,
	}

	// Добавляем закрытие соединения в closer
	closer.CL.Add(d.Close)

	return d, nil
}

// Close Метод для закрытия соединения
func (d *DatabaseRepositories) Close(ctx context.Context) error {
	if d.db != nil {
		if err := d.db.Close(ctx); err != nil {
			return fmt.Errorf("%w: %v", ErrCloseDatabaseFailed, err)
		}
		d.db = nil
	}
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
