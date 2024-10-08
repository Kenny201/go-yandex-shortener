package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
	"github.com/Kenny201/go-yandex-shortener.git/internal/utils/closer"
)

var (
	ErrParseConfigPGXPool = errors.New("error parse config for pgxpool")
	ErrOpenDatabaseFailed = errors.New("unable to open database connection")
	ErrOpenMigrateFailed  = errors.New("unable to open migrate files")
	ErrCopyFrom           = errors.New("error during copy operation")
	ErrCopyCount          = errors.New("discrepancy in copied data count")
	ErrURLAlreadyExist    = errors.New("duplicate key found")
	ErrEmptyURL           = errors.New("empty URL list provided")
	ErrUserListURL        = errors.New("no short URLs found for user ID")
	ErrURLDeleted         = errors.New("URL is deleted")
	ErrURLNotFound        = errors.New("URL not found")
)

type ShortenerDatabase struct {
	db          *pgxpool.Pool
	databaseDNS string
	baseURL     string
}

// NewShortenerDatabase создает новый экземпляр ShortenerDatabase и устанавливает подключение к базе данных.
func NewShortenerDatabase(baseURL string, databaseDNS string) (*ShortenerDatabase, error) {
	config, err := pgxpool.ParseConfig(databaseDNS)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParseConfigPGXPool, err)
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

	repo := &ShortenerDatabase{
		db:          dbPool,
		databaseDNS: databaseDNS,
		baseURL:     baseURL,
	}

	// Добавляем закрытие соединения в closer
	closer.CL.Add(repo.Close)

	return repo, nil
}

// Close закрывает соединение с базой данных.
func (dr *ShortenerDatabase) Close(ctx context.Context) error {
	// Закрытие пула подключений, ошибки не возвращаются
	dr.db.Close()
	slog.Info("Database connection gracefully closed")
	return nil
}

// Get извлекает информацию о коротком URL из базы данных по короткому ключу.
func (dr *ShortenerDatabase) Get(shortKey string) (*entity.URL, error) {
	url := &entity.URL{}
	query := "SELECT id, short_key, original_url, is_deleted FROM shorteners WHERE short_key = $1"

	if err := dr.db.QueryRow(context.Background(), query, shortKey).Scan(&url.ID, &url.ShortKey, &url.OriginalURL, &url.DeletedFlag); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrURLNotFound
		}
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	if url.DeletedFlag {
		return nil, ErrURLDeleted // URL помечен как удаленный
	}

	slog.Info("URL retrieved", slog.String("shortKey", shortKey), slog.String("originalURL", url.OriginalURL))
	return url, nil
}

// Create добавляет новый URL в базу данных.
func (dr *ShortenerDatabase) Create(url *entity.URL) (*entity.URL, error) {
	query := "INSERT INTO shorteners (id, user_id, short_key, original_url) VALUES ($1,$2,$3,$4)"
	_, err := dr.db.Exec(context.Background(), query, url.ID, url.UserID, url.ShortKey, url.OriginalURL)
	if err != nil {
		if pgErr := parsePGError(err); pgErr != nil {
			return dr.handleDuplicateURL(url.OriginalURL)
		}
		return nil, err
	}
	return url, nil
}

// handleDuplicateURL обрабатывает ситуацию с дублирующимся URL.
func (dr *ShortenerDatabase) handleDuplicateURL(originalURL string) (*entity.URL, error) {
	existingURL, err := dr.findExistingURL(originalURL)
	if err != nil {
		return nil, err
	}
	return existingURL, ErrURLAlreadyExist
}

// findExistingURL ищет уже существующий URL в базе данных по оригинальному URL.
func (dr *ShortenerDatabase) findExistingURL(originalURL string) (*entity.URL, error) {
	var existingURL entity.URL
	query := "SELECT id, short_key, original_url FROM shorteners WHERE original_url = $1"

	if err := dr.db.QueryRow(context.Background(), query, originalURL).Scan(&existingURL.ID, &existingURL.ShortKey, &existingURL.OriginalURL); err != nil {
		return nil, err
	}
	return &existingURL, nil
}

