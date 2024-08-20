package entity

import (
	"github.com/google/uuid"
)

type URLItem struct {
	ID          string `json:"correlation_id"`
	ShortURL    string `json:"short_url"`
	ShortKey    string `json:"short_key,omitempty"`
	OriginalURL string `json:"original_url,omitempty"`
}

type URL struct {
	ID          string `json:"uuid"`
	ShortKey    string `json:"short_key"`
	OriginalURL string `json:"original_url"`
}

func NewURL(originalURL string, shortKey string) *URL {
	id := uuid.New()

	return &URL{
		ID:          id.String(),
		ShortKey:    shortKey,
		OriginalURL: originalURL,
	}
}
