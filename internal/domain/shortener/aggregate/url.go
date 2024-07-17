package aggregate

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
)

type URL struct {
	id          string
	originalURL string
	shortURL    valueobject.ShortURL
}

func NewURL(originalURL string, shortURL valueobject.ShortURL) *URL {
	return &URL{
		id:          shortURL.ShortString(),
		originalURL: originalURL,
		shortURL:    shortURL,
	}
}

// Получить ID
func (u *URL) ID() string {
	return u.id
}

// Получить URL на который будет редирект
func (u *URL) OriginalURL() string {
	return u.originalURL
}

// Получить короткую ссылку
func (u *URL) ShortURL() valueobject.ShortURL {
	return u.shortURL
}
