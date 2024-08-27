package repository

import (
	"fmt"
	"log/slog"
	"runtime"
	"sync"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
)

type ShortenerMemory struct {
	baseURL string
	urls    map[string]entity.URL
	mu      sync.Mutex
}

// NewShortenerMemory создает новый репозиторий сокращения ссылок в памяти.
func NewShortenerMemory(baseURL string) *ShortenerMemory {
	return &ShortenerMemory{
		baseURL: baseURL,
		urls:    make(map[string]entity.URL),
	}
}

// Get возвращает URL-адрес по короткому ключу, если он существует.
func (mr *ShortenerMemory) Get(shortKey string) (*entity.URL, error) {
	for _, v := range mr.urls {
		if v.ShortKey == shortKey {
			slog.Info("URL retrieved successfully", slog.String("shortKey", shortKey))
			return &v, nil
		}
	}
	return nil, fmt.Errorf("URL %v not found", shortKey)
}

// Create добавляет новый URL в репозиторий, если его еще нет.
func (mr *ShortenerMemory) Create(url *entity.URL) (*entity.URL, error) {
	if v, exists := mr.urls[url.OriginalURL]; exists {
		return &v, ErrURLAlreadyExist
	}

	mr.urls[url.OriginalURL] = *url
	return url, nil
}

// CreateList добавляет список новых URL в репозиторий, возвращая их сокращенные версии.
func (mr *ShortenerMemory) CreateList(userID interface{}, urls []*entity.URLItem) ([]*entity.URLItem, error) {
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
			UserID:      userID,
			ShortKey:    shortURL.ShortKey(),
			OriginalURL: urlItem.OriginalURL,
		}

		shortUrls = append(shortUrls, &entity.URLItem{ID: urlEntity.ID, ShortURL: fmt.Sprintf("%s/%s", mr.baseURL, urlEntity.ShortKey)})
		mr.urls[urlItem.OriginalURL] = urlEntity
	}

	slog.Info("All URLs created successfully", slog.Int("count", len(shortUrls)))
	return shortUrls, nil
}

// GetAll получает все ссылки определённого пользователя
func (mr *ShortenerMemory) GetAll(userID string) ([]*entity.URLItem, error) {
	shortUrls := make([]*entity.URLItem, 0, len(mr.urls))

	for _, urlItem := range mr.urls {
		if urlItem.UserID == userID {
			shortUrls = append(shortUrls, &entity.URLItem{ShortURL: fmt.Sprintf("%s/%s", mr.baseURL, urlItem.ShortKey), OriginalURL: urlItem.OriginalURL})
		}
	}

	// Если ссылки не найдены
	if len(shortUrls) == 0 {
		return nil, fmt.Errorf("%w:%s", ErrUserListURL, userID)
	}

	return shortUrls, nil
}

// MarkAsDeleted устанавливает поле IsDeleted в true для списка URL по коротким ключам.
func (mr *ShortenerMemory) MarkAsDeleted(userID string, shortKeys []string) error {
	if len(shortKeys) == 0 {
		return fmt.Errorf("empty shortKey list provided")
	}

	const batchSize = 10           // Размер батча для обновлений
	numBatches := runtime.NumCPU() // Количество воркеров

	// Создание группы ошибок и канала для передачи батчей URL
	batchChan := make(chan []string, numBatches)

	// Запуск воркеров с использованием errgroup
	for i := 0; i < numBatches; i++ {
		go mr.processBatchUpdates(userID, batchChan)
	}

	// Наполнение batchChan и закрытие канала
	go func() {
		for i := 0; i < len(shortKeys); i += batchSize {
			end := i + batchSize
			if end > len(shortKeys) {
				end = len(shortKeys)
			}
			batchChan <- shortKeys[i:end]
		}
		close(batchChan)
	}()

	return nil
}

// processBatchUpdates обрабатывает обновления URL в батчах.
func (mr *ShortenerMemory) processBatchUpdates(userID string, batchChan <-chan []string) {
	for batch := range batchChan {
		mr.updateUrls(userID, batch)
	}
}

// updateUrls обновление urls
func (mr *ShortenerMemory) updateUrls(userID string, shortKeys []string) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	// Обновление записей по ключам.
	for _, shortKey := range shortKeys {
		for originalURL, url := range mr.urls {
			if url.ShortKey == shortKey && !url.DeletedFlag && url.UserID == userID {
				url.DeletedFlag = true
				mr.urls[originalURL] = url
				break
			}
		}
	}
}

// CheckHealth проверяет состояние репозитория, возвращая ошибку, если он не инициализирован.
func (mr *ShortenerMemory) CheckHealth() error {
	if mr.urls == nil {
		return fmt.Errorf("memory URLs structure is not initialized")
	}
	slog.Info("Memory repository health check passed")
	return nil
}
