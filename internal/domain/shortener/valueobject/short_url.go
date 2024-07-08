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
	serverAddress ServerAddress
	shortString   string
}

func NewShortURL(originalURL ServerAddress) ShortURL {
	shortString := generateShortKey()

	return ShortURL{originalURL, shortString}
}

func (su ShortURL) ToString() string {
	return fmt.Sprintf("%s://%s:%d/%s", su.serverAddress.scheme, su.serverAddress.host, su.serverAddress.port, su.shortString)
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
