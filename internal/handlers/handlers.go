package handlers

import (
	"fmt"
	"github.com/Kenny201/go-yandex-shortener.git/internal/storage"
	"github.com/Kenny201/go-yandex-shortener.git/internal/urlgenerator"
	"net/http"
	"strings"
)

var urlStorage = *storage.GetStorage()

func ShortHandler(w http.ResponseWriter, r *http.Request) {
	firstSegmentUrl := strings.Split(r.URL.Path[1:], "/")[0]
	id := fmt.Sprintf("/%s", firstSegmentUrl)

	switch r.URL.Path {
	case "/":
		switch r.Method {
		case "POST":
			handlePostShort(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusBadRequest)
		}
	case id:
		switch r.Method {
		case "GET":
			handleGetShort(w, r, firstSegmentUrl)
		default:
			http.Error(w, "Method not allowed", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Not support URL", http.StatusBadRequest)
	}
}

func handleGetShort(w http.ResponseWriter, r *http.Request, id string) {
	if _, ok := urlStorage[id]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, urlStorage[id], http.StatusTemporaryRedirect)
}

func handlePostShort(w http.ResponseWriter, r *http.Request) {
	inputUrl := r.FormValue("url")

	if err := r.ParseForm(); err != nil {
		http.Error(w, string([]byte(err.Error())), http.StatusBadRequest)
		return
	}

	if inputUrl == "" {
		http.Error(w, "Field url required", http.StatusBadRequest)
		return
	}

	body := urlgenerator.GetShortUrl(inputUrl, r)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(body))
}
