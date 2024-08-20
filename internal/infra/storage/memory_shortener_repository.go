package storage

import (
	"fmt"
	"log/slog"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
)

type MemoryShortenerRepository struct {
	baseURL string
	urls    map[string]entity.URL
}

// NewMemoryShortenerRepository создает новый репозиторий сокращения ссылок в памяти.
func NewMemoryShortenerRepository(baseURL string) *MemoryShortenerRepository {
	return &MemoryShortenerRepository{
		baseURL: baseURL,
		urls:    make(map[string]entity.URL),
	}
}

// Get возвращает URL-адрес по короткому ключу, если он существует.
func (mr *MemoryShortenerRepository) Get(shortKey string) (*entity.URL, error) {
	for _, v := range mr.urls {
		if v.ShortKey == shortKey {
			slog.Info("URL retrieved successfully", slog.String("shortKey", shortKey))
			return &v, nil
		}
	}
	return nil, fmt.Errorf("URL %v not found", shortKey)
}

// Create добавляет новый URL в репозиторий, если его еще нет.
func (mr *MemoryShortenerRepository) Create(url *entity.URL) (*entity.URL, error) {
	if v, exists := mr.urls[url.OriginalURL]; exists {
		return &v, ErrURLAlreadyExist
	}

	mr.urls[url.OriginalURL] = *url
	return url, nil
}

// CreateList добавляет список новых URL в репозиторий, возвращая их сокращенные версии.
func (mr *MemoryShortenerRepository) CreateList(urls []*entity.URLItem) ([]*entity.URLItem, error) {
	shortUrls := make([]*entity.URLItem, 0, len(urls))
	baseURL, err := valueobject.NewBaseURL(mr.baseURL)

	if err != nil {
		return nil, err
	}

	for _, urlItem := range urls {
		shortURL := valueobject.NewShortURL(baseURL)

		if existingURL, exists := mr.urls[urlItem.OriginalURL]; exists {
			return []*entity.URLItem{{ID: urlItem.ID, ShortURL: fmt.Sprintf("%s/%s", mr.baseURL, existingURL.ShortKey)}}, ErrURLAlreadyExist
		}

		urlEntity := entity.URL{
			ID:          urlItem.ID,
			ShortKey:    shortURL.ShortKey(),
			OriginalURL: urlItem.OriginalURL,
		}

		shortUrls = append(shortUrls, &entity.URLItem{ID: urlEntity.ID, ShortURL: fmt.Sprintf("%s/%s", mr.baseURL, urlEntity.ShortKey)})
		mr.urls[urlItem.OriginalURL] = urlEntity
	}

	slog.Info("All URLs created successfully", slog.Int("count", len(shortUrls)))
	return shortUrls, nil
}

// CheckHealth проверяет состояние репозитория, возвращая ошибку, если он не инициализирован.
func (mr *MemoryShortenerRepository) CheckHealth() error {
	if mr.urls == nil {
		return fmt.Errorf("memory URLs structure is not initialized")
	}
	slog.Info("Memory repository health check passed")
	return nil
}
