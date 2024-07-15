package http

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/aggregate"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

type ShortenerService interface {
	Put(url string) (string, error)
	Get(url string) (*aggregate.URL, error)
}

type ShortenerHandler struct {
	shortenerService ShortenerService
}

func NewShortenerHandler(ss ShortenerService) ShortenerHandler {
	return ShortenerHandler{
		shortenerService: ss,
	}
}

func (sh ShortenerHandler) GetByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	url, err := sh.shortenerService.Get(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Location", url.OriginalURL().ToString())
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (sh ShortenerHandler) PostHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	var shortURL string

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = r.Body.Close(); err != nil {
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
