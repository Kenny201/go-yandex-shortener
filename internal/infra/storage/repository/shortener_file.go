package repository

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
)

var (
	ErrOpenFile   = errors.New("failed to open or create file")
	ErrDecodeFile = errors.New("failed to decode file")
	ErrEncodeFile = errors.New("failed to encode file")
	ErrCreateDir  = errors.New("failed to create or open directory")
)

type ShortenerFile struct {
	baseURL  string
	filePath string
	urls     map[string]entity.URL
	mu       sync.Mutex
}

// NewShortenerFile создает новый репозиторий сокращения ссылок с сохранением данных в файл.
func NewShortenerFile(baseURL, filePath string) (*ShortenerFile, error) {
	repo := &ShortenerFile{
		baseURL:  baseURL,
		filePath: filePath,
		urls:     make(map[string]entity.URL),
	}

	// Чтение всех существующих URL-ов из файла при инициализации репозитория.
	if err := repo.readAll(); err != nil {
		return nil, err
	}

	return repo, nil
}

// Get возвращает URL по короткому ключу, если он существует в файле.
func (fr *ShortenerFile) Get(shortKey string) (*entity.URL, error) {
	for _, v := range fr.urls {
		if v.ShortKey == shortKey && v.DeletedFlag {
			return nil, ErrURLDeleted // URL помечен как удаленный
		}

		if v.ShortKey == shortKey && !v.DeletedFlag {
			slog.Info("URL retrieved successfully", slog.String("shortKey", shortKey))
			return &v, nil
		}
	}
	return nil, fmt.Errorf("URL %v not found", shortKey)
}

// Create добавляет новый URL в файл и возвращает его сокращенную версию.
func (fr *ShortenerFile) Create(url *entity.URL) (*entity.URL, error) {
	if existingURL := fr.findExistingURL(url.OriginalURL); existingURL != nil {
		return existingURL, ErrURLAlreadyExist
	}

	fr.urls[url.OriginalURL] = *url
	err := fr.saveURLToFile(*url)
	if err != nil {
		return nil, err
	}

	return url, nil
}

// CreateList добавляет список новых URL в файл и возвращает их сокращенные версии.
func (fr *ShortenerFile) CreateList(userID interface{}, urls []*entity.URLItem) ([]*entity.URLItem, error) {
	shortUrls := make([]*entity.URLItem, 0, len(urls))

	baseURL, err := valueobject.NewBaseURL(fr.baseURL)

	if err != nil {
		return nil, err
	}

	for _, urlItem := range urls {
		shortURL := valueobject.NewShortURL(baseURL)

		if existingURL := fr.findExistingURL(urlItem.OriginalURL); existingURL != nil {
			return []*entity.URLItem{{ID: urlItem.ID, ShortURL: fmt.Sprintf("%s/%s", fr.baseURL, existingURL.ShortKey)}}, ErrURLAlreadyExist
		}

		urlEntity := entity.URL{
			ID:          urlItem.ID,
			UserID:      userID,
			ShortKey:    shortURL.ShortKey(),
			OriginalURL: urlItem.OriginalURL,
		}

		if err := fr.saveURLToFile(urlEntity); err != nil {
			return nil, err
		}
		shortUrls = append(shortUrls, &entity.URLItem{ID: urlEntity.ID, ShortURL: fmt.Sprintf("%s/%s", fr.baseURL, urlEntity.ShortKey)})
		fr.urls[urlItem.OriginalURL] = urlEntity
	}

	slog.Info("All URLs created successfully", slog.Int("count", len(shortUrls)))
	return shortUrls, nil
}

// GetAll получает все ссылки определённого пользователя
func (fr *ShortenerFile) GetAll(userID string) ([]*entity.URLItem, error) {
	shortUrls := make([]*entity.URLItem, 0, len(fr.urls))

	for _, urlItem := range fr.urls {
		if urlItem.UserID == userID {
			shortUrls = append(shortUrls, &entity.URLItem{ShortURL: fmt.Sprintf("%s/%s", fr.baseURL, urlItem.ShortKey), OriginalURL: urlItem.OriginalURL})
		}
	}

	// Если ссылки не найдены
	if len(shortUrls) == 0 {
		return nil, fmt.Errorf("%w:%s", ErrUserListURL, userID)
	}

	return shortUrls, nil
}

