package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
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
)

type DatabaseShortenerRepository struct {
	DB          *sql.DB
	databaseDNS string
	baseURL     string
}

func NewDatabaseShortenerRepository(baseURL string, databaseDNS string) (*DatabaseShortenerRepository, error) {
	db, err := sql.Open("pgx", databaseDNS)

	if err != nil {
		return nil, ErrOpenDatabaseFailed
	}

	if db.Ping() != nil {
		db.Close()
		return nil, ErrOpenDatabaseFailed
	}

	d := &DatabaseShortenerRepository{
		DB:          db,
		databaseDNS: databaseDNS,
		baseURL:     baseURL,
	}

	closer.CL.Add(func(ctx context.Context) error {
		return d.close()
	})

	return d, nil
}

func (d *DatabaseShortenerRepository) Create(originalURL string) (string, error) {
	url := entity.URL{}
	baseURL, err := valueobject.NewBaseURL(d.baseURL)

	if err != nil {
		return "", err
	}

	value, err := d.checkExistsOriginalURL(originalURL)

	if value == nil && err != nil {
		return "", err
	} else if value != nil {
		return fmt.Sprintf("%s/%s", baseURL.ToString(), value.ShortKey), nil
	}

	shortURL := valueobject.NewShortURL(baseURL)
	urlEntity := entity.NewURL(originalURL, shortURL.ShortKey())

	stmt, err := d.DB.Prepare("INSERT INTO shorteners (short_url, original_url) values ($1, $2) RETURNING id, short_url, original_url")

	if err != nil {
		return "", fmt.Errorf("%s, %w, %v, func: Create", errInsertFailed, ErrPrepareStatement, err.Error())
	}

	if err := stmt.QueryRow(urlEntity.ShortKey, urlEntity.OriginalURL).Scan(&url.ID, &url.ShortKey, &url.OriginalURL); err != nil { // scan will release the connection
		return "", fmt.Errorf("%s, %w, %v, func: Create", errInsertFailed, ErrScanQuery, err.Error())
	}

	return fmt.Sprintf("%s/%s", baseURL.ToString(), url.ShortKey), nil
}

func (d *DatabaseShortenerRepository) Get(shortKey string) (*entity.URL, error) {
	url := &entity.URL{}

	stmt, err := d.DB.Prepare("SELECT id, short_url, original_url FROM shorteners WHERE  short_url = $1")

	if err != nil {
		return nil, fmt.Errorf("%s, %w, %v, func: Get", errGetByShortKey, ErrPrepareStatement, err.Error())
	}

	if err = stmt.QueryRow(shortKey).Scan(&url.ID, &url.ShortKey, &url.OriginalURL); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s %w:%v,  func: Get", errGetByShortKey, ErrScanQuery, err.Error())
		}

		// нет ошибок и не нашлось записей
		return nil, fmt.Errorf("url %v not found, func: Get", shortKey)
	}

	return url, nil
}

func (d *DatabaseShortenerRepository) GetAll() map[string]*entity.URL {
	//TODO implement me

	return nil
}

func (d *DatabaseShortenerRepository) checkExistsOriginalURL(originalURL string) (*entity.URL, error) {
	url := &entity.URL{}

	stmt, err := d.DB.Prepare("SELECT id, short_url, original_url FROM shorteners WHERE  original_url = $1")

	if err != nil {
		return nil, fmt.Errorf("%s, %w, %v func:checkExistsOriginalURL", errInsertFailed, ErrPrepareStatement, err.Error())
	}

	if err = stmt.QueryRow(originalURL).Scan(&url.ID, &url.ShortKey, &url.OriginalURL); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s, %w, %v func:checkExistsOriginalURL", errInsertFailed, ErrScanQuery, err.Error())
		}
		// нет ошибок и не нашлось записей
		return nil, nil
	}

	return url, nil
}

func (d *DatabaseShortenerRepository) close() error {
	if err := d.DB.Close(); err != nil {
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
