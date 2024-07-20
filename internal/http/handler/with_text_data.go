package handler

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (sh Handler) GetWithTextData(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	url, err := sh.shortenerService.Get(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Location", url.OriginalURL())
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (sh Handler) PostWithTextData(w http.ResponseWriter, r *http.Request) {
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