// CreateList добавляет список новых URL в базу данных и возвращает их сокращенные версии.
func (dr *ShortenerDatabase) CreateList(userID interface{}, urls []*entity.URLItem) ([]*entity.URLItem, error) {
	if len(urls) == 0 {
		return nil, ErrEmptyURL
	}

	baseURL, err := valueobject.NewBaseURL(dr.baseURL)
	if err != nil {
		return nil, err
	}

	var (
		duplicates = make([]*entity.URLItem, 0, len(urls))
		shortURLs  = make([]*entity.URLItem, 0, len(urls))
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
		}

		shortURL := valueobject.NewShortURL(baseURL)
		urlItem.ShortKey = shortURL.ShortKey()
		urlItem.ShortURL = fmt.Sprintf("%s/%s", baseURL.ToString(), urlItem.ShortKey)
		urlBatch = append(urlBatch, []interface{}{urlItem.ID, userID, urlItem.ShortKey, urlItem.OriginalURL})
		shortURLs = append(shortURLs, &entity.URLItem{ID: urlItem.ID, ShortURL: fmt.Sprintf("%s/%s", baseURL.ToString(), urlItem.ShortKey)})
	}

	if err := dr.copyURLsToDB(urlBatch); err != nil {
		return nil, err
	}

	return shortURLs, nil
}

// GetAll получает все ссылки определённого пользователя
func (dr *ShortenerDatabase) GetAll(userID string) ([]*entity.URLItem, error) {
	var shortURLs []*entity.URLItem

	query := `
        SELECT short_key, original_url
        FROM shorteners 
        WHERE user_id = $1`

	rows, err := dr.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get short URLs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var shortURL entity.URLItem
		if err := rows.Scan(&shortURL.ShortURL, &shortURL.OriginalURL); err != nil {
			return nil, fmt.Errorf("failed to scan short URL: %w", err)
		}
		shortURL.ShortURL = fmt.Sprintf("%s/%s", dr.baseURL, shortURL.ShortURL)
		shortURLs = append(shortURLs, &shortURL)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Если ссылки не найдены
	if len(shortURLs) == 0 {
		return nil, fmt.Errorf("%s:%s", ErrUserListURL, userID)
	}

	return shortURLs, nil
}

// MarkAsDeleted обновляет поле is_deleted на true для списка коротких URL.
func (dr *ShortenerDatabase) MarkAsDeleted(batch []string, userID string) error {
	batchObj := &pgx.Batch{}

	for _, key := range batch {
		query := "UPDATE shorteners SET is_deleted = true WHERE short_key = $1 AND user_id = $2"
		batchObj.Queue(query, key, userID)
	}

	return dr.executeBatch(batchObj)
}

// executeBatch выполняет пакетный запрос и передает ошибки в errorsChan.
func (dr *ShortenerDatabase) executeBatch(batch *pgx.Batch) error {
	br := dr.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		tag, err := br.Exec()
		if err != nil {
			slog.Error("Failed to execute batch query", slog.String("error", err.Error()))
			return fmt.Errorf("failed to update URL: %w", err)
		}
		if tag.RowsAffected() == 0 {
			slog.Warn("No rows affected for URL", slog.Int("batchIndex", i))
		}
	}
	return nil
}

// copyURLsToDB копирует данные URL в базу данных с использованием CopyFrom.
func (dr *ShortenerDatabase) copyURLsToDB(urlBatch [][]interface{}) error {
	rowsCopied, err := dr.db.CopyFrom(
		context.Background(),
		pgx.Identifier{"shorteners"},
		[]string{"id", "user_id", "short_key", "original_url"},
		pgx.CopyFromRows(urlBatch),
	)
	if err != nil || int(rowsCopied) != len(urlBatch) {
		return dr.handleCopyError(err, rowsCopied, len(urlBatch))
	}
	return nil
}

// handleCopyError обрабатывает ошибки при копировании данных в базу данных.
func (dr *ShortenerDatabase) handleCopyError(err error, rowsCopied int64, expectedRows int) error {
	if pgErr := parsePGError(err); pgErr != nil {
		return fmt.Errorf("%w for id: %v", ErrURLAlreadyExist, pgErr)
	}
	if int(rowsCopied) != expectedRows {
		return fmt.Errorf("%w: %d rows copied, expected %d", ErrCopyCount, rowsCopied, expectedRows)
	}
	return fmt.Errorf("%w: %v", ErrCopyFrom, err)
}

func parsePGError(err error) *pgconn.PgError {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return pgErr
	}
	return nil
}

// Migrate выполняет миграцию базы данных.
func (dr *ShortenerDatabase) Migrate() error {
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
func (dr *ShortenerDatabase) CheckHealth() error {
	if err := dr.db.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}
	slog.Info("Database connection is healthy")
	return nil
}
