package storage

import (
	"fmt"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
)

type RepositoryMemory struct {
	urls map[string]*entity.URL
}

func NewRepositoryMemory() *RepositoryMemory {
	return &RepositoryMemory{
		urls: make(map[string]*entity.URL),
	}
}

func (rm *RepositoryMemory) Get(shortKey string) (*entity.URL, error) {
	url, ok := rm.urls[shortKey]

	if !ok {
		err := fmt.Errorf("url %v not found", shortKey)
		return nil, err
	}

	return url, nil
}

func (rm *RepositoryMemory) GetAll() map[string]*entity.URL {
	return rm.urls
}

// Put Добавить новый элемент
func (rm *RepositoryMemory) Put(urlEntity *entity.URL) {
	rm.urls[urlEntity.ShortKey] = urlEntity
}
