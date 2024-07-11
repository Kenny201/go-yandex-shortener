package aggregate

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
)

type URL struct {
	id          string
	originalURL valueobject.OriginalURL
	shortURL    valueobject.ShortURL
}

func NewURL(originalURL valueobject.OriginalURL, shortURL valueobject.ShortURL) *URL {
	return &URL{
		id:          shortURL.ShortString(),
		originalURL: originalURL,
		shortURL:    shortURL,
	}
}

func (u *URL) ID() string {
	return u.id
}

func (u *URL) OriginalURL() valueobject.OriginalURL {
	return u.originalURL
}

func (u *URL) ShortURL() valueobject.ShortURL {
	return u.shortURL
}
