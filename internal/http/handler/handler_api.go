package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http/middleware"
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, FailedReadRequestBody, err.Error())
		return
	}

	if err := json.Unmarshal(body, &request); err != nil {
		respondWithError(w, http.StatusBadRequest, FailedUnmarshall, err.Error())
		return
	}

	if request.URL == "" {
		respondWithError(w, http.StatusBadRequest, BadRequest, ErrURLIsEmpty.Error())
		return
	}

	shortURL, err := h.shortenerService.CreateShortURL(r.Context(), request.URL)

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

	body, err := io.ReadAll(r.Body)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, FailedReadRequestBody, err.Error())
		return
	}

	if err := json.Unmarshal(body, &requestBatch); err != nil {
		respondWithError(w, http.StatusBadRequest, FailedUnmarshall, err.Error())
		return
	}

	urls, err := h.shortenerService.CreateListShortURL(r.Context(), requestBatch)

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

func (h Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(string)

	slog.Info("Fetching URLs for user", slog.String("userID", userID))

	urls, err := h.shortenerService.GetAllShortURL(userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserListURL) {
			slog.Info("No URLs found for user", slog.String("userID", userID))
			respondWithError(w, http.StatusNoContent, "", "")
			return
		}

		slog.Error("Error fetching URLs for user", slog.String("userID", userID), slog.String("error", err.Error()))
		respondWithError(w, http.StatusBadRequest, "", err.Error())
		return
	}

	slog.Info("Successfully fetched URLs for user", slog.String("userID", userID))
	respondWithJSON(w, http.StatusOK, urls)
}

func (h Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(string)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	var shortKeys []string

	if err := json.Unmarshal(body, &shortKeys); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	err = h.shortenerService.Delete(shortKeys, userID)

	if err != nil {
		slog.Error("Failed to delete URLs", slog.String("userID", userID), slog.String("error", err.Error()))
	}

	slog.Info("Successfully deleted URLs for user", slog.String("userID", userID))
	respondWithJSON(w, http.StatusAccepted, nil)
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
		// Здесь не нужно вызывать respondWithError снова, так как это приведет к повторному вызову WriteHeader.
		http.Error(w, fmt.Sprintf(`{"error": "failed to marshal response", "details": "%s"}`, err.Error()), http.StatusInternalServerError)
	}
}
