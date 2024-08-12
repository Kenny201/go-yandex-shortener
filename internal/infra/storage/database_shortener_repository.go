package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/closer"
)

var (
	ErrOpenDatabaseFailed  = fmt.Errorf("unable to connect to database")
	ErrCloseDatabaseFailed = fmt.Errorf("unable to close database connection")
)

type DatabaseShortenerRepository struct {
	db          *sql.DB
	databaseDNS string
	baseURL     string
	closer      *closer.Closer
}

func NewDatabaseShortenerRepository(baseURL string, databaseDNS string, closer *closer.Closer) (*DatabaseShortenerRepository, error) {
	db, err := sql.Open("pgx", databaseDNS)

	if err != nil {
		return nil, ErrOpenDatabaseFailed
	}

	if db.Ping() != nil {
		db.Close()
		return nil, ErrOpenDatabaseFailed
	}

	d := &DatabaseShortenerRepository{
		db:          db,
		databaseDNS: databaseDNS,
		baseURL:     baseURL,
		closer:      closer,
	}

	d.closer.Add(func(ctx context.Context) error {
		return d.close()
	})

	return d, nil
}

func (d *DatabaseShortenerRepository) Create(originalURL string) (string, error) {
	//TODO implement me

	return "", nil
}

func (d *DatabaseShortenerRepository) Get(shortKey string) (*entity.URL, error) {
	//TODO implement me

	return nil, nil
}

func (d *DatabaseShortenerRepository) GetAll() map[string]*entity.URL {
	//TODO implement me

	return nil
}

func (d *DatabaseShortenerRepository) checkExistsOriginalURL(originalURL string) (*entity.URL, bool) {
	//TODO implement me

	return nil, false
}

func (d *DatabaseShortenerRepository) close() error {
	if err := d.db.Close(); err != nil {
		slog.Error(ErrCloseDatabaseFailed.Error(), slog.String("Error:", err.Error()))
		return err
	} else {
		slog.Info("Db connection gracefully closed")
	}

	return nil
}

func (d *DatabaseShortenerRepository) CheckHealth() error {
	if err := d.db.Ping(); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}
	return nil
}
