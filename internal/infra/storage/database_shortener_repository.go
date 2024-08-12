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
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

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
	ErrCopyCount           = errors.New("differences in the amount of data copied")
	ErrUrlAlreadyExist     = errors.New("duplicate key found")
)

type DatabaseShortenerRepository struct {
	db          *pgx.Conn
	databaseDNS string
	baseURL     string
}

func NewDatabaseShortenerRepository(baseURL, databaseDNS string) (*DatabaseShortenerRepository, error) {
	db, err := pgx.Connect(context.Background(), databaseDNS)
	if err != nil || db.Ping(context.Background()) != nil {
		if db != nil {
			db.Close(context.Background())
		}
		return nil, ErrOpenDatabaseFailed
	}

	repo := &DatabaseShortenerRepository{
		db:          db,
		databaseDNS: databaseDNS,
		baseURL:     baseURL,
	}

	closer.CL.Add(repo.close)
	return repo, nil
}

func (d *DatabaseShortenerRepository) Get(shortKey string) (*entity.URL, error) {
	query := "SELECT id, short_url, original_url FROM shorteners WHERE short_url = $1"
	url := &entity.URL{}

	if err := d.db.QueryRow(context.Background(), query, shortKey).Scan(&url.ID, &url.ShortKey, &url.OriginalURL); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("url %v not found: %w", shortKey, errGetByShortKey)
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

	shortURL := valueobject.NewShortURL(baseURL)
	urlEntity := entity.NewURL(originalURL, shortURL.ShortKey())

	query := "INSERT INTO shorteners (id, short_url, original_url) values ($1, $2, $3)"

	if _, err := d.db.Exec(context.Background(), query, urlEntity.ID, urlEntity.ShortKey, urlEntity.OriginalURL); err != nil {
		return d.handleInsertError(err, urlEntity.ID)
	}

	return fmt.Sprintf("%s/%s", baseURL.ToString(), urlEntity.ShortKey), nil
}

func (d *DatabaseShortenerRepository) CreateList(urls []*entity.URLItem) ([]*entity.URLItem, error) {
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

	if err != nil || int(copyCount) != len(linkedSubjects) {
		return nil, d.handleCopyError(err, copyCount, linkedSubjects)
	}

	return shortUrls, nil
}

func (d *DatabaseShortenerRepository) prepareInsertData(urls []*entity.URLItem, baseURL valueobject.BaseURL) ([][]interface{}, []*entity.URLItem) {
	var linkedSubjects [][]interface{}
	shortUrls := make([]*entity.URLItem, 0, len(urls))

	for _, v := range urls {
		shortURL := valueobject.NewShortURL(baseURL)
		linkedSubjects = append(linkedSubjects, []interface{}{v.ID, shortURL.ShortKey(), v.OriginalURL})
		shortUrls = append(shortUrls, &entity.URLItem{ID: v.ID, ShortURL: fmt.Sprintf("%s/%s", baseURL.ToString(), shortURL.ShortKey())})
	}

	return linkedSubjects, shortUrls
}

func (d *DatabaseShortenerRepository) handleInsertError(err error, id string) (string, error) {
	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
		return "", fmt.Errorf("%w for id: %s", ErrUrlAlreadyExist, id)
	}
	return "", fmt.Errorf("%s: %w", errInsertFailed, err)
}

func (d *DatabaseShortenerRepository) handleCopyError(err error, copyCount int64, linkedSubjects [][]interface{}) error {
	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
		return fmt.Errorf("%w for id: %s", ErrUrlAlreadyExist, linkedSubjects[copyCount][0])
	}
	if int(copyCount) != len(linkedSubjects) {
		return fmt.Errorf("%w: %d rows copied, expected %d", ErrCopyCount, copyCount, len(linkedSubjects))
	}
	return fmt.Errorf("%w: %v", ErrCopyFrom, err)
}

func (d *DatabaseShortenerRepository) close(ctx context.Context) error {
	if err := d.db.Close(ctx); err != nil {
		slog.Error(ErrCloseDatabaseFailed.Error(), slog.String("Error", err.Error()))
		return err
	}
	slog.Info("Database connection gracefully closed")
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
