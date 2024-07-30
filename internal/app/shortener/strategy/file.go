package strategy

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

type File struct {
	baseURL    string
	repository *storage.RepositoryFile
}

func NewFile(baseURL string, fileName string) Strategy {
	file := &File{}

	file.repository = storage.NewRepositoryFile(fileName)
	err := file.repository.ReadAll()

	if err != nil {
		panic(err)
	}

	file.baseURL = baseURL

	return file
}

func (file *File) Get(shortKey string) (*entity.URL, error) {
	return file.repository.Get(shortKey)
}

func (file *File) GetAll() (map[string]*entity.URL, error) {
	return file.repository.GetAll()
}

func (file *File) Put(originalURL string) (string, error) {
	baseURL, err := valueobject.NewBaseURL(file.baseURL)

	if err != nil {
		return "", err
	}

	// Сохраняем ссылку в хранилище и получаем обратно
	shortURL, _ := file.repository.Put(originalURL, baseURL)
	return shortURL, nil
}
