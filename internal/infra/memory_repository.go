package infra

import (
	"errors"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/aggregate"
)

type RepositoryMemory struct {
	urls map[string]*aggregate.URL
}

func NewMemoryRepositories() *RepositoryMemory {
	return &RepositoryMemory{
		urls: make(map[string]*aggregate.URL),
	}
}

func (r *RepositoryMemory) Get(id string) (*aggregate.URL, error) {
	if _, ok := r.urls[id]; !ok {
		err := errors.New("short shortener not found")
		return nil, err
	}

	return r.urls[id], nil
}

func (r *RepositoryMemory) GetAll() []aggregate.URL {
	var urls []aggregate.URL

	for _, url := range r.urls {
		urls = append(urls, *url)
	}

	return urls
}

func (r *RepositoryMemory) Put(url *aggregate.URL) (*aggregate.URL, error) {
	r.urls[url.ID()] = url

	return url, nil
}

func (r *RepositoryMemory) CheckExistsOriginalURL(originalURL string) (*aggregate.URL, bool) {
	for _, value := range r.urls {
		if value.OriginalURL().ToString() == originalURL {
			return value, true
		}
	}

	return nil, false
}
