package url

import (
	"fmt"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/url"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/url/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra"
	"net/http"
)

type UrlConfiguration func(us *UrlService) error

type UrlService struct {
	ur url.UrlRepository
}

func NewUrlService(cfgs ...UrlConfiguration) (*UrlService, error) {
	us := &UrlService{}

	for _, cfg := range cfgs {
		err := cfg(us)
		if err != nil {
			return nil, err
		}
	}

	return us, nil
}

func WithUrlRepository(ur url.UrlRepository) UrlConfiguration {
	return func(us *UrlService) error {
		us.ur = ur
		return nil
	}
}

func WithMemoryUrlRepository() UrlConfiguration {
	mr := infra.NewMemoryRepositories()
	return WithUrlRepository(mr)
}

func (us *UrlService) PutURL(url string, r *http.Request) (string, error) {
	var body string

	if len(us.ur.GetAllURL()) != 0 {
		if key, ok := us.ur.CheckExistsOriginalURL(url); ok {
			body = key.FullShortURL()

			fmt.Print(us.ur.GetAllURL())
		} else {
			urlEntity := entity.NewUrl(url, r.Host)
			urlEntity, _ = us.ur.PutURL(urlEntity)
			body = urlEntity.FullShortURL()

			fmt.Print(us.ur.GetAllURL())
		}
	} else {
		urlEntity := entity.NewUrl(url, r.Host)
		urlEntity, _ = us.ur.PutURL(urlEntity)
		body = urlEntity.FullShortURL()

		fmt.Print(us.ur.GetAllURL())
	}

	return body, nil
}

func (us *UrlService) GetURL(url string) (*entity.URL, error) {
	result, err := us.ur.GetURL(url)

	if err != nil {
		return nil, err
	}

	return result, nil
}
