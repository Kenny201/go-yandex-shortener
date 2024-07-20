package handler

import "github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/aggregate"

type (
	ShortenerService interface {
		Put(url string) (string, error)
		Get(url string) (*aggregate.URL, error)
	}

	Handler struct {
		shortenerService ShortenerService
	}
)

func New(ss ShortenerService) Handler {
	return Handler{
		shortenerService: ss,
	}
}
