package urlgenerator

import (
	"fmt"
	"github.com/Kenny201/go-yandex-shortener.git/internal/storage"
	"math/rand"
	"net/http"
)

const (
	lengthShortUrl = 5
	letterBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func GenerateShortKey() string {
	b := make([]byte, lengthShortUrl)

	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)
}

func GetShortUrl(inputUrl string, r *http.Request) string {
	var body string

	urlStorage := *storage.GetStorage()

	if len(urlStorage) != 0 {
		if key, ok := storage.CheckExistsValueIntoUrlStorage(inputUrl); ok {
			body = fmt.Sprintf("http://%v/%s", r.Host, key)
		} else {
			shortUrl := GenerateShortKey()
			urlStorage[shortUrl] = inputUrl
			body = fmt.Sprintf("http://%v/%s", r.Host, shortUrl)
		}
	} else {
		shortUrl := GenerateShortKey()
		urlStorage[shortUrl] = inputUrl
		body = fmt.Sprintf("http://%v/%s", r.Host, shortUrl)
	}

	return body
}
