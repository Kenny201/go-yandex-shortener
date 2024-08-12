package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/closer"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	errInsertFailed  = "unable to insert data to database"
	errGetByShortKey = "error get by short key"
)

var (
	ErrOpenDatabaseFailed  = errors.New("unable to connect to database")
	ErrCloseDatabaseFailed = errors.New("unable to close database connection")
	ErrOpenMigrateFailed   = errors.New("unable to open migrate files")
	ErrScanQuery           = errors.New("error scan query")
	ErrCopyFrom            = errors.New("error copy from")
	ErrCopyCount           = errors.New("error differences in the amount of data copied")
)

type DatabaseShortenerRepository struct {
	db          *pgx.Conn
	databaseDNS string
	baseURL     string
}

func NewDatabaseShortenerRepository(baseURL, databaseDNS string) (*DatabaseShortenerRepository, error) {
	db, err := pgx.Connect(context.Background(), databaseDNS)

	if err != nil {
		return nil, ErrOpenDatabaseFailed
	}

	if err := db.Ping(context.Background()); err != nil {
		db.Close(context.Background())
		return nil, ErrOpenDatabaseFailed
	}

	repo := &DatabaseShortenerRepository{
		db:          db,
		databaseDNS: databaseDNS,
		baseURL:     baseURL,
	}

	closer.CL.Add(func(ctx context.Context) error {
		return repo.close()
	})

	return repo, nil
}

func (d *DatabaseShortenerRepository) Get(shortKey string) (*entity.URL, error) {
	const query = "SELECT id, short_url, original_url FROM shorteners WHERE short_url = $1"
	url := &entity.URL{}

	if err := d.db.QueryRow(context.Background(), query, shortKey).Scan(&url.ID, &url.ShortKey, &url.OriginalURL); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("url %v not found, func: Get", shortKey)
		}
		return nil, fmt.Errorf("%s: %w, func: Get", errGetByShortKey, ErrScanQuery)
	}

	return url, nil
}

func (d *DatabaseShortenerRepository) Create(originalURL string) (string, error) {
	baseURL, err := valueobject.NewBaseURL(d.baseURL)

	if err != nil {
		return "", err
	}

	existingURL, err := d.checkExistsOriginalURL(originalURL)

	if err != nil {
		return "", err
	}

	if existingURL != nil {
		return fmt.Sprintf("%s/%s", baseURL.ToString(), existingURL.ShortKey), nil
	}

	shortURL := valueobject.NewShortURL(baseURL)
	urlEntity := entity.NewURL(originalURL, shortURL.ShortKey())

	const query = "INSERT INTO shorteners (id, short_url, original_url) VALUES ($1, $2, $3) RETURNING id, short_url, original_url"

	if err := d.db.QueryRow(context.Background(), query, urlEntity.ID, urlEntity.ShortKey, urlEntity.OriginalURL).
		Scan(&urlEntity.ID, &urlEntity.ShortKey, &urlEntity.OriginalURL); err != nil {
		return "", fmt.Errorf("%s: %w, func: Create", errInsertFailed, ErrScanQuery)
	}

	return fmt.Sprintf("%s/%s", baseURL.ToString(), urlEntity.ShortKey), nil
}

func (d *DatabaseShortenerRepository) CreateList(urls []*entity.URLItem) ([]*entity.URLItem, error) {
	shortUrls := make([]*entity.URLItem, 0, len(urls))
	baseURL, err := valueobject.NewBaseURL(d.baseURL)

	if err != nil {
		return nil, err
	}

	linkedSubjects := make([][]interface{}, 0, len(urls))

	for _, v := range urls {
		existingURL, err := d.checkExistsOriginalURL(v.OriginalURL)

		if err != nil {
			return nil, err
		}

		if existingURL != nil {
			shortUrls = append(shortUrls, &entity.URLItem{ID: existingURL.ID, ShortURL: fmt.Sprintf("%s/%s", baseURL.ToString(), existingURL.ShortKey)})
			continue
		}

		shortURL := valueobject.NewShortURL(baseURL)
		linkedSubjects = append(linkedSubjects, []interface{}{v.ID, shortURL.ShortKey(), v.OriginalURL})
		shortUrls = append(shortUrls, &entity.URLItem{ID: v.ID, ShortURL: fmt.Sprintf("%s/%s", baseURL.ToString(), shortURL.ShortKey())})
	}

	copyCount, err := d.db.CopyFrom(
		context.Background(),
		pgx.Identifier{"shorteners"},
		[]string{"id", "short_url", "original_url"},
		pgx.CopyFromRows(linkedSubjects),
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w, func: CreateList", errInsertFailed, ErrCopyFrom)
	}

	if int(copyCount) != len(linkedSubjects) {
		return nil, fmt.Errorf("%s: %w, func: CreateList", errInsertFailed, ErrCopyCount)
	}

	return shortUrls, nil
}

func (d *DatabaseShortenerRepository) checkExistsOriginalURL(originalURL string) (*entity.URL, error) {
	const query = "SELECT id, short_url, original_url FROM shorteners WHERE original_url = $1"
	url := &entity.URL{}

	if err := d.db.QueryRow(context.Background(), query, originalURL).Scan(&url.ID, &url.ShortKey, &url.OriginalURL); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w, func: checkExistsOriginalURL", errInsertFailed, ErrScanQuery)
	}

	return url, nil
}

func (d *DatabaseShortenerRepository) close() error {
	if err := d.db.Close(context.Background()); err != nil {
		slog.Error(ErrCloseDatabaseFailed.Error(), slog.String("Error", err.Error()))
		return err
	}

	slog.Info("Db connection gracefully closed")
	return nil
}

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

func (d *DatabaseShortenerRepository) CheckHealth() error {
	if err := d.db.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}
	return nil
}
