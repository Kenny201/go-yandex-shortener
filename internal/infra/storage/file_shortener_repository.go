package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
)

var (
	ErrOpenFile   = errors.New("failed to open or create file")
	ErrDecodeFile = errors.New("failed to decode file")
	ErrEncodeFile = errors.New("failed to encode file")
	ErrCreateDir  = errors.New("failed to create or open directory")
)

type FileShortenerRepository struct {
	baseURL  string
	filePath string
	urls     map[string]entity.URL
}

// NewFileShortenerRepository создает новый репозиторий сокращения ссылок с сохранением данных в файл.
func NewFileShortenerRepository(baseURL, filePath string) (*FileShortenerRepository, error) {
	repo := &FileShortenerRepository{
		baseURL:  baseURL,
		filePath: filePath,
		urls:     make(map[string]entity.URL),
	}

	// Чтение всех существующих URL-ов из файла при инициализации репозитория.
	if err := repo.ReadAll(); err != nil {
		return nil, err
	}

	return repo, nil
}

// Get возвращает URL по короткому ключу, если он существует в файле.
func (repo *FileShortenerRepository) Get(shortKey string) (*entity.URL, error) {
	for _, v := range repo.urls {
		if v.ShortKey == shortKey {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("URL %v not found", shortKey)
}

// Create добавляет новый URL в файл и возвращает его сокращенную версию.
func (repo *FileShortenerRepository) Create(originalURL string) (string, error) {
	if shortURL, err := repo.findOrCreateURL(originalURL); err != nil {
		if errors.Is(err, ErrURLAlreadyExist) {
			return shortURL, err
		}
		return "", err
	} else {
		return shortURL, nil
	}
}

// CreateList добавляет список новых URL в файл и возвращает их сокращенные версии.
func (repo *FileShortenerRepository) CreateList(urls []*entity.URLItem) ([]*entity.URLItem, error) {
	shortUrls := make([]*entity.URLItem, 0, len(urls))

	for _, urlItem := range urls {
		if shortURL, err := repo.findOrCreateURL(urlItem.OriginalURL); err != nil {
			if errors.Is(err, ErrURLAlreadyExist) {
				return []*entity.URLItem{{ID: urlItem.ID, ShortURL: shortURL}}, err
			}
			return nil, err
		} else {
			shortUrls = append(shortUrls, &entity.URLItem{ID: urlItem.ID, ShortURL: shortURL})
		}
	}

	return shortUrls, nil
}

// findOrCreateURL ищет существующий URL в файле или создает новый, если не найден.
func (repo *FileShortenerRepository) findOrCreateURL(originalURL string) (string, error) {
	baseURL, err := valueobject.NewBaseURL(repo.baseURL)

	if err != nil {
		return "", err
	}

	// Проверка существования записи в мапе urls.
	if url, exists := repo.urls[originalURL]; exists {
		return fmt.Sprintf("%s/%s", baseURL.ToString(), url.ShortKey), ErrURLAlreadyExist
	}

	shortURL := valueobject.NewShortURL(baseURL)
	urlEntity := entity.NewURL(originalURL, shortURL.ShortKey())

	// Сохранение нового URL в файл.
	if err := repo.saveURLToFile(urlEntity); err != nil {
		return "", err
	}

	repo.urls[originalURL] = *urlEntity
	return shortURL.ToString(), nil
}

// saveURLToFile сохраняет новый URL в файл в формате JSON.
func (repo *FileShortenerRepository) saveURLToFile(url *entity.URL) error {
	if err := repo.makeDir(); err != nil {
		return err
	}

	f, err := os.OpenFile(repo.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer repo.closeFile(f)

	if err != nil {
		return fmt.Errorf("%w: %v", ErrOpenFile, err)
	}

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(url); err != nil {
		return fmt.Errorf("%w: %v", ErrEncodeFile, err)
	}

	return nil
}

// ReadAll читает все URL из файла и загружает их в память.
func (repo *FileShortenerRepository) ReadAll() error {
	if err := repo.makeDir(); err != nil {
		return err
	}

	f, err := os.OpenFile(repo.filePath, os.O_RDONLY|os.O_CREATE, 0666)
	defer repo.closeFile(f)

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
		repo.urls[url.OriginalURL] = url
	}

	return nil
}

// makeDir создает директорию для хранения файла, если она не существует.
func (repo *FileShortenerRepository) makeDir() error {
	folder := path.Dir(repo.filePath)

	if _, err := os.Stat(folder); os.IsNotExist(err) {

		if err := os.MkdirAll(folder, 0755); err != nil {
			return fmt.Errorf("%w: %v", ErrCreateDir, err)
		}

	}
	return nil
}

// closeFile закрывает файл и логгирует ошибку, если она произошла.
func (repo *FileShortenerRepository) closeFile(f *os.File) {
	if err := f.Close(); err != nil {
		slog.Error("failed to close file", slog.String("filePath", repo.filePath), slog.String("error", err.Error()))
	}
}

// CheckHealth проверяет состояние репозитория, проверяя наличие файла на диске.
func (repo *FileShortenerRepository) CheckHealth() error {
	if _, err := os.Stat(repo.filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %w", err)
	}
	return nil
}
