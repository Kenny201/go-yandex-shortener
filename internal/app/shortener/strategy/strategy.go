package strategy

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
)

type Strategy interface {
	Put(originalURL string) (string, error)
	Get(shortKey string) (*entity.URL, error)
	GetAll() (map[string]*entity.URL, error)
}
