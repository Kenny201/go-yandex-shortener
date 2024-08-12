package shortener

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
)

type Repository interface {
	Get(id string) (*entity.URL, error)
	GetAll() map[string]*entity.URL
	Create(originalURL string) (string, error)
	CheckHealth() error
}

type Shortener struct {
	Repository Repository
}

func New(repository Repository) *Shortener {
	return &Shortener{Repository: repository}
}

// GetShortURL Получить сокращённую ссылку
func (s *Shortener) GetShortURL(shortKey string) (*entity.URL, error) {
	result, err := s.Repository.Get(shortKey)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// CreateShortURL Сохранить url в хранилище.
// Возвращает сокращённую ссылку
func (s *Shortener) CreateShortURL(originalURL string) (string, error) {
	// Сохраняем ссылку в хранилище и получаем обратно
	shortURL, err := s.Repository.Create(originalURL)

	if err != nil {
		return "", err
	}

	return shortURL, nil
}

func (s *Shortener) GetAllShortURL() map[string]*entity.URL {
	return s.Repository.GetAll()
}

func (s *Shortener) CheckHealth() error {
	if err := s.Repository.CheckHealth(); err != nil {
		return err
	}

	return nil
}
