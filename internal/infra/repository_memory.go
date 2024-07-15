package infra

import (
	"errors"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/aggregate"
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
	if _, ok := rm.urls[id]; !ok {
		err := errors.New("url not found")
		return nil, err
	}

	return rm.urls[id], nil
}

func (rm *RepositoryMemory) GetAll() map[string]*aggregate.URL {
	return rm.urls
}

func (rm *RepositoryMemory) Put(url *aggregate.URL) (*aggregate.URL, error) {
	rm.urls[url.ID()] = url

	return url, nil
}

func (rm *RepositoryMemory) CheckExistsBaseURL(baseURL string) (*aggregate.URL, bool) {
	for _, value := range rm.urls {

		if value.BaseURL().ToString() == baseURL {
			return value, true
		}
	}

	return nil, false
}
