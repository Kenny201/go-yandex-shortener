package storage

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage/repository"
)

type FileRepository struct {
	shortener shortener.ShortenerRepository
}

func NewFileRepositories(baseURL string, filePath string) (*FileRepository, error) {
	shortenerFile, err := repository.NewShortenerFile(baseURL, filePath)

	if err != nil {
		return nil, err
	}

	return &FileRepository{
		shortener: *shortenerFile,
	}, nil
}

func (r *FileRepository) GetShortenRepository() shortener.ShortenerRepository {
	return r.shortener
}
