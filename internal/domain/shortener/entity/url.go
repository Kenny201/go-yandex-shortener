package entity

import (
	"github.com/google/uuid"
)

type URLItem struct {
	ID          string `json:"correlation_id,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	ShortURL    string `json:"short_url"`
	ShortKey    string `json:"short_key,omitempty"`
	OriginalURL string `json:"original_url,omitempty"`
}

type URL struct {
	ID          string      `json:"uuid"`
	UserID      interface{} `json:"user_id"`
	ShortKey    string      `json:"short_key"`
	OriginalURL string      `json:"original_url"`
}

func NewURL(userID interface{}, originalURL string, shortKey string) *URL {
	id := uuid.New()

	return &URL{
		ID:          id.String(),
		UserID:      userID,
		ShortKey:    shortKey,
		OriginalURL: originalURL,
	}
}
