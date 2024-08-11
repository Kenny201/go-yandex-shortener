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
	ErrOpenFile   = errors.New("failed open or create file")
	ErrDecodeFile = errors.New("failed decode file")
	ErrEncodeFile = errors.New("failed encode file")
	ErrCreateDir  = errors.New("failed create or open directory")
)

type FileShortenerRepository struct {
	baseURL  string
	filePath string
	urls     map[string]entity.URL
}

func NewFileShortenerRepository(baseURL, filePath string) (*FileShortenerRepository, error) {
	file := &FileShortenerRepository{
		filePath: filePath,
		urls:     make(map[string]entity.URL),
		baseURL:  baseURL,
	}

	err := file.ReadAll()

	if err != nil {
		return nil, err
	}

	return file, nil
}

func (file *FileShortenerRepository) Get(shortKey string) (*entity.URL, error) {
	for _, v := range file.urls {
		if v.ShortKey == shortKey {
			return &v, nil
		}
	}

	return nil, fmt.Errorf("url %v not found", shortKey)
}

func (file *FileShortenerRepository) Create(originalURL string) (string, error) {
	baseURL, err := valueobject.NewBaseURL(file.baseURL)

	if err != nil {
		return "", err
	}

	f, err := os.OpenFile(file.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer file.close(f)

	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOpenFile, err.Error())
	}

	encoder := json.NewEncoder(f)

	// Если запись существует повторную запись не производим
	if v, ok := file.urls[originalURL]; ok {
		return fmt.Sprintf("%s/%s", baseURL.ToString(), v.ShortKey), nil
	}

	shortURL := valueobject.NewShortURL(baseURL)
	urlEntity := entity.NewURL(originalURL, shortURL.ShortKey())
	err = encoder.Encode(&urlEntity)

	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrEncodeFile, err.Error())
	}

	// Сохраняем ссылку в хранилище in-memory
	file.urls[urlEntity.OriginalURL] = *urlEntity

	return shortURL.ToString(), nil
}

func (file *FileShortenerRepository) CreateList(urls []*entity.URLItem) ([]*entity.URLItem, error) {
	baseURL, err := valueobject.NewBaseURL(file.baseURL)
	shortUrls := make([]*entity.URLItem, 0, len(urls))

	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(file.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer file.close(f)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOpenFile, err.Error())
	}

	encoder := json.NewEncoder(f)

	for _, v := range urls {
		// Если запись существует повторную запись не производим
		if url, ok := file.urls[v.OriginalURL]; ok {
			shortUrls = append(
				shortUrls,
				&entity.URLItem{
					ID:       url.ID,
					ShortURL: fmt.Sprintf("%s/%s", baseURL.ToString(), url.ShortKey),
				},
			)
			continue
		}

		shortURL := valueobject.NewShortURL(baseURL)

		urlEntity := &entity.URL{
			ID:          v.ID,
			ShortKey:    shortURL.ShortKey(),
			OriginalURL: v.OriginalURL,
		}

		err = encoder.Encode(&urlEntity)

		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrEncodeFile, err.Error())
		}

		// Сохраняем ссылку в хранилище in-memory
		file.urls[urlEntity.OriginalURL] = *urlEntity

		shortUrls = append(
			shortUrls,
			&entity.URLItem{
				ID:       v.ID,
				ShortURL: fmt.Sprintf("%s/%s", baseURL.ToString(), shortURL.ShortKey()),
			},
		)
	}

	return shortUrls, nil
}

// ReadAll Прочитать строки из файла и декодировать в entity.URL
// Добавляет декодированные элементы в repositoryMemory
func (file *FileShortenerRepository) ReadAll() error {
	var url entity.URL

	err := file.makeDir()

	if err != nil {
		return err
	}

	f, err := os.OpenFile(file.filePath, os.O_RDONLY|os.O_CREATE, 0666)
	defer file.close(f)

	if err != nil {
		return fmt.Errorf("%w: %s", ErrOpenFile, err.Error())
	}

	decoder := json.NewDecoder(f)

	for {
		err := decoder.Decode(&url)

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("%w: %s", ErrDecodeFile, err.Error())
		}

		file.urls[url.OriginalURL] = url
	}

	return nil
}

func (file *FileShortenerRepository) makeDir() error {
	folder := path.Dir(file.filePath)

	if _, err := os.Stat(folder); os.IsNotExist(err) {
		err := os.MkdirAll(folder, 0755)

		if err != nil {
			return ErrCreateDir
		}
	}

	return nil
}

func (file *FileShortenerRepository) close(f *os.File) {
	err := f.Close()

	if err != nil {
		slog.Error(
			"failed close file",
			slog.String("fileName", file.filePath),
			slog.String("error", err.Error()),
		)
	}
}
