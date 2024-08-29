package shortener

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"

	"golang.org/x/sync/errgroup"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http/middleware"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

// Repository определяет интерфейс для работы с хранилищем сокращённых ссылок.
// Реализации этого интерфейса могут быть на базе различных хранилищ данных (память, файл, база данных и т.д.).
type Repository interface {
	// Get возвращает URL-объект по его идентификатору (короткому ключу).
	Get(shortKey string) (*entity.URL, error)
	// Create создает новый короткий URL и возвращает его.
	Create(url *entity.URL) (*entity.URL, error)
	// CreateList создает несколько коротких URL и возвращает список созданных элементов.
	CreateList(userID interface{}, urls []*entity.URLItem) ([]*entity.URLItem, error)
	// GetAll получает все сокращённые ссылки пользователя
	GetAll(userID string) ([]*entity.URLItem, error)
	// MarkAsDeleted помечает определённые ссылки как удалённые
	MarkAsDeleted(batch []string, userID string) error
	// CheckHealth проверяет состояние хранилища (доступность, целостность и т.д.).
	CheckHealth() error
}

// Shortener представляет собой основной сервис для работы с сокращёнными ссылками.
// Он использует репозиторий для сохранения и получения данных.
type Shortener struct {
	repo    Repository
	baseURL string
}

// New создает новый экземпляр сервиса Shortener с заданным репозиторием.
func New(repository Repository, baseURL string) *Shortener {
	return &Shortener{repo: repository, baseURL: baseURL}
}

// GetShortURL возвращает сокращённую ссылку по короткому ключу или ошибку, если ссылка не найдена.
func (s *Shortener) GetShortURL(shortKey string) (*entity.URL, error) {
	return s.repo.Get(shortKey)
}

// CreateShortURL сохраняет оригинальный URL в хранилище и возвращает сокращённую ссылку.
// В случае ошибки возвращает пустую ссылку и ошибку.
func (s *Shortener) CreateShortURL(ctx context.Context, originalURL string) (string, error) {
	var userID interface{}
	var ok bool

	userID, ok = ctx.Value(middleware.UserIDContextKey).(string)

	baseURL, err := valueobject.NewBaseURL(s.baseURL)
	if err != nil {
		return "", err
	}

	shortURL := valueobject.NewShortURL(baseURL)

	if !ok || userID == "" {
		userID = nil // Передаем NULL
	}

	urlEntity := entity.NewURL(userID, originalURL, shortURL.ShortKey())
	shortURLStr := fmt.Sprintf("%s/%s", baseURL.ToString(), urlEntity.ShortKey)
	url, err := s.repo.Create(urlEntity)

	if err != nil {
		if errors.Is(err, storage.ErrURLAlreadyExist) {
			return fmt.Sprintf("%s/%s", baseURL.ToString(), url.ShortKey), storage.ErrURLAlreadyExist
		}

		return "", err
	}

	slog.Info("URL created", slog.String("originalURL", originalURL), slog.String("shortURL", shortURLStr))
	return shortURLStr, nil
}

// CreateListShortURL сохраняет список оригинальных URL в хранилище и возвращает список сокращённых ссылок.
// В случае ошибки возвращает список частично созданных ссылок и ошибку.
func (s *Shortener) CreateListShortURL(ctx context.Context, urls []*entity.URLItem) ([]*entity.URLItem, error) {

	if len(urls) == 0 {
		return nil, storage.ErrEmptyURL
	}

	userID, ok := ctx.Value(middleware.UserIDContextKey).(string)

	var userIDValue interface{}

	if !ok || userID == "" {
		userIDValue = nil // Передаем NULL
	} else {
		userIDValue = userID
	}

	savedURLs, err := s.repo.CreateList(userIDValue, urls)

	if err != nil {
		if errors.Is(err, storage.ErrURLAlreadyExist) {
			return savedURLs, storage.ErrURLAlreadyExist
		}

		return nil, err
	}

	slog.Info("Batch URL creation successful", slog.Int("count", len(savedURLs)))
	return savedURLs, nil
}

// GetAllShortURL сохраняет список оригинальных URL в хранилище и возвращает список сокращённых ссылок.
// В случае ошибки возвращает список частично созданных ссылок и ошибку.
func (s *Shortener) GetAllShortURL(userID string) ([]*entity.URLItem, error) {
	return s.repo.GetAll(userID)
}

func (s *Shortener) Delete(shortKeys []string, userID string) error {
	if len(shortKeys) == 0 {
		return fmt.Errorf("empty URL list provided")
	}

	const batchSize = 10                // Размер батча для обновлений
	numBatches := runtime.GOMAXPROCS(0) // Количество воркеров

	// Создаем каналы и группу ошибок
	eg := new(errgroup.Group)
	batchChan := make(chan []string, numBatches)
	doneChan := make(chan struct{})
	defer close(doneChan)

	// Запускаем воркеры
	for i := 0; i < numBatches; i++ {
		workerID := i // Локальная переменная для избежания захвата
		eg.Go(func() error {
			slog.Info("Worker started", slog.Int("workerID", workerID))
			err := s.processBatchUpdates(userID, batchChan, doneChan, workerID)
			if err != nil {
				slog.Error("Worker error", slog.Int("workerID", workerID), slog.String("error", err.Error()))
			}
			slog.Info("Worker finished", slog.Int("workerID", workerID))
			return err
		})
	}

	// Наполняем batchChan
	go func() {
		defer close(batchChan)
		for i := 0; i < len(shortKeys); i += batchSize {
			end := i + batchSize
			if end > len(shortKeys) {
				end = len(shortKeys)
			}
			select {
			case batchChan <- shortKeys[i:end]:
			case <-doneChan:
				return
			}
		}
	}()

	if err := eg.Wait(); err != nil {
		slog.Error("Error occurred during batch processing", slog.String("error", err.Error()))
		return fmt.Errorf("one or more errors occurred: %w", err)
	}

	slog.Info("All batches processed successfully")
	return nil
}

// processBatchUpdates обрабатывает обновления URL в батчах.
func (s *Shortener) processBatchUpdates(userID string, batchChan <-chan []string, doneChan <-chan struct{}, workerID int) error {
	for {
		select {
		case batch, ok := <-batchChan:
			if !ok {
				slog.Info("Worker received all batches and exiting", slog.Int("workerID", workerID))
				return nil
			}
			slog.Info("Worker processing batch", slog.Int("workerID", workerID), slog.Int("batchSize", len(batch)))

			if err := s.repo.MarkAsDeleted(batch, userID); err != nil {
				slog.Error("Failed to mark batch as deleted", slog.Int("workerID", workerID), slog.String("error", err.Error()))
				return err
			}

		case <-doneChan:
			slog.Info("Worker received done signal", slog.Int("workerID", workerID))
			return nil
		}
	}
}

// CheckHealth проверяет состояние репозитория, с которым работает сервис.
func (s *Shortener) CheckHealth() error {
	return s.repo.CheckHealth()
}
