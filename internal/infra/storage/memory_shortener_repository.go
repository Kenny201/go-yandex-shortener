package storage

import (
	"fmt"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
)

type MemoryShortenerRepository struct {
	baseURL string
	urls    map[string]*entity.URL
}

func NewMemoryShortenerRepository(baseURL string) *MemoryShortenerRepository {
	return &MemoryShortenerRepository{
		baseURL: baseURL,
		urls:    make(map[string]*entity.URL),
	}
}

func (rm *MemoryShortenerRepository) Get(shortKey string) (*entity.URL, error) {
	url, ok := rm.urls[shortKey]

	if !ok {
		err := fmt.Errorf("url %v not found", shortKey)
		return nil, err
	}

	return url, nil
}

func (rm *MemoryShortenerRepository) GetAll() map[string]*entity.URL {
	return rm.urls
}

// Put Добавить новый элемент
func (rm *MemoryShortenerRepository) Put(originalURL string) (string, error) {
	baseURL, err := valueobject.NewBaseURL(rm.baseURL)

	if err != nil {
		return "", err
	}

	//  Проверка существования записи в мапе urls.
	if value, ok := rm.checkExistsOriginalURL(originalURL); ok {
		return fmt.Sprintf("%s/%s", baseURL.ToString(), value.ShortKey), nil
	}

	shortURL := valueobject.NewShortURL(baseURL)
	urlEntity := entity.NewURL(originalURL, shortURL.ShortKey())

	// Сохраняем ссылку в хранилище in-memory
	rm.urls[urlEntity.ShortKey] = urlEntity

	return shortURL.ToString(), nil
}

// Проверка существования записи в мапе
func (rm *MemoryShortenerRepository) checkExistsOriginalURL(originalURL string) (*entity.URL, bool) {
	for _, value := range rm.GetAll() {
		if value.OriginalURL == originalURL {
			return value, true
		}
	}

	return nil, false
}
