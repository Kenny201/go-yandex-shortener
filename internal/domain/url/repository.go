package url

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/url/entity"
)

type Repository interface {
	Get(id string) (*entity.URL, error)
	GetAll() []entity.URL
	Put(url *entity.URL) *entity.URL
	CheckExistsOriginal(shortValue string) (*entity.URL, bool)
}
