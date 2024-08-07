package entity

import (
	"github.com/google/uuid"
)

type URL struct {
	ID          string `json:"uuid"`
	ShortKey    string `json:"shortURL"`
	OriginalURL string `json:"originalURL"`
}

func NewURL(originalURL string, shortKey string) *URL {
	id := uuid.New()

	return &URL{
		ID:          id.String(),
		ShortKey:    shortKey,
		OriginalURL: originalURL,
	}
}
