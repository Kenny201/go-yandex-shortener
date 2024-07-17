package valueobject

import (
	"fmt"
	"math/rand"
)

const (
	lengthShortURL = 5
	letterBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type ShortURL struct {
	baseURL     BaseURL
	shortString string
}

func NewShortURL(baseURL BaseURL) ShortURL {
	shortString := generateShortKey()

	return ShortURL{baseURL, shortString}
}

// ToString Преобразовать в строку формата: url/shortURL
func (su ShortURL) ToString() string {
	return fmt.Sprintf("%s/%s", su.baseURL.ToString(), su.shortString)
}

func (su ShortURL) ShortString() string {
	return su.shortString
}

func generateShortKey() string {
	b := make([]byte, lengthShortURL)

	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)
}
