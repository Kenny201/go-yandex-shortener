package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
	"github.com/Kenny201/go-yandex-shortener.git/internal/utils/closer"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	ErrOpenDatabaseFailed  = errors.New("unable to connect to database")
	ErrCloseDatabaseFailed = errors.New("unable to close database connection")
	ErrOpenMigrateFailed   = errors.New("unable to open migrate files")
	ErrScanQuery           = errors.New("error scan query")
	ErrCopyFrom            = errors.New("error copy from")
	ErrCopyCount           = errors.New("differences in the amount of data copied")
	ErrURLAlreadyExist     = errors.New("duplicate key found")
	ErrEmptyURL            = errors.New("empty URL list provided")
)

type DatabaseShortenerRepository struct {
	db          *pgx.Conn
	databaseDNS string
	baseURL     string
}

// NewDatabaseShortenerRepository создает новый экземпляр DatabaseShortenerRepository и устанавливает подключение к базе данных.
func NewDatabaseShortenerRepository(baseURL, databaseDNS string) (*DatabaseShortenerRepository, error) {
	db, err := pgx.Connect(context.Background(), databaseDNS)
	if err != nil {
		return nil, ErrOpenDatabaseFailed
	}

	repo := &DatabaseShortenerRepository{
		db:          db,
		databaseDNS: databaseDNS,
		baseURL:     baseURL,
	}

	// Добавляет функцию закрытия соединения в closer.
	closer.CL.Add(repo.close)
	return repo, nil
}

// Get извлекает информацию о коротком URL из базы данных по короткому ключу.
func (d *DatabaseShortenerRepository) Get(shortKey string) (*entity.URL, error) {
	query := "SELECT id, short_url, original_url FROM shorteners WHERE short_url = $1"
	url := &entity.URL{}

	err := d.db.QueryRow(context.Background(), query, shortKey).Scan(&url.ID, &url.ShortKey, &url.OriginalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("url %v не найден: %w", shortKey, ErrScanQuery)
		}
		return nil, fmt.Errorf("%w, func: Get", ErrScanQuery)
	}

	return url, nil
}

// Create создает новый короткий URL и сохраняет его в базе данных.
func (d *DatabaseShortenerRepository) Create(originalURL string) (string, error) {
	baseURL, err := valueobject.NewBaseURL(d.baseURL)
	if err != nil {
		return "", err
	}

	shortURL := valueobject.NewShortURL(baseURL)
	urlEntity := entity.NewURL(originalURL, shortURL.ShortKey())

	query := "INSERT INTO shorteners (id, short_url, original_url) VALUES ($1, $2, $3)"
	_, err = d.db.Exec(context.Background(), query, urlEntity.ID, urlEntity.ShortKey, urlEntity.OriginalURL)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			existingShortKey, getErr := d.getShortKeyByOriginalURL(originalURL)
			if getErr != nil {
				return "", fmt.Errorf("%w", getErr)
			}
			return fmt.Sprintf("%s/%s", baseURL.ToString(), existingShortKey), ErrURLAlreadyExist
		}
		return "", fmt.Errorf("%w", err)
	}

	return fmt.Sprintf("%s/%s", baseURL.ToString(), urlEntity.ShortKey), nil
}

// getShortKeyByOriginalURL возвращает существующий короткий ключ для заданного оригинального URL.
func (d *DatabaseShortenerRepository) getShortKeyByOriginalURL(originalURL string) (string, error) {
	var shortKey string
	query := "SELECT short_url FROM shorteners WHERE original_url = $1"

	err := d.db.QueryRow(context.Background(), query, originalURL).Scan(&shortKey)
	if err != nil {
		return "", fmt.Errorf("error retrieving short URL: %w", err)
	}

	return shortKey, nil
}

