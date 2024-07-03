package handlers

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/storage"
	"github.com/Kenny201/go-yandex-shortener.git/internal/urlgenerator"
	"io"
	"net/http"
)

var urlStorage = *storage.GetStorage()

func GetByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if _, ok := urlStorage[id]; !ok {
		http.Error(w, "resource not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", urlStorage[id])
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		http.Error(w, "request body is empty", http.StatusBadRequest)
		return
	}

	response := urlgenerator.GetShortURL(string(body), r)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(response))
}
