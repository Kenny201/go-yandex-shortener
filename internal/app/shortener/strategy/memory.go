package strategy

import (
	"fmt"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

type Memory struct {
	baseURL    string
	repository *storage.InMemoryShortenerRepository
}

func NewMemory(baseURL string) Strategy {
	return &Memory{
		baseURL:    baseURL,
		repository: storage.NewInMemoryShortenerRepository(),
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

func (memory *Memory) GetAll() map[string]*entity.URL {
	return memory.repository.GetAll()
}

// Put Возвращает сокращённую строку формата: url/shortKey
func (memory *Memory) Put(originalURL string) (string, error) {
	baseURL, err := valueobject.NewBaseURL(memory.baseURL)

	if err != nil {
		return "", err
	}

	//  Проверка существования записи в мапе urls.
	if value, ok := memory.checkExistsOriginalURL(originalURL); ok {
		return fmt.Sprintf("%s/%s", baseURL.ToString(), value.ShortKey), nil
	}

	shortURL := valueobject.NewShortURL(baseURL)
	urlEntity := entity.NewURL(originalURL, shortURL.ShortKey())

	// Сохраняем ссылку в хранилище in-memory
	memory.repository.Put(urlEntity)

	return shortURL.ToString(), nil
}

// Проверка существования записи в мапе
func (memory *Memory) checkExistsOriginalURL(originalURL string) (*entity.URL, bool) {
	for _, value := range memory.GetAll() {
		if value.OriginalURL == originalURL {
			return value, true
		}
	}

	return nil, false
}
