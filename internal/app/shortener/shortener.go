package shortener

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
)

type Repository interface {
	Get(id string) (*entity.URL, error)
	GetAll() map[string]*entity.URL
	Put(originalURL string) (string, error)
}

type Shortener struct {
	repository Repository
}

func New(repository Repository) *Shortener {
	return &Shortener{repository: repository}
}

// Get Получить сокращённую ссылку
func (s *Shortener) Get(shortKey string) (*entity.URL, error) {
	result, err := s.repository.Get(shortKey)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Put Сохранить url в хранилище.
// Возвращает сокращённую ссылку
func (s *Shortener) Put(originalURL string) (string, error) {
	// Сохраняем ссылку в хранилище и получаем обратно
	shortURL, err := s.repository.Put(originalURL)

	if err != nil {
		return "", err
	}

	return shortURL, nil
}

func (s *Shortener) GetAll() map[string]*entity.URL {
	return s.repository.GetAll()
}
