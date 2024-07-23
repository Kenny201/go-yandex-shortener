package storage

import (
	"fmt"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/aggregate"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
)

type RepositoryMemory struct {
	urls map[string]*aggregate.URL
}

func NewRepositoryMemory() *RepositoryMemory {
	return &RepositoryMemory{
		urls: make(map[string]*aggregate.URL),
	}
}

func (rm *RepositoryMemory) Get(id string) (*aggregate.URL, error) {
	url, ok := rm.urls[id]

	if !ok {
		err := fmt.Errorf("url %v not found", id)
		return nil, err
	}

	return url, nil
}

func (rm *RepositoryMemory) GetAll() map[string]*aggregate.URL {
	return rm.urls
}

func (rm *RepositoryMemory) Put(originalURL string, shortURL valueobject.ShortURL) *aggregate.URL {
	urlEntity := aggregate.NewURL(originalURL, shortURL)

	rm.urls[urlEntity.ID()] = urlEntity

	return urlEntity
}

func (rm *RepositoryMemory) CheckExistsOriginalURL(originalURL string) (*aggregate.URL, bool) {
	for _, value := range rm.urls {
		if value.OriginalURL() == originalURL {
			return value, true
		}
	}

	return nil, false
}
