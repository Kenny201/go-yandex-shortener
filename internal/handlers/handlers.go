package handlers

import (
	"fmt"
	"net/http"
	"strings"
)

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
			handleGetShort(w, r, id)
		default:
			http.Error(w, "Method not allowed", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Not support URL", http.StatusBadRequest)
	}
}

func handleGetShort(w http.ResponseWriter, r *http.Request, id string) {
	w.Write([]byte(id))
}

func handlePostShort(w http.ResponseWriter, r *http.Request) {
	body := fmt.Sprintf("Method: %s\r\n", r.Method)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	body += fmt.Sprintf("http://%v: \r\n", r.Host)

	if err := r.ParseForm(); err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	for k, v := range r.Form {
		body += fmt.Sprintf("%s: %v\r\n", k, v)
	}

	w.Write([]byte(body))
}
