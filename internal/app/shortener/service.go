package shortener

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/aggregate"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
)

type Repository interface {
	Get(id string) (*aggregate.URL, error)
	GetAll() map[string]*aggregate.URL
	Put(originalURL string, shortURL valueobject.ShortURL) *aggregate.URL
	CheckExistsOriginalURL(originalURL string) (*aggregate.URL, bool)
}

type Service struct {
	BaseURL string
	Sr      Repository
}

func NewService(baseURL string, repository Repository) *Service {
	s := &Service{}
	s.Sr = repository
	s.BaseURL = baseURL

	return s
}

// Put Сохранить url в хранилище. Возвращает сгенерированную короткую ссылку
func (s *Service) Put(url string) (string, error) {
	baseURL, err := valueobject.NewBaseURL(s.BaseURL)
	if err != nil {
		return "", err
	}

	shortURL := valueobject.NewShortURL(baseURL)

	if len(s.Sr.GetAll()) != 0 {
		if key, ok := s.Sr.CheckExistsOriginalURL(url); ok {
			return key.ShortURL().ToString(), nil
		}
	}

	urlEntity := s.Sr.Put(url, shortURL)

	return urlEntity.ShortURL().ToString(), nil
}

// Get Получить сокращённую ссылку по id
func (s *Service) Get(url string) (*aggregate.URL, error) {
	result, err := s.Sr.Get(url)

	if err != nil {
		return nil, err
	}

	return result, nil
}
