package storage

import (
	"errors"
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
func (rm *MemoryShortenerRepository) Get(shortKey string) (*entity.URL, error) {
	for _, v := range rm.urls {
		if v.ShortKey == shortKey {
			slog.Info("URL retrieved successfully", slog.String("shortKey", shortKey))
			return &v, nil
		}
	}
	return nil, fmt.Errorf("url %v not found", shortKey)
}

// Create добавляет новый URL в репозиторий, если его еще нет.
func (rm *MemoryShortenerRepository) Create(originalURL string) (string, error) {
	shortURL, err := rm.findOrCreateURL(originalURL)
	if err != nil {
		return shortURL, err
	}
	slog.Info("URL created successfully", slog.String("originalURL", originalURL), slog.String("shortURL", shortURL))
	return shortURL, nil
}

// CreateList добавляет список новых URL в репозиторий, возвращая их сокращенные версии.
func (rm *MemoryShortenerRepository) CreateList(urls []*entity.URLItem) ([]*entity.URLItem, error) {
	if len(urls) == 0 {
		return nil, ErrEmptyURL
	}

	shortUrls := make([]*entity.URLItem, 0, len(urls))

	for _, urlItem := range urls {
		shortURL, err := rm.findOrCreateURL(urlItem.OriginalURL)

		if err != nil {
			if errors.Is(err, ErrURLAlreadyExist) {
				return []*entity.URLItem{{ID: urlItem.ID, ShortURL: shortURL}}, err
			}
			return nil, err
		}

		shortUrls = append(shortUrls, &entity.URLItem{ID: urlItem.ID, ShortURL: shortURL})
	}

	slog.Info("All URLs created successfully", slog.Int("count", len(shortUrls)))
	return shortUrls, nil
}

// CheckHealth проверяет состояние репозитория, возвращая ошибку, если он не инициализирован.
func (rm *MemoryShortenerRepository) CheckHealth() error {
	if rm.urls == nil {
		return fmt.Errorf("memory urls structure is not initialized")
	}
	slog.Info("Memory repository health check passed")
	return nil
}

// findOrCreateURL ищет существующий URL или создает новый, если его нет в репозитории.
func (rm *MemoryShortenerRepository) findOrCreateURL(originalURL string) (string, error) {
	baseURL, err := valueobject.NewBaseURL(rm.baseURL)
	if err != nil {
		return "", err
	}

	// Проверка существования записи в карте urls.
	if value, ok := rm.urls[originalURL]; ok {
		return fmt.Sprintf("%s/%s", baseURL.ToString(), value.ShortKey), ErrURLAlreadyExist
	}

	shortURL := valueobject.NewShortURL(baseURL)
	urlEntity := entity.NewURL(originalURL, shortURL.ShortKey())

	// Сохраняем ссылку в хранилище in-memory
	rm.urls[urlEntity.OriginalURL] = *urlEntity
	slog.Info("URL added to memory repository", slog.String("originalURL", originalURL), slog.String("shortKey", urlEntity.ShortKey))

	return shortURL.ToString(), nil
}
