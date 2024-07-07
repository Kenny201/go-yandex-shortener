package http

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/url/entity"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

type URLService interface {
	Put(url string, r *http.Request) string
	Get(url string) (*entity.URL, error)
}

type URLHandler struct {
	urlService URLService
}

func NewURLHandler(us URLService) URLHandler {
	return URLHandler{
		urlService: us,
	}
}

func (uh URLHandler) GetByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	url, err := uh.urlService.Get(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	w.Header().Set("Location", url.OriginalURL())
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (uh URLHandler) PostHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		http.Error(w, "request body is empty", http.StatusBadRequest)
		return
	}

	shortURL := uh.urlService.Put(string(body), r)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}
