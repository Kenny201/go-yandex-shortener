package url

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/url"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/url/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra"
	"net/http"
)

type Storage func(us *Service)

type Service struct {
	Ur url.Repository
}

func NewService(storages ...Storage) *Service {
	us := &Service{}

	for _, storage := range storages {
		storage(us)
	}

	return us
}

func WithRepository(ur url.Repository) Storage {
	return func(us *Service) {
		us.Ur = ur
	}
}

func WithMemoryRepository() Storage {
	mr := infra.NewMemoryRepositories()
	return WithRepository(mr)
}

func (us *Service) Put(url string, r *http.Request) string {
	var body string

	if len(us.Ur.GetAll()) != 0 {
		if key, ok := us.Ur.CheckExistsOriginal(url); ok {
			body = key.FullShortURL()
		} else {
			urlEntity := entity.NewURL(url, r.Host)
			urlEntity = us.Ur.Put(urlEntity)
			body = urlEntity.FullShortURL()
		}
	} else {
		urlEntity := entity.NewURL(url, r.Host)
		urlEntity = us.Ur.Put(urlEntity)
		body = urlEntity.FullShortURL()
	}

	return body
}

func (us *Service) Get(url string) (*entity.URL, error) {
	result, err := us.Ur.Get(url)

	if err != nil {
		return nil, err
	}

	return result, nil
}
