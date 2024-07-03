package entity

import (
	"fmt"
	"math/rand"
)

const (
	lengthShortURL = 5
	letterBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type URL struct {
	id           string
	originalURL  string
	fullShortURL string
}

func NewURL(originalURL string, host string) *URL {
	shortURL := generateShortKey()

	return &URL{
		id:           shortURL,
		originalURL:  originalURL,
		fullShortURL: fmt.Sprintf("http://%v/%s", host, shortURL),
	}
}

func (u URL) ID() string {
	return u.id
}

func (u URL) OriginalURL() string {
	return u.originalURL
}

func (u URL) FullShortURL() string {
	return u.fullShortURL
}

func generateShortKey() string {
	b := make([]byte, lengthShortURL)

	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)
}
