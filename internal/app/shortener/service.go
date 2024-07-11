package shortener

import (
	"fmt"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/aggregate"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra"
	"net/http"
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

func WithMemoryRepository() Storage {
	mr := infra.NewMemoryRepositories()
	return WithRepository(mr)
}

func (s *Service) Put(url string, r *http.Request) (string, error) {
	var body string
	var host string

	if r.TLS != nil {
		host = fmt.Sprintf("https://%s", r.Host)
	} else {
		host = fmt.Sprintf("http://%s", r.Host)
	}

	serverAddress, err := valueobject.NewServerAddress(host)

	if err != nil {
		return "", err
	}

	baseURL, err := valueobject.NewBaseURL(url)

	if err != nil {
		return "", err
	}

	shortURL := valueobject.NewShortURL(serverAddress)

	if len(s.Sr.GetAll()) != 0 {
		if key, ok := s.Sr.CheckExistsBaseURL(url); ok {
			body = key.ShortURL().ToString()
		} else {
			urlEntity := aggregate.NewURL(baseURL, shortURL)
			urlEntity, err = s.Sr.Put(urlEntity)
			body = urlEntity.ShortURL().ToString()
		}
	} else {
		urlEntity := aggregate.NewURL(baseURL, shortURL)
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
