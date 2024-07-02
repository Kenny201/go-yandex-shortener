package handlers

import (
	"fmt"
	"github.com/Kenny201/go-yandex-shortener.git/internal/storage"
	"github.com/Kenny201/go-yandex-shortener.git/internal/urlgenerator"
	"io"
	"net/http"
	"strings"
)

var urlStorage = *storage.GetStorage()

func Handler(w http.ResponseWriter, r *http.Request) {
	firstSegmentURL := strings.Split(r.URL.Path[1:], "/")[0]
	id := fmt.Sprintf("/%s", firstSegmentURL)

	switch r.URL.Path {
	case "/":
		switch r.Method {
		case "POST":
			postHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusBadRequest)
		}
	case id:
		switch r.Method {
		case "GET":
			getByIDHandler(w, firstSegmentURL)
		default:
			http.Error(w, "Method not allowed", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Not support URL", http.StatusBadRequest)
	}
}

func getByIDHandler(w http.ResponseWriter, id string) {
	if _, ok := urlStorage[id]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", urlStorage[id])
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
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
