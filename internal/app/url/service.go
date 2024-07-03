package url

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/url"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/url/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra"
	"net/http"
)

type Configuration func(us *Service) error

type Service struct {
	Ur url.Repository
}

func NewService(cfgs ...Configuration) (*Service, error) {
	us := &Service{}

	for _, cfg := range cfgs {
		err := cfg(us)
		if err != nil {
			return nil, err
		}
	}

	return us, nil
}

func WithRepository(ur url.Repository) Configuration {
	return func(us *Service) error {
		us.Ur = ur
		return nil
	}
}

func WithMemoryRepository() Configuration {
	mr := infra.NewMemoryRepositories()
	return WithRepository(mr)
}

func (us *Service) Put(url string, r *http.Request) (string, error) {
	var body string

	if len(us.Ur.GetAll()) != 0 {
		if key, ok := us.Ur.CheckExistsOriginal(url); ok {
			body = key.FullShortURL()
		} else {
			urlEntity := entity.NewURL(url, r.Host)
			urlEntity, _ = us.Ur.Put(urlEntity)
			body = urlEntity.FullShortURL()
		}
	} else {
		urlEntity := entity.NewURL(url, r.Host)
		urlEntity, _ = us.Ur.Put(urlEntity)
		body = urlEntity.FullShortURL()
	}

	return body, nil
}

func (us *Service) Get(url string) (*entity.URL, error) {
	result, err := us.Ur.Get(url)

	if err != nil {
		return nil, err
	}

	return result, nil
}
