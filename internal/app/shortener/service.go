package shortener

import (
	"github.com/Kenny201/go-yandex-shortener.git/cmd/shortener/config"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/aggregate"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra"
	"net/http"
)

type Storage func(ss *Service)

type Service struct {
	Sr shortener.Repository
}

func NewService(storages ...Storage) *Service {
	ss := &Service{}

	for _, storage := range storages {
		storage(ss)
	}

	return ss
}

func WithRepository(sr shortener.Repository) Storage {
	return func(ss *Service) {
		ss.Sr = sr
	}
}

func WithMemoryRepository() Storage {
	mr := infra.NewMemoryRepositories()

	return WithRepository(mr)
}

func (ss *Service) Put(url string, r *http.Request) (string, error) {
	var body valueobject.ShortURL

	originalURL, err := valueobject.NewOriginalURL(url)

	if err != nil {
		return "", err
	}

	baseURL, err := valueobject.NewBaseURL(config.Args.BaseURL)

	if err != nil {
		return "", err
	}

	shortURL := valueobject.NewShortURL(baseURL)

	if len(ss.Sr.GetAll()) != 0 {
		if key, ok := ss.Sr.CheckExistsOriginalURL(url); ok {
			body = key.ShortURL()
		} else {
			urlEntity := aggregate.NewURL(originalURL, shortURL)
			urlEntity, err = ss.Sr.Put(urlEntity)
			body = urlEntity.ShortURL()
		}
	} else {
		urlEntity := aggregate.NewURL(originalURL, shortURL)
		urlEntity, err = ss.Sr.Put(urlEntity)
		body = urlEntity.ShortURL()
	}

	if err != nil {
		return "", err
	}

	return body.ToString(), nil
}

func (ss *Service) Get(url string) (*aggregate.URL, error) {
	result, err := ss.Sr.Get(url)
	if err != nil {
		return nil, err
	}

	return result, nil
}
