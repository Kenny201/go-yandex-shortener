package shortener

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
)

type Repository interface {
	Get(id string) (*entity.URL, error)
	GetAll() map[string]*entity.URL
	Put(originalURL string) (string, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

// Get Получить сокращённую ссылку
func (s *Service) Get(shortKey string) (*entity.URL, error) {
	result, err := s.repository.Get(shortKey)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Put Сохранить url в хранилище.
// Возвращает сокращённую ссылку
func (s *Service) Put(originalURL string) (string, error) {
	// Сохраняем ссылку в хранилище и получаем обратно
	shortURL, err := s.repository.Put(originalURL)

	if err != nil {
		return "", err
	}

	return shortURL, nil
}

func (s *Service) GetAll() map[string]*entity.URL {
	return s.repository.GetAll()
}