// MarkAsDeleted обновляет поле IsDeleted в true для списка URL по коротким ключам.
func (fr *ShortenerFile) MarkAsDeleted(shortKeys []string, userId string) error {
	if len(shortKeys) == 0 {
		return fmt.Errorf("empty URL list provided")
	}

	const batchSize = 10           // Размер батча для обновлений
	numBatches := runtime.NumCPU() // Количество воркеров

	// Создание группы ошибок и канала для передачи батчей URL
	eg := new(errgroup.Group)
	batchChan := make(chan []string, numBatches)

	// Запуск воркеров с использованием errgroup
	for i := 0; i < numBatches; i++ {
		eg.Go(func() error {
			return fr.processBatchUpdates(userId, batchChan)
		})
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

	// Ожидание завершения всех воркеров и обработки ошибок
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("one or more errors occurred: %v", err)
	}

	return nil
}

// processBatchUpdates обрабатывает обновления URL в батчах.
func (fr *ShortenerFile) processBatchUpdates(userID string, batchChan <-chan []string) error {
	for batch := range batchChan {
		if err := fr.updateFile(userID, batch); err != nil {
			return err // Возвращаем ошибку, чтобы она была обработана errgroup
		}
	}
	return nil
}

// updateFile обновляет файл, установив поле IsDeleted в true для заданных коротких ключей.
func (fr *ShortenerFile) updateFile(userID string, shortKeys []string) error {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	// Создание временного файла
	tmpFilePath := fr.filePath + ".tmp"
	tmpFile, err := os.Create(tmpFilePath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFilePath)
	}()

	file, err := os.Open(fr.filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(tmpFile)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if containsKey(line, shortKeys) && containsUserID(line, userID) {
			line = strings.ReplaceAll(line, `"is_deleted":false`, `"is_deleted":true`)
		}
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("error writing to temp file: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing temp file: %w", err)
	}

	if err := os.Rename(tmpFilePath, fr.filePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// containsKey проверяет, содержит ли строка любой из ключей.
func containsKey(line string, shortKeys []string) bool {
	for _, v := range shortKeys {
		if strings.Contains(line, v) {
			return true
		}
	}
	return false
}

// containsUserID проверяет, принадлежит ли строка указанному userID.
func containsUserID(line, userID string) bool {
	return strings.Contains(line, fmt.Sprintf(`"user_id":"%s"`, userID))
}

// findOrCreateURL ищет существующий URL в файле или создает новый, если не найден.
func (fr *ShortenerFile) findExistingURL(originalURL string) *entity.URL {
	if url, exists := fr.urls[originalURL]; exists {
		return &url
	}

	return nil
}

// saveURLToFile сохраняет новый URL в файл в формате JSON.
func (fr *ShortenerFile) saveURLToFile(url entity.URL) error {
	if err := fr.makeDir(); err != nil {
		return err
	}

	f, err := os.OpenFile(fr.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer fr.closeFile(f)

	if err != nil {
		return fmt.Errorf("%w: %v", ErrOpenFile, err)
	}

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(url); err != nil {
		return fmt.Errorf("%w: %v", ErrEncodeFile, err)
	}

	return nil
}

// readAll читает все URL из файла и загружает их в память.
func (fr *ShortenerFile) readAll() error {
	if err := fr.makeDir(); err != nil {
		return err
	}

	f, err := os.OpenFile(fr.filePath, os.O_RDONLY|os.O_CREATE, 0666)
	defer fr.closeFile(f)

	if err != nil {
		return fmt.Errorf("%w: %v", ErrOpenFile, err)
	}

	decoder := json.NewDecoder(f)

	for {
		var url entity.URL
		if err := decoder.Decode(&url); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("%w: %v", ErrDecodeFile, err)
		}
		fr.urls[url.OriginalURL] = url
	}

	slog.Info("All URLs loaded successfully from file", slog.Int("count", len(fr.urls)))
	return nil
}

// makeDir создает директорию для хранения файла, если она не существует.
func (fr *ShortenerFile) makeDir() error {
	folder := path.Dir(fr.filePath)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		if err := os.MkdirAll(folder, 0755); err != nil {
			return fmt.Errorf("%w: %v", ErrCreateDir, err)
		}
		slog.Info("Directory created for file storage", slog.String("folder", folder))
	}
	return nil
}

// closeFile закрывает файл и логгирует ошибку, если она произошла.
func (fr *ShortenerFile) closeFile(f *os.File) {
	if err := f.Close(); err != nil {
		slog.Error("Failed to close file", slog.String("filePath", fr.filePath), slog.String("error", err.Error()))
	}
}

// CheckHealth проверяет состояние репозитория, проверяя наличие файла на диске.
func (fr *ShortenerFile) CheckHealth() error {
	if _, err := os.Stat(fr.filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %w", err)
	}
	slog.Info("File repository health check passed", slog.String("filePath", fr.filePath))
	return nil
}
