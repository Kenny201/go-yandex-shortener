package valueobject

import (
	"crypto/rand"
	"fmt"
)

const (
	shortURLLength = 5
	letterBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type ShortURL struct {
	baseURL  BaseURL
	shortKey string
}

// NewShortURL создает новый объект ShortURL с заданной базовой URL и сгенерированным коротким ключом.
func NewShortURL(baseURL BaseURL) ShortURL {
	return ShortURL{
		baseURL:  baseURL,
		shortKey: generateShortKey(),
	}
}

// ToString возвращает строку в формате: url/shortKey.
func (su ShortURL) ToString() string {
	return fmt.Sprintf("%s/%s", su.baseURL.ToString(), su.shortKey)
}

// ShortKey возвращает сокращенный ключ.
func (su ShortURL) ShortKey() string {
	return su.shortKey
}

// generateShortKey генерирует короткий ключ для сокращенной ссылки.
// Использует криптографически стойкий генератор случайных чисел.
func generateShortKey() string {
	b := make([]byte, shortURLLength)

	_, err := rand.Read(b)

	if err != nil {
		// На случай ошибки генерации случайных чисел
		panic(fmt.Sprintf("failed to generate short key: %v", err))
	}

	// Преобразуем байты в строку, используя только разрешенные символы
	for i := range b {
		b[i] = letterBytes[int(b[i])%len(letterBytes)]
	}

	return string(b)
}
