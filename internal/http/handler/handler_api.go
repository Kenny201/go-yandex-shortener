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

func (sh Handler) PostAPI(w http.ResponseWriter, r *http.Request) {
	var (
		shortURL string
		url      Request
	)

	body, err := io.ReadAll(r.Body)

	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, NotReadRequestBody, err.Error())
		return
	}

	if err = json.Unmarshal(body, &url); err != nil {
		ErrorJSON(w, http.StatusBadRequest, NotUnmarshall, err.Error())
		return
	}

	if url.Url == "" {
		ErrorJSON(w, http.StatusBadRequest, BadRequest, ErrUrlIsEmpty.Error())
		return
	}

	shortURL, err = sh.shortenerService.Put(url.Url)

	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, BadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, Response{
		Result: shortURL,
	})
}
