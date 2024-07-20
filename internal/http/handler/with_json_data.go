package handler

import (
	"encoding/json"
	"io"
	"net/http"
)

type (
	Request struct {
		Url string
	}

	Response struct {
		Result string `json:"result"`
	}
)

func (sh Handler) PostWithDataJSON(w http.ResponseWriter, r *http.Request) {
	var (
		shortURL     string
		url          Request
		responseData []byte
	)

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(body, &url); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if url.Url == "" {
		http.Error(w, "url is empty", http.StatusBadRequest)
		return
	}

	shortURL, err = sh.shortenerService.Put(url.Url)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	responseData, err = json.Marshal(Response{
		Result: shortURL,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(responseData))
}
