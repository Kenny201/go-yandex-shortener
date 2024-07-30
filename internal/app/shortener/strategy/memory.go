package strategy

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

type Memory struct {
	baseURL    string
	repository *storage.RepositoryMemory
}

func NewMemory(baseURL string) Strategy {
	return &Memory{
		baseURL:    baseURL,
		repository: storage.NewRepositoryMemory(),
	}
}

func (memory *Memory) Get(url string) (*entity.URL, error) {
	// Получить сокращённую ссылку из in-memory
	result, err := memory.repository.Get(url)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (memory *Memory) GetAll() (map[string]*entity.URL, error) {
	return memory.repository.GetAll()
}

func (memory *Memory) Put(originalURL string) (string, error) {
	baseURL, err := valueobject.NewBaseURL(memory.baseURL)

	if err != nil {
		return "", err
	}

	// Сохраняем ссылку в хранилище in-memory и получаем обратно
	shortURL, _ := memory.repository.Put(originalURL, baseURL)
	return shortURL, nil
}
