package http

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/url/entity"
	"io"
	"net/http"
)

type UrlService interface {
	PutURL(url string, r *http.Request) (string, error)
	GetURL(url string) (*entity.URL, error)
}

type UrlHandler struct {
	urlService UrlService
}

func NewUrlHandler(us UrlService) UrlHandler {
	return UrlHandler{
		urlService: us,
	}
}

func (uh UrlHandler) GetByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	url, err := uh.urlService.GetURL(id)

	if err != nil {
		http.Error(w, "resource not found", http.StatusBadRequest)

		return
	}

	w.Header().Set("Location", url.OriginalURL())
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (uh UrlHandler) PostHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		http.Error(w, "request body is empty", http.StatusBadRequest)
		return
	}

	shortURL, _ := uh.urlService.PutURL(string(body), r)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}
