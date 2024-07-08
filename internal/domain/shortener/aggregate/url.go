package aggregate

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/valueobject"
)

type URL struct {
	id       string
	baseURL  valueobject.BaseURL
	shortURL valueobject.ShortURL
}

func NewURL(baseURL valueobject.BaseURL, shortURL valueobject.ShortURL) *URL {
	return &URL{
		id:       shortURL.ShortString(),
		baseURL:  baseURL,
		shortURL: shortURL,
	}
}

func (u *URL) ID() string {
	return u.id
}

func (u *URL) BaseURL() valueobject.BaseURL {
	return u.baseURL
}

func (u *URL) ShortURL() valueobject.ShortURL {
	return u.shortURL
}
