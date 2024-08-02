package storage

import (
	"fmt"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
)

type MemoryShortenerRepository struct {
	urls map[string]*entity.URL
}

func NewMemoryShortenerRepository() *MemoryShortenerRepository {
	return &MemoryShortenerRepository{
		urls: make(map[string]*entity.URL),
	}
}

func (rm *MemoryShortenerRepository) Get(shortKey string) (*entity.URL, error) {
	url, ok := rm.urls[shortKey]

	if !ok {
		err := fmt.Errorf("url %v not found", shortKey)
		return nil, err
	}

	return url, nil
}

func (rm *MemoryShortenerRepository) GetAll() map[string]*entity.URL {
	return rm.urls
}

// Put Добавить новый элемент
func (rm *MemoryShortenerRepository) Put(urlEntity *entity.URL) {
	rm.urls[urlEntity.ShortKey] = urlEntity
}
