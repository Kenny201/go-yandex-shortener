package infra

import (
	"errors"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/url/entity"
)

type RepositoryMemory struct {
	urls map[string]*entity.URL
}

func NewMemoryRepositories() *RepositoryMemory {
	return &RepositoryMemory{
		urls: make(map[string]*entity.URL),
	}
}

func (r *RepositoryMemory) Get(id string) (*entity.URL, error) {
	if _, ok := r.urls[id]; !ok {
		err := errors.New("short url not found")
		return nil, err
	}

	return r.urls[id], nil
}

func (r *RepositoryMemory) GetAll() []entity.URL {
	var urls []entity.URL

	for _, url := range r.urls {
		urls = append(urls, *url)
	}

	return urls
}

func (r *RepositoryMemory) Put(url *entity.URL) (*entity.URL, error) {
	r.urls[url.ID()] = url

	return url, nil
}

func (r *RepositoryMemory) CheckExistsOriginal(shortValue string) (*entity.URL, bool) {
	for _, value := range r.urls {
		if value.OriginalURL() == shortValue {
			return value, true
		}
	}

	return nil, false
}
