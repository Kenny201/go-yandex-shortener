package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
)

var (
	ErrOpenFile   = errors.New("failed open or create file")
	ErrDecodeFile = errors.New("failed decode file")
	ErrEncodeFile = errors.New("failed encode file")
)

type RepositoryFile struct {
	fileName string
	urls     map[string]*entity.URL
}

func NewRepositoryFile(fileName string) *RepositoryFile {
	return &RepositoryFile{
		fileName: fileName,
		urls:     make(map[string]*entity.URL),
	}
}

func (rf *RepositoryFile) Get(shortKey string) (*entity.URL, error) {
	url, ok := rf.urls[shortKey]

	if !ok {
		err := fmt.Errorf("url %v not found", shortKey)
		return nil, err
	}

	return url, nil
}

func (rf *RepositoryFile) GetAll() (map[string]*entity.URL, error) {
	return rf.urls, nil
}

// Put Записать значение в файл
func (rf *RepositoryFile) Put(originalURL string, baseURL valueobject.BaseURL) (string, error) {
	file, err := os.OpenFile(rf.fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOpenFile, err.Error())
	}

	encoder := json.NewEncoder(file)

	defer func() {
		err := file.Close()

		if err != nil {
			slog.Error(
				"failed close file",
				slog.String("fileName", rf.fileName),
				slog.String("error", err.Error()),
			)
		}
	}()

	// Если запись существует повторную запись не производим
	if value, ok := rf.checkExistsOriginalURL(originalURL); ok {
		return fmt.Sprintf("%s/%s", baseURL.ToString(), value.ShortKey), nil
	}

	shortURL := valueobject.NewShortURL(baseURL)
	urlEntity := entity.NewURL(originalURL, shortURL.ShortKey())
	err = encoder.Encode(&urlEntity)

	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrEncodeFile, err.Error())
	}

	rf.urls[shortURL.ShortKey()] = urlEntity

	return shortURL.ToString(), nil
}

// ReadAll Прочитать строки из файла и декодировать в entity.URL
// Добавляет декодированные элементы в map[string]*entity.URL
func (rf *RepositoryFile) ReadAll() error {
	var url entity.URL

	file, err := os.OpenFile(rf.fileName, os.O_RDONLY|os.O_CREATE, 0666)

	if err != nil {
		return fmt.Errorf("%w: %s", ErrOpenFile, err.Error())
	}

	defer func() {
		err := file.Close()

		if err != nil {
			slog.Error(
				"failed close file",
				slog.String("fileName", rf.fileName),
				slog.String("error", err.Error()),
			)
		}
	}()

	decoder := json.NewDecoder(file)

	for {
		err := decoder.Decode(&url)

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("%w: %s", ErrDecodeFile, err.Error())
		}

		rf.urls[url.ShortKey] = &url
	}

	return nil
}

// Проверка существования записи в файле
func (rf *RepositoryFile) checkExistsOriginalURL(originalURL string) (*entity.URL, bool) {
	for _, value := range rf.urls {
		if value.OriginalURL == originalURL {
			return value, true
		}
	}

	return nil, false
}
