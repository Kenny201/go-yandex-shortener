package storage

import (
	"fmt"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
)

type RepositoryMemory struct {
	urls map[string]*entity.URL
}

func NewRepositoryMemory() *RepositoryMemory {
	return &RepositoryMemory{
		urls: make(map[string]*entity.URL),
	}
}

func (rm *RepositoryMemory) Get(shortKey string) (*entity.URL, error) {
	url, ok := rm.urls[shortKey]

	if !ok {
		err := fmt.Errorf("url %v not found", shortKey)
		return nil, err
	}

	return url, nil
}

func (rm *RepositoryMemory) GetAll() (map[string]*entity.URL, error) {
	return rm.urls, nil
}

// Put Добавить новый элемент
// Возвращает сокращённую строку формата: url/shortKey
func (rm *RepositoryMemory) Put(originalURL string, baseURL valueobject.BaseURL) (string, error) {
	//  Проверка существования записи в мапе urls.
	if value, ok := rm.checkExistsOriginalURL(originalURL); ok {
		return fmt.Sprintf("%s/%s", baseURL.ToString(), value.ShortKey), nil
	}

	shortURL := valueobject.NewShortURL(baseURL)
	urlEntity := entity.NewURL(originalURL, shortURL.ShortKey())
	rm.urls[urlEntity.ShortKey] = urlEntity

	return shortURL.ToString(), nil
}

// Проверка существования записи в мапе
func (rm *RepositoryMemory) checkExistsOriginalURL(originalURL string) (*entity.URL, bool) {
	for _, value := range rm.urls {
		if value.OriginalURL == originalURL {
			return value, true
		}
	}

	return nil, false
}
