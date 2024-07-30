package shortener

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener/strategy"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
)

type (
	Repository interface {
		Get(id string) (*entity.URL, error)
		GetAll() map[string]*entity.URL
		Put(url *entity.URL)
	}

	Service struct {
		strategy strategy.Strategy
	}
)

func NewService() *Service {
	s := &Service{}

	return s
}

func (s *Service) SetStrategy(strategy strategy.Strategy) *Service {
	s.strategy = strategy

	return s
}

// Get Получить сокращённую ссылку
func (s *Service) Get(shortKey string) (*entity.URL, error) {
	result, err := s.strategy.Get(shortKey)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Put Сохранить url в хранилище.
// Возвращает сокращённую ссылку
func (s *Service) Put(originalURL string) (string, error) {
	// Сохраняем ссылку в хранилище и получаем обратно
	shortURL, err := s.strategy.Put(originalURL)

	if err != nil {
		return "", err
	}

	return shortURL, nil
}

func (s *Service) GetAll() map[string]*entity.URL {
	return s.strategy.GetAll()
}
