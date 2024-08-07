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

type File struct {
	baseURL  string
	filePath string
	urls     map[string]*entity.URL
}

func NewFile(baseURL, filePath string) (*File, error) {
	file := &File{
		filePath: filePath,
		urls:     make(map[string]*entity.URL),
		baseURL:  baseURL,
	}

	err := file.ReadAll()

	if err != nil {
		return nil, err
	}

	return file, nil
}

func (file *File) Get(shortKey string) (*entity.URL, error) {
	url, ok := file.urls[shortKey]

	if !ok {
		err := fmt.Errorf("url %v not found", shortKey)
		return nil, err
	}

	return url, nil
}

func (file *File) GetAll() map[string]*entity.URL {
	return file.urls
}

func (file *File) Put(originalURL string) (string, error) {
	baseURL, err := valueobject.NewBaseURL(file.baseURL)

	if err != nil {
		return "", err
	}

	f, err := os.OpenFile(file.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOpenFile, err.Error())
	}

	defer file.close(f)

	encoder := json.NewEncoder(f)

	// Если запись существует повторную запись не производим
	if value, ok := file.checkExistsOriginalURL(originalURL); ok {
		return fmt.Sprintf("%s/%s", baseURL.ToString(), value.ShortKey), nil
	}

	shortURL := valueobject.NewShortURL(baseURL)
	urlEntity := entity.NewURL(originalURL, shortURL.ShortKey())
	err = encoder.Encode(&urlEntity)

	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrEncodeFile, err.Error())
	}

	// Сохраняем ссылку в хранилище in-memory
	file.urls[urlEntity.ShortKey] = urlEntity

	return shortURL.ToString(), nil
}

// ReadAll Прочитать строки из файла и декодировать в entity.URL
// Добавляет декодированные элементы в repositoryMemory
func (file *File) ReadAll() error {
	var url *entity.URL

	err := file.makeDir()

	if err != nil {
		return err
	}

	f, err := os.OpenFile(file.filePath, os.O_RDONLY|os.O_CREATE, 0666)

	if err != nil {
		return fmt.Errorf("%w: %s", ErrOpenFile, err.Error())
	}

	defer file.close(f)

	decoder := json.NewDecoder(f)

	for {
		err := decoder.Decode(&url)

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("%w: %s", ErrDecodeFile, err.Error())
		}

		file.urls[url.ShortKey] = url
	}

	return nil
}

// Проверка существования записи в файле
func (file *File) checkExistsOriginalURL(originalURL string) (*entity.URL, bool) {
	for _, value := range file.GetAll() {
		if value.OriginalURL == originalURL {
			return value, true
		}
	}

	return nil, false
}

func (file *File) makeDir() error {
	folder := path.Dir(file.filePath)

	if _, err := os.Stat(folder); os.IsNotExist(err) {
		err := os.MkdirAll(folder, 0755)

		if err != nil {
			return ErrCreateDir
		}
	}

	return nil
}

func (file *File) close(f *os.File) {
	err := f.Close()

	if err != nil {
		slog.Error(
			"failed close file",
			slog.String("fileName", file.filePath),
			slog.String("error", err.Error()),
		)
	}
}
