package storage

import (
	"context"
	"database/sql"
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
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/closer"
)

const (
	errInsertFailed  = "unable to insert data to database"
	errGetByShortKey = "error get by short key"
)

var (
	ErrOpenDatabaseFailed  = fmt.Errorf("unable to connect to database")
	ErrCloseDatabaseFailed = fmt.Errorf("unable to close database connection")
	ErrOpenMigrateFailed   = fmt.Errorf("unable to open Migrate files")
	ErrPrepareStatement    = fmt.Errorf("error prepare statement")
	ErrScanQuery           = fmt.Errorf("error scan query")
	ErrCopyFrom            = fmt.Errorf("error copy from")
	ErrCopyCount           = fmt.Errorf("error differences in the amount of data copied")
	ErrorUrlAlreadyExist   = fmt.Errorf("duplicate key found")
	pgErr                  *pgconn.PgError
)

type DatabaseShortenerRepository struct {
	db          *pgx.Conn
	databaseDNS string
	baseURL     string
}

func NewDatabaseShortenerRepository(baseURL string, databaseDNS string) (*DatabaseShortenerRepository, error) {
	db, err := pgx.Connect(context.Background(), databaseDNS)

	if err != nil {
		return nil, ErrOpenDatabaseFailed
	}

	if db.Ping(context.Background()) != nil {
		db.Close(context.Background())
		return nil, ErrOpenDatabaseFailed
	}

	d := &DatabaseShortenerRepository{
		db:          db,
		databaseDNS: databaseDNS,
		baseURL:     baseURL,
	}

	closer.CL.Add(func(ctx context.Context) error {
		return d.close()
	})

	return d, nil
}

func (d *DatabaseShortenerRepository) Get(shortKey string) (*entity.URL, error) {
	url := entity.URL{}

	_, err := d.db.Prepare(context.Background(), "get_url", "SELECT id, short_url, original_url FROM shorteners WHERE  short_url = $1")

	if err != nil {
		return nil, fmt.Errorf("%s, %w, %v, func: Get", errGetByShortKey, ErrPrepareStatement, err.Error())
	}

	if err = d.db.QueryRow(context.Background(), "get_url", shortKey).Scan(&url.ID, &url.ShortKey, &url.OriginalURL); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s %w:%v,  func: Get", errGetByShortKey, ErrScanQuery, err.Error())
		}

		// нет ошибок и не нашлось записей
		return nil, fmt.Errorf("url %v not found, func: Get", shortKey)
	}

	return &url, nil
}

func (d *DatabaseShortenerRepository) Create(originalURL string) (string, error) {
	baseURL, err := valueobject.NewBaseURL(d.baseURL)

	if err != nil {
		return "", err
	}

	shortURL := valueobject.NewShortURL(baseURL)
	urlEntity := entity.NewURL(originalURL, shortURL.ShortKey())

	_, err = d.db.Prepare(context.Background(), "insert_url", "INSERT INTO shorteners (id, short_url, original_url) values ($1, $2, $3) RETURNING id, short_url, original_url")

	if err != nil {
		return "", fmt.Errorf("%s, %w, %v, func: Create", errInsertFailed, ErrPrepareStatement, err.Error())
	}

	if _, err = d.db.Exec(context.Background(), "insert_url", urlEntity.ID, urlEntity.ShortKey, urlEntity.OriginalURL); err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return "", fmt.Errorf("%s, %w, %v func:Create", errInsertFailed, fmt.Errorf("%w for id: %s", ErrorUrlAlreadyExist, urlEntity.ID), err.Error())
		}

		return "", fmt.Errorf("%s, %w, %v, func: Create", errInsertFailed, ErrScanQuery, err.Error())
	}

	return fmt.Sprintf("%s/%s", baseURL.ToString(), urlEntity.ShortKey), nil
}

func (d *DatabaseShortenerRepository) CreateList(urls []*entity.URLItem) ([]*entity.URLItem, error) {
	shortUrls := make([]*entity.URLItem, 0, len(urls))
	var linkedSubjects [][]interface{}
	baseURL, err := valueobject.NewBaseURL(d.baseURL)

	if err != nil {
		return nil, err
	}

	for _, v := range urls {
		shortURL := valueobject.NewShortURL(baseURL)
		linkedSubjects = append(linkedSubjects, []interface{}{v.ID, shortURL.ShortKey(), v.OriginalURL})

		shortUrls = append(
			shortUrls, &entity.URLItem{ID: v.ID, ShortURL: fmt.Sprintf("%s/%s", baseURL.ToString(), shortURL.ShortKey())},
		)
	}

	copyCount, err := d.db.CopyFrom(
		context.Background(),
		pgx.Identifier{"shorteners"},
		[]string{"id", "short_url", "original_url"},
		pgx.CopyFromRows(linkedSubjects),
	)

	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, fmt.Errorf("%s, %w, %v func:CreateList", errInsertFailed, fmt.Errorf("%w for id: %s", ErrorUrlAlreadyExist, linkedSubjects[copyCount][0]), err.Error())
		}

		return nil, fmt.Errorf("%s, %w, %v, func: CreateList", errInsertFailed, ErrCopyFrom, err.Error())
	}

	if int(copyCount) != len(linkedSubjects) {
		return nil, fmt.Errorf("%s, %w, %v, func: CreateList", errInsertFailed, ErrCopyCount, err.Error())
	}

	return shortUrls, nil
}

func (d *DatabaseShortenerRepository) close() error {
	if err := d.db.Close(context.Background()); err != nil {
		slog.Error(ErrCloseDatabaseFailed.Error(), slog.String("Error:", err.Error()))
		return err
	} else {
		slog.Info("Db connection gracefully closed")
	}

	return nil
}

func (d *DatabaseShortenerRepository) Migrate() error {
	m, err := migrate.New(
		"file://internal/migrations",
		d.databaseDNS,
	)

	if err != nil {
		return ErrOpenMigrateFailed
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %v Migrate", err.Error())
	}

	return nil
}

func (d *DatabaseShortenerRepository) CheckHealth() error {
	if err := d.db.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}
	return nil
}
