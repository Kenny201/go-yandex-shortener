package handler

import (
	"encoding/json"
	"io"
	"net/http"
)

type (
	Request struct {
		URL string
	}

	Response struct {
		Result string `json:"result"`
	}
)

func (sh Handler) PostAPI(w http.ResponseWriter, r *http.Request) {
	var (
		shortURL string
		URL      Request
	)

	body, err := io.ReadAll(r.Body)

	if err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, NotReadRequestBody, err.Error())
		return
	}

	if err = json.Unmarshal(body, &URL); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, NotUnmarshall, err.Error())
		return
	}

	if URL.URL == "" {
		ErrorJSONResponse(w, http.StatusBadRequest, BadRequest, ErrURLIsEmpty.Error())
		return
	}

	shortURL, err = sh.shortenerService.Put(URL.URL)

	if err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, BadRequest, err.Error())
		return
	}

	JSONResponse(w, http.StatusCreated, Response{
		Result: shortURL,
	})
}
