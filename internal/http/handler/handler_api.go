package handler

import (
	"encoding/json"
	"io"
	"net/http"
)

type (
	ErrorResponse struct {
		Code   int    `json:"code"`
		Error  string `json:"error"`
		Detail string `json:"detail,omitempty"`
	}

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
		request  Request
	)

	body, err := io.ReadAll(r.Body)

	if err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, NotReadRequestBody, err.Error())
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, NotUnmarshall, err.Error())
		return
	}

	if request.URL == "" {
		ErrorJSONResponse(w, http.StatusBadRequest, BadRequest, ErrURLIsEmpty.Error())
		return
	}

	shortURL, err = sh.shortenerService.Put(request.URL)

	if err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, BadRequest, err.Error())
		return
	}

	JSONResponse(w, http.StatusCreated, Response{
		Result: shortURL,
	})
}

func ErrorJSONResponse(w http.ResponseWriter, code int, error string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Code: code, Error: error, Detail: message})
}

func JSONResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if payload == nil {
		return
	}

	data, err := json.Marshal(payload)

	if err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, NotMarshall, err.Error())
		return
	}

	w.Write(data)
}
