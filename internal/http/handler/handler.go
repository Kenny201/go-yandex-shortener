package handler

import (
	"io"
	"net/http"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/aggregate"
	"github.com/go-chi/chi/v5"
)

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

func (sh Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	url, err := sh.shortenerService.Get(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Location", url.OriginalURL())
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (sh Handler) Post(w http.ResponseWriter, r *http.Request) {
	var shortURL string

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		http.Error(w, "request body is empty", http.StatusBadRequest)
		return
	}

	shortURL, err = sh.shortenerService.Put(string(body))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}
