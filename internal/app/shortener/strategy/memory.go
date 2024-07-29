package strategy

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

// Memory defines an implementation of a Strategy to execute.
type Memory struct {
	baseURL          string
	repositoryMemory *storage.RepositoryMemory
}

// NewMemory creates a new instance of strategy A.
func NewMemory(baseURL string) Strategy {
	return &Memory{
		baseURL:          baseURL,
		repositoryMemory: storage.NewRepositoryMemory(),
	}
}

func (s *Memory) Get(url string) (*entity.URL, error) {
	// Получить сокращённую ссылку из in-memory
	result, err := s.repositoryMemory.Get(url)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Memory) GetAll() (map[string]*entity.URL, error) {
	return s.repositoryMemory.GetAll()
}

func (s *Memory) Put(originalURL string) (string, error) {
	baseURL, err := valueobject.NewBaseURL(s.baseURL)

	if err != nil {
		return "", err
	}

	// Сохраняем ссылку в хранилище in-memory и получаем обратно
	shortURL, _ := s.repositoryMemory.Put(originalURL, baseURL)
	return shortURL, nil
}
