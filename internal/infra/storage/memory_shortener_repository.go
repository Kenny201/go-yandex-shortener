package storage

import (
	"fmt"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
)

type MemoryShortenerRepository struct {
	baseURL string
	urls    map[string]entity.URL
}

func NewMemoryShortenerRepository(baseURL string) *MemoryShortenerRepository {
	return &MemoryShortenerRepository{
		baseURL: baseURL,
		urls:    make(map[string]entity.URL),
	}
}

func (rm *MemoryShortenerRepository) Get(shortKey string) (*entity.URL, error) {
	for _, v := range rm.urls {
		if v.ShortKey == shortKey {
			return &v, nil
		}
	}

	return nil, fmt.Errorf("url %v not found", shortKey)
}

// Create Добавить новый элемент
func (rm *MemoryShortenerRepository) Create(originalURL string) (string, error) {
	baseURL, err := valueobject.NewBaseURL(rm.baseURL)

	if err != nil {
		return "", err
	}

	//  Проверка существования записи в мапе urls.
	if value, ok := rm.urls[originalURL]; ok {
		return fmt.Sprintf("%s/%s", baseURL.ToString(), value.ShortKey), ErrorUrlAlreadyExist
	}

	shortURL := valueobject.NewShortURL(baseURL)
	urlEntity := entity.NewURL(originalURL, shortURL.ShortKey())

	// Сохраняем ссылку в хранилище in-memory
	rm.urls[urlEntity.OriginalURL] = *urlEntity

	return shortURL.ToString(), nil
}

func (rm *MemoryShortenerRepository) CreateList(urls []*entity.URLItem) ([]*entity.URLItem, error) {
	baseURL, err := valueobject.NewBaseURL(rm.baseURL)
	shortUrls := make([]*entity.URLItem, 0, len(urls))

	if err != nil {
		return nil, err
	}

	for _, v := range urls {
		//  Проверка существования записи в мапе urls.
		if v, ok := rm.urls[v.OriginalURL]; ok {
			duplicateShortUrls := make([]*entity.URLItem, 0, len(urls))

			duplicateShortUrls = append(
				duplicateShortUrls,
				&entity.URLItem{ID: v.ID, ShortURL: fmt.Sprintf("%s/%s", baseURL.ToString(), v.ShortKey)},
			)

			return duplicateShortUrls, ErrorUrlAlreadyExist
		}

		shortURL := valueobject.NewShortURL(baseURL)

		urlEntity := &entity.URL{ID: v.ID, ShortKey: shortURL.ShortKey(), OriginalURL: v.OriginalURL}

		// Сохраняем ссылку в хранилище in-memory
		rm.urls[urlEntity.OriginalURL] = *urlEntity

		shortUrls = append(
			shortUrls,
			&entity.URLItem{ID: v.ID, ShortURL: fmt.Sprintf("%s/%s", baseURL.ToString(), shortURL.ShortKey())},
		)
	}

	return shortUrls, nil
}
