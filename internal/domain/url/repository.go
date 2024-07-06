package url

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/url/entity"
)

type UrlRepository interface {
	GetURL(id string) (*entity.URL, error)
	GetAllURL() []entity.URL
	PutURL(url *entity.URL) (*entity.URL, error)
	CheckExistsOriginalURL(shortValue string) (*entity.URL, bool)
}
