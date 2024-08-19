package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
	"github.com/Kenny201/go-yandex-shortener.git/internal/utils/closer"
)

// Ошибки, возвращаемые репозиторием.
var (
	ErrOpenDatabaseFailed  = errors.New("unable to connect to database")
	ErrCloseDatabaseFailed = errors.New("unable to close database connection")
	ErrOpenMigrateFailed   = errors.New("unable to open migrate files")
	ErrCopyFrom            = errors.New("error during copy operation")
	ErrCopyCount           = errors.New("discrepancy in copied data count")
	ErrURLAlreadyExist     = errors.New("duplicate key found")
	ErrEmptyURL            = errors.New("empty URL list provided")
)

// DatabaseShortenerRepository предоставляет методы для работы с URL-ами в базе данных.
type DatabaseShortenerRepository struct {
	db          *pgx.Conn
	databaseDNS string
	baseURL     string
}

// NewDatabaseShortenerRepository создает новый экземпляр DatabaseShortenerRepository и устанавливает подключение к базе данных.
func NewDatabaseShortenerRepository(baseURL, databaseDNS string) (*DatabaseShortenerRepository, error) {
	db, err := pgx.Connect(context.Background(), databaseDNS)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOpenDatabaseFailed, err)
	}

	repo := &DatabaseShortenerRepository{
		db:          db,
		databaseDNS: databaseDNS,
		baseURL:     baseURL,
	}

	// Добавляем функцию закрытия соединения в closer.
	closer.CL.Add(repo.close)
	return repo, nil
}

// Get извлекает информацию о коротком URL из базы данных по короткому ключу.
func (dr *DatabaseShortenerRepository) Get(shortKey string) (*entity.URL, error) {
	url := &entity.URL{}

	query := "SELECT id, short_key, original_url FROM shorteners WHERE short_key = $1"

	err := dr.db.QueryRow(context.Background(), query, shortKey).Scan(&url.ID, &url.ShortKey, &url.OriginalURL)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("URL %v not found", shortKey)
		}
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	slog.Info("URL retrieved", slog.String("shortKey", shortKey), slog.String("originalURL", url.OriginalURL))
	return url, nil
}

// Create добавляет новый URL в базу данных.
func (dr *DatabaseShortenerRepository) Create(url *entity.URL) (*entity.URL, error) {
	query := "INSERT INTO shorteners (id, short_key, original_url) VALUES ($1, $2, $3)"

	_, err := dr.db.Exec(context.Background(), query, url.ID, url.ShortKey, url.OriginalURL)

	if err != nil {
		if pgErr := handlePGError(err); pgErr != nil {
			return dr.handleDuplicateURL(url.OriginalURL)
		}
		return nil, err
	}

	return url, nil
}

// handleDuplicateURL обрабатывает ситуацию с дублирующимся URL.
func (dr *DatabaseShortenerRepository) handleDuplicateURL(originalURL string) (*entity.URL, error) {
	existingURL, err := dr.findExistingURL(originalURL)

	if err != nil {
		return nil, err
	}

	return existingURL, ErrURLAlreadyExist
}

// findExistingURL ищет уже существующий URL в базе данных по оригинальному URL.
func (dr *DatabaseShortenerRepository) findExistingURL(originalURL string) (*entity.URL, error) {
	var existingURL entity.URL
	query := "SELECT id, short_key, original_url FROM shorteners WHERE original_url = $1"

	err := dr.db.QueryRow(context.Background(), query, originalURL).Scan(&existingURL.ID, &existingURL.ShortKey, &existingURL.OriginalURL)

	if err != nil {
		return nil, err
	}

	return &existingURL, nil
}

// CreateList добавляет список новых URL в базу данных и возвращает их сокращенные версии.
func (dr *DatabaseShortenerRepository) CreateList(urls []*entity.URLItem) ([]*entity.URLItem, error) {
	if len(urls) == 0 {
		return nil, ErrEmptyURL
	}
	baseURL, err := valueobject.NewBaseURL(dr.baseURL)

	if err != nil {
		return nil, err
	}

	var (
		duplicates []*entity.URLItem
		shortURLs  []*entity.URLItem
		urlBatch   [][]interface{}
	)

	for _, urlItem := range urls {
		existingURL, err := dr.findExistingURL(urlItem.OriginalURL)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

		if existingURL != nil {
			duplicates = append(duplicates, &entity.URLItem{ID: existingURL.ID, ShortURL: fmt.Sprintf("%s/%s", baseURL.ToString(), existingURL.ShortKey)})
			return duplicates, ErrURLAlreadyExist
		} else {
			shortURL := valueobject.NewShortURL(baseURL)
			urlItem.ShortKey = shortURL.ShortKey()
			urlItem.ShortURL = fmt.Sprintf("%s/%s", baseURL.ToString(), urlItem.ShortKey)
			urlBatch = append(urlBatch, []interface{}{urlItem.ID, urlItem.ShortKey, urlItem.OriginalURL})
			shortURLs = append(shortURLs, &entity.URLItem{ID: urlItem.ID, ShortURL: fmt.Sprintf("%s/%s", baseURL.ToString(), urlItem.ShortKey)})
		}
	}

	if err := dr.copyURLsToDB(urlBatch); err != nil {
		return nil, err
	}

	return shortURLs, nil
}

// copyURLsToDB копирует данные URL в базу данных с использованием CopyFrom.
func (dr *DatabaseShortenerRepository) copyURLsToDB(urlBatch [][]interface{}) error {
	rowsCopied, err := dr.db.CopyFrom(
		context.Background(),
		pgx.Identifier{"shorteners"},
		[]string{"id", "short_key", "original_url"},
		pgx.CopyFromRows(urlBatch),
	)
	if err != nil || int(rowsCopied) != len(urlBatch) {
		return dr.handleCopyError(err, rowsCopied, len(urlBatch))
	}
	return nil
}

// handleCopyError обрабатывает ошибки при копировании данных в базу данных.
func (dr *DatabaseShortenerRepository) handleCopyError(err error, rowsCopied int64, expectedRows int) error {
	if pgErr := handlePGError(err); pgErr != nil {
		return fmt.Errorf("%w for id: %v", ErrURLAlreadyExist, pgErr)
	}
	if int(rowsCopied) != expectedRows {
		return fmt.Errorf("%w: %d rows copied, expected %d", ErrCopyCount, rowsCopied, expectedRows)
	}
	return fmt.Errorf("%w: %v", ErrCopyFrom, err)
}

// handlePGError обрабатывает ошибки Postgres.
func handlePGError(err error) *pgconn.PgError {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return pgErr
	}
	return nil
}

// close закрывает соединение с базой данных.
func (dr *DatabaseShortenerRepository) close(ctx context.Context) error {
	if err := dr.db.Close(ctx); err != nil {
		slog.Error(ErrCloseDatabaseFailed.Error(), slog.String("error", err.Error()))
		return err
	}
	slog.Info("Database connection gracefully closed")
	return nil
}

// Migrate выполняет миграцию базы данных.
func (dr *DatabaseShortenerRepository) Migrate() error {
	m, err := migrate.New("file://internal/migrations", dr.databaseDNS)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOpenMigrateFailed, err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %v", err)
	}
	slog.Info("Database migration successful")
	return nil
}

// CheckHealth проверяет состояние соединения с базой данных.
func (dr *DatabaseShortenerRepository) CheckHealth() error {
	if err := dr.db.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}
	slog.Info("Database connection is healthy")
	return nil
}
