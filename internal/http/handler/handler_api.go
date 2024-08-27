package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http/middleware"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage/repository"
	"io"
	"log/slog"
	"net/http"
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
	shortURL, err := h.shortenerService.CreateShortURL(r.Context(), request.URL)

	if err != nil {
		if errors.Is(err, repository.ErrURLAlreadyExist) {
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
	urls, err := h.shortenerService.CreateListShortURL(r.Context(), requestBatch)

	if err != nil {
		if errors.Is(err, repository.ErrURLAlreadyExist) {
			respondWithError(w, http.StatusConflict, BadRequest, urls)
			return
		}
		respondWithError(w, http.StatusBadRequest, BadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, urls)
}

func (h Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok || userID == "" {
		slog.Warn("Unauthorized access attempt: userID not found or empty")
		respondWithError(w, http.StatusUnauthorized, "", "")
		return
	}

	// Логируем userID для дальнейшего анализа
	slog.Info("Fetching URLs for user", slog.String("userID", userID))

	// Получаем список URL пользователя
	urls, err := h.shortenerService.GetAllShortURL(userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserListURL) {
			slog.Info("No URLs found for user", slog.String("userID", userID))
			respondWithError(w, http.StatusNoContent, "", "")
			return
		}

		slog.Error("Error fetching URLs for user", slog.String("userID", userID), slog.String("error", err.Error()))
		respondWithError(w, http.StatusBadRequest, "", err.Error())
		return
	}

	if len(urls) == 0 {
		slog.Info("No URLs found for user", slog.String("userID", userID))
		respondWithError(w, http.StatusNoContent, "", "")
		return
	}

	slog.Info("Successfully fetched URLs for user", slog.String("userID", userID))
	respondWithJSON(w, http.StatusOK, urls)
}

func (h Handler) Delete(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok || userID == "" {
		slog.Warn("Unauthorized access attempt: userID not found or empty")
		respondWithError(w, http.StatusUnauthorized, "", "")
		return
	}
	// Чтение всего тела запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	// Объявляем переменную для хранения данных в виде слайса строк
	var lines []string

	// Парсим JSON в слайс строк
	if err := json.Unmarshal(body, &lines); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}
	err = h.shortenerService.Delete(lines, userID)

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
