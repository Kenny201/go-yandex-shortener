package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Kenny201/go-yandex-shortener.git/internal/http/respond"
	"github.com/Kenny201/go-yandex-shortener.git/internal/utils/message"
)

type (
	Request struct {
		Url string
	}

	Response struct {
		Result string `json:"result"`
	}
)

func (sh Handler) PostAPI(w http.ResponseWriter, r *http.Request) {
	var (
		shortURL string
		url      Request
	)

	body, err := io.ReadAll(r.Body)

	if err != nil {
		respond.Error(w, http.StatusBadRequest, message.ErrBadRequest)
		return
	}

	if err = json.Unmarshal(body, &url); err != nil {
		respond.Error(w, http.StatusBadRequest, message.ErrBadRequest)
		return
	}

	if url.Url == "" {
		respond.Error(w, http.StatusBadRequest, message.ErrEmptyURL)
		return
	}

	shortURL, err = sh.shortenerService.Put(url.Url)

	if err != nil {
		respond.Error(w, http.StatusBadRequest, message.ErrBadRequest)
		return
	}

	respond.Json(w, http.StatusCreated, Response{
		Result: shortURL,
	})
}
