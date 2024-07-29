package strategy

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

type File struct {
	baseURL        string
	repositoryFile *storage.RepositoryFile
}

func NewFile(baseURL string, fileName string) Strategy {
	f := &File{}

	f.repositoryFile = storage.NewRepositoryFile(fileName)
	err := f.repositoryFile.ReadAll()

	if err != nil {
		panic(err)
	}

	f.baseURL = baseURL

	return f
}

func (f *File) Get(shortKey string) (*entity.URL, error) {
	return f.repositoryFile.Get(shortKey)
}

func (f *File) GetAll() (map[string]*entity.URL, error) {
	return f.repositoryFile.GetAll()
}

func (f *File) Put(originalURL string) (string, error) {
	baseURL, err := valueobject.NewBaseURL(f.baseURL)

	if err != nil {
		return "", err
	}

	// Сохраняем ссылку в хранилище и получаем обратно
	shortURL, _ := f.repositoryFile.Put(originalURL, baseURL)
	return shortURL, nil
}
