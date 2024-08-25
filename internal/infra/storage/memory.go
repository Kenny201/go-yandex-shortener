package storage

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage/repository"
)

type MemoryRepository struct {
	shortener shortener.ShortenerRepository
}

func NewMemoryRepositories(baseURL string) *MemoryRepository {
	return &MemoryRepository{
		shortener: repository.NewShortenerMemory(baseURL),
	}
}

func (r *MemoryRepository) GetShortenerRepository() shortener.ShortenerRepository {
	return r.shortener
}
