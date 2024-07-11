package shortener

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/aggregate"
)

type Repository interface {
	Get(id string) (*aggregate.URL, error)
	GetAll() []aggregate.URL
	Put(url *aggregate.URL) (*aggregate.URL, error)
	CheckExistsOriginalURL(baseURL string) (*aggregate.URL, bool)
}
