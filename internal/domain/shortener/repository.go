package shortener

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/aggregate"
)

type Repository interface {
	Get(id string) (*aggregate.URL, error)
	GetAll() map[string]*aggregate.URL
	Put(url *aggregate.URL) (*aggregate.URL, error)
	CheckExistsBaseURL(baseURL string) (*aggregate.URL, bool)
}
