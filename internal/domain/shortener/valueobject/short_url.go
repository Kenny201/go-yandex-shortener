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
	baseURL  BaseURL
	shortKey string
}

func NewShortURL(baseURL BaseURL) ShortURL {
	shortKey := generateShortKey()

	return ShortURL{baseURL, shortKey}
}

// ToString Преобразовать в строку формата: url/shortKey
func (su ShortURL) ToString() string {
	return fmt.Sprintf("%s/%s", su.baseURL.ToString(), su.shortKey)
}

// ShortKey Получить сокращённую ссылку
func (su ShortURL) ShortKey() string {
	return su.shortKey
}

// Сгенерировать ключ, который будет добавлен к сокращённой ссылке формата url/shortKey
func generateShortKey() string {
	b := make([]byte, lengthShortURL)

	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)
}