// CreateList добавляет список новых коротких URL в базу данных.
func (d *DatabaseShortenerRepository) CreateList(urls []*entity.URLItem) ([]*entity.URLItem, error) {
	if len(urls) == 0 {
		return nil, ErrEmptyURL
	}

	baseURL, err := valueobject.NewBaseURL(d.baseURL)
	if err != nil {
		return nil, err
	}

	linkedSubjects, shortUrls := d.prepareInsertData(urls, baseURL)

	copyCount, err := d.db.CopyFrom(
		context.Background(),
		pgx.Identifier{"shorteners"},
		[]string{"id", "short_url", "original_url"},
		pgx.CopyFromRows(linkedSubjects),
	)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return d.handleUniqueViolation(urls, baseURL)
		}
		return nil, d.handleCopyError(err, copyCount, linkedSubjects)
	}

	if int(copyCount) != len(linkedSubjects) {
		return nil, fmt.Errorf("%w: %d rows copied, expected %d", ErrCopyCount, copyCount, len(linkedSubjects))
	}

	return shortUrls, nil
}

// prepareInsertData подготавливает данные для вставки в базу данных.
func (d *DatabaseShortenerRepository) prepareInsertData(urls []*entity.URLItem, baseURL valueobject.BaseURL) ([][]interface{}, []*entity.URLItem) {
	linkedSubjects := make([][]interface{}, 0, len(urls))
	shortUrls := make([]*entity.URLItem, 0, len(urls))

	for _, v := range urls {
		shortURL := valueobject.NewShortURL(baseURL)
		linkedSubjects = append(linkedSubjects, []interface{}{v.ID, shortURL.ShortKey(), v.OriginalURL})
		shortUrls = append(shortUrls, &entity.URLItem{ID: v.ID, ShortURL: fmt.Sprintf("%s/%s", baseURL.ToString(), shortURL.ShortKey())})
	}

	return linkedSubjects, shortUrls
}

// handleUniqueViolation обрабатывает случаи нарушения уникальности при вставке данных в базу данных.
func (d *DatabaseShortenerRepository) handleUniqueViolation(urls []*entity.URLItem, baseURL valueobject.BaseURL) ([]*entity.URLItem, error) {
	duplicatedItems := make([]*entity.URLItem, 0)
	for _, v := range urls {
		existingShortKey, err := d.getShortKeyByOriginalURL(v.OriginalURL)
		if err == nil {
			duplicatedItems = append(duplicatedItems, &entity.URLItem{
				ID:       v.ID,
				ShortURL: fmt.Sprintf("%s/%s", baseURL.ToString(), existingShortKey),
			})
		}
	}
	return duplicatedItems, ErrURLAlreadyExist
}

// handleCopyError обрабатывает ошибки при копировании данных в базу данных.
func (d *DatabaseShortenerRepository) handleCopyError(err error, copyCount int64, linkedSubjects [][]interface{}) error {
	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
		return fmt.Errorf("%w for id: %s", ErrURLAlreadyExist, linkedSubjects[copyCount][0])
	}
	if int(copyCount) != len(linkedSubjects) {
		return fmt.Errorf("%w: %d rows copied, expected %d", ErrCopyCount, copyCount, len(linkedSubjects))
	}
	return fmt.Errorf("%w: %v", ErrCopyFrom, err)
}

// close закрывает соединение с базой данных.
func (d *DatabaseShortenerRepository) close(ctx context.Context) error {
	if err := d.db.Close(ctx); err != nil {
		slog.Error(ErrCloseDatabaseFailed.Error(), slog.String("Error", err.Error()))
		return err
	}
	slog.Info("Database connection gracefully closed")
	return nil
}

// Migrate выполняет миграцию базы данных.
func (d *DatabaseShortenerRepository) Migrate() error {
	m, err := migrate.New("file://internal/migrations", d.databaseDNS)
	if err != nil {
		return ErrOpenMigrateFailed
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %v", err)
	}
	return nil
}

// CheckHealth проверяет состояние соединения с базой данных.
func (d *DatabaseShortenerRepository) CheckHealth() error {
	if err := d.db.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}
	return nil
}
