package urlgenerator

import (
	"fmt"
	"github.com/Kenny201/go-yandex-shortener.git/internal/storage"
	"math/rand"
	"net/http"
)

const (
	lengthShortURL = 5
	letterBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func generateShortKey() string {
	b := make([]byte, lengthShortURL)

	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)
}

func GetShortURL(inputURL string, r *http.Request) string {
	var body string

	urlStorage := *storage.GetStorage()

	if len(urlStorage) != 0 {
		if key, ok := storage.CheckExistsValueIntoURLStorage(inputURL); ok {
			body = fmt.Sprintf("http://%v/%s", r.Host, key)
		} else {
			shortURL := generateShortKey()
			urlStorage[shortURL] = inputURL
			body = fmt.Sprintf("http://%v/%s", r.Host, shortURL)
		}
	} else {
		shortURL := generateShortKey()
		urlStorage[shortURL] = inputURL
		body = fmt.Sprintf("http://%v/%s", r.Host, shortURL)
	}

	return body
}
