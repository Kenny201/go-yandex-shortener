package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

type (
	ErrorResponse struct {
		Code   int         `json:"code"`
		Error  string      `json:"error"`
		Detail interface{} `json:"detail,omitempty"`
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
		ErrorJSONResponse(w, http.StatusBadRequest, FailedReadRequestBody, err.Error())
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, FailedUnmarshall, err.Error())
		return
	}

	if request.URL == "" {
		ErrorJSONResponse(w, http.StatusBadRequest, BadRequest, ErrURLIsEmpty.Error())
		return
	}

	shortURL, err = sh.shortenerService.CreateShortURL(request.URL)

	if err != nil {
		if errors.Is(err, storage.ErrorUrlAlreadyExist) {
			ErrorJSONResponse(w, http.StatusConflict, BadRequest, shortURL)
			return
		}

		ErrorJSONResponse(w, http.StatusBadRequest, BadRequest, err.Error())
		return
	}

	JSONResponse(w, http.StatusCreated, Response{
		Result: shortURL,
	})
}

func (sh Handler) PostBatch(w http.ResponseWriter, r *http.Request) {
	var requestBatch []*entity.URLItem

	body, err := io.ReadAll(r.Body)

	if err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, FailedReadRequestBody, err.Error())
		return
	}

	if err = json.Unmarshal(body, &requestBatch); err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, FailedUnmarshall, err.Error())
		return
	}

	urls, err := sh.shortenerService.CreateListShortURL(requestBatch)

	if err != nil {
		if errors.Is(err, storage.ErrorUrlAlreadyExist) {
			ErrorJSONResponse(w, http.StatusConflict, BadRequest, urls)
			return
		}

		ErrorJSONResponse(w, http.StatusBadRequest, BadRequest, err.Error())
		return
	}

	JSONResponse(w, http.StatusCreated, urls)
}

func ErrorJSONResponse(w http.ResponseWriter, code int, error string, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Code: code, Error: error, Detail: payload})
}

func JSONResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if payload == nil {
		return
	}

	data, err := json.Marshal(payload)

	if err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, FailedMarshall, err.Error())
		return
	}

	w.Write(data)
}
