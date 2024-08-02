package strategy

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
)

type Repository interface {
	Get(id string) (*entity.URL, error)
	GetAll() map[string]*entity.URL
	Put(url *entity.URL)
}

type Strategy interface {
	Put(originalURL string) (string, error)
	Get(shortKey string) (*entity.URL, error)
	GetAll() map[string]*entity.URL
	checkExistsOriginalURL(originalURL string) (*entity.URL, bool)
}
