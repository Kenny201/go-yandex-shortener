package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

// ErrorResponse представляет формат ответа об ошибке.
type ErrorResponse struct {
	Code   int         `json:"code"`
	Error  string      `json:"error"`
	Detail interface{} `json:"detail,omitempty"`
}

// Request представляет запрос на создание короткого URL.
type Request struct {
	URL string `json:"url"`
}

// Response представляет успешный ответ с результатом.
type Response struct {
	Result string `json:"result"`
}

// PostAPI обрабатывает POST-запрос для создания короткого URL.
// Ожидает JSON с полем URL и возвращает короткий URL или ошибку.
func (h Handler) PostAPI(w http.ResponseWriter, r *http.Request) {
	var request Request

	// Чтение тела запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, FailedReadRequestBody, err.Error())
		return
	}

	// Разбор тела запроса
	if err := json.Unmarshal(body, &request); err != nil {
		respondWithError(w, http.StatusBadRequest, FailedUnmarshall, err.Error())
		return
	}

	// Проверка наличия URL
	if request.URL == "" {
		respondWithError(w, http.StatusBadRequest, BadRequest, ErrURLIsEmpty.Error())
		return
	}

	// Создание короткого URL
	shortURL, err := h.shortenerService.CreateShortURL(request.URL)

	if err != nil {
		if errors.Is(err, storage.ErrURLAlreadyExist) {
			respondWithError(w, http.StatusConflict, BadRequest, shortURL)
			return
		}
		respondWithError(w, http.StatusBadRequest, BadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, Response{Result: shortURL})
}

// PostBatch обрабатывает POST-запрос для создания нескольких коротких URL.
// Ожидает массив JSON объектов с полем URL и возвращает массив созданных URL или ошибку.
func (h Handler) PostBatch(w http.ResponseWriter, r *http.Request) {
	var requestBatch []*entity.URLItem

	// Чтение тела запроса
	body, err := io.ReadAll(r.Body)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, FailedReadRequestBody, err.Error())
		return
	}

	// Разбор тела запроса
	if err := json.Unmarshal(body, &requestBatch); err != nil {
		respondWithError(w, http.StatusBadRequest, FailedUnmarshall, err.Error())
		return
	}

	// Создание списка коротких URL
	urls, err := h.shortenerService.CreateListShortURL(requestBatch)

	if err != nil {
		if errors.Is(err, storage.ErrURLAlreadyExist) {
			respondWithError(w, http.StatusConflict, BadRequest, urls)
			return
		}
		respondWithError(w, http.StatusBadRequest, BadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, urls)
}

// respondWithError отправляет ответ об ошибке в формате JSON.
func respondWithError(w http.ResponseWriter, code int, errorMessage string, detail interface{}) {
	respondWithJSON(w, code, ErrorResponse{Code: code, Error: errorMessage, Detail: detail})
}

// respondWithJSON отправляет успешный ответ в формате JSON.
func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if payload == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		respondWithError(w, http.StatusInternalServerError, FailedMarshall, err.Error())
	}
}
