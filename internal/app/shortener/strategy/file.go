package strategy

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
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

var (
	ErrOpenFile   = errors.New("failed open or create file")
	ErrDecodeFile = errors.New("failed decode file")
	ErrEncodeFile = errors.New("failed encode file")
	ErrCreateDir  = errors.New("failed create or open directory")
)

type File struct {
	baseURL    string
	filePath   string
	repository Repository
}

func NewFile(baseURL, filePath string) Strategy {
	file := &File{}
	file.filePath = filePath
	file.repository = storage.NewMemoryShortenerRepository()
	err := file.ReadAll()

	if err != nil {
		panic(err)
	}

	file.baseURL = baseURL

	return file
}

func (file *File) Get(shortKey string) (*entity.URL, error) {
	return file.repository.Get(shortKey)
}

func (file *File) GetAll() map[string]*entity.URL {
	return file.repository.GetAll()
}

func (file *File) Put(originalURL string) (string, error) {
	baseURL, err := valueobject.NewBaseURL(file.baseURL)

	if err != nil {
		return "", err
	}

	err = file.makeDir()

	if err != nil {
		return "", err
	}

	f, err := os.OpenFile(file.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOpenFile, err.Error())
	}

	encoder := json.NewEncoder(f)

	defer file.close(f)

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

	// Сохраняем ссылку в хранилище
	file.repository.Put(urlEntity)

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

		file.repository.Put(url)
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
