package shortener

import (
	"github.com/Kenny201/go-yandex-shortener.git/cmd/shortener/config"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/aggregate"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra"
)

type Storage func(s *Service)

type Service struct {
	Sr shortener.Repository
}

func NewService(storages ...Storage) *Service {
	s := &Service{}

	for _, storage := range storages {
		storage(s)
	}

	return s
}

func WithRepository(sr shortener.Repository) Storage {
	return func(s *Service) {
		s.Sr = sr
	}
}

func WithRepositoryMemory() Storage {
	mr := infra.NewRepositoryMemory()

	return WithRepository(mr)
}

func (s *Service) Put(url string) (string, error) {
	var body string

	originalURL, err := valueobject.NewOriginalURL(url)

	if err != nil {
		return "", err
	}

	baseURL, err := valueobject.NewBaseURL(config.Args.BaseURL)

	if err != nil {
		return "", err
	}

	shortURL := valueobject.NewShortURL(baseURL)

	if len(s.Sr.GetAll()) != 0 {
		if key, ok := s.Sr.CheckExistsOriginalURL(url); ok {
			body = key.ShortURL().ToString()
		} else {
			urlEntity := aggregate.NewURL(originalURL, shortURL)
			urlEntity, err = s.Sr.Put(urlEntity)
			body = urlEntity.ShortURL().ToString()
		}
	} else {
		urlEntity := aggregate.NewURL(originalURL, shortURL)
		urlEntity, err = s.Sr.Put(urlEntity)
		body = urlEntity.ShortURL().ToString()
	}

	if err != nil {
		return "", err
	}

	return body, nil
}

func (s *Service) Get(url string) (*aggregate.URL, error) {
	result, err := s.Sr.Get(url)

	if err != nil {
		return nil, err
	}

	return result, nil
}
