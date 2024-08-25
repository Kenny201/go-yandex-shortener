package storage

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
)

type Repository interface {
	GetShortenerRepository() shortener.ShortenerRepository
}
