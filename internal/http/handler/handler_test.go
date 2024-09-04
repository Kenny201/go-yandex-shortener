package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/spf13/viper"

	"github.com/Kenny201/go-yandex-shortener.git/cmd/shortener/config"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/dto"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http/middleware"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
	"github.com/Kenny201/go-yandex-shortener.git/internal/mocks"
)

// setupTestEnvironment инициализирует окружение для теста.
func setupTestEnvironment(t *testing.T) (*mocks.MockRepository, *gomock.Controller, *shortener.Shortener) {

	args := initArgs(t)

	ctrl := gomock.NewController(t)
	mockRepository := mocks.NewMockRepository(ctrl)

	shortenerService := shortener.New(mockRepository, args.BaseURL)

	return mockRepository, ctrl, shortenerService
}

// TestPostHandler тестирует обработчик создания короткого URL.
func TestPostHandler(t *testing.T) {
	tests := []struct {
		name                    string
		body                    string
		wantResponseContentType string
		wantStatusCode          int
	}{
		{
			name:                    "post_request_for_body_https://yandex.ru",
			body:                    "https://yandex.ru",
			wantResponseContentType: "text/plain",
			wantStatusCode:          http.StatusCreated,
		},
		{
			name:                    "post_request_for_body_https://practicum.yandex.ru",
			body:                    "https://practicum.yandex.ru",
			wantResponseContentType: "text/plain",
			wantStatusCode:          http.StatusCreated,
		},
		{
			name:                    "post_request_for_empty_body",
			body:                    "",
			wantResponseContentType: "text/plain; charset=utf-8",
			wantStatusCode:          http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepository, ctrl, shortenerService := setupTestEnvironment(t)
			defer ctrl.Finish()

			if tt.wantStatusCode == http.StatusCreated {
				mockRepository.EXPECT().Create(gomock.Any()).Return(nil, nil)
			}

			rw, req := sendRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			New(shortenerService, nil).Post(rw, req)

			response := rw.Result()
			defer responseClose(t, response)

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("expected status: got %v, want %v", response.StatusCode, tt.wantStatusCode)
			}

			if contentType := response.Header.Get("Content-Type"); contentType != tt.wantResponseContentType {
				t.Errorf("response content type header does not match: got %v, want %v", contentType, tt.wantResponseContentType)
			}
		})
	}
}

// TestGetHandler тестирует обработчик получения оригинального URL по короткому ключу.
func TestGetHandler(t *testing.T) {
	args := initArgs(t)

	tests := []struct {
		name               string
		id                 string
		wantLocationHeader string
		wantStatusCode     int
	}{
		{
			name:               "redirect_for_existing_short_url_yandex",
			id:                 "some-short-url",
			wantLocationHeader: "https://yandex.ru",
			wantStatusCode:     http.StatusTemporaryRedirect,
		},
		{
			name:               "redirect_for_existing_short_url_practicum",
			id:                 "some-short-url-practicum",
			wantLocationHeader: "https://practicum.yandex.ru",
			wantStatusCode:     http.StatusTemporaryRedirect,
		},
		{
			name:           "id_not_found",
			id:             "nonexistent-id",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "id_deleted",
			id:             "deleted-id",
			wantStatusCode: http.StatusGone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepository, ctrl, shortenerService := setupTestEnvironment(t)
			defer ctrl.Finish()

			if tt.wantStatusCode == http.StatusGone {
				mockRepository.EXPECT().Get(tt.id).Return(nil, storage.ErrURLDeleted)
			} else if tt.wantStatusCode != http.StatusNotFound {
				mockRepository.EXPECT().Get(tt.id).Return(&entity.URL{
					ID:          "some-id",
					ShortKey:    tt.id,
					OriginalURL: tt.wantLocationHeader,
				}, nil)
			} else {
				mockRepository.EXPECT().Get(tt.id).Return(nil, fmt.Errorf("not found"))
			}

			// Полный URL с коротким ключом
			rw, req := sendRequest(http.MethodGet, fmt.Sprintf("%s/%s", args.BaseURL, tt.id), nil)
			req = withURLParam(req, "id", tt.id)
			New(shortenerService, nil).Get(rw, req)

			response := rw.Result()
			defer responseClose(t, response)

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("expected status: got %v, want %v", response.StatusCode, tt.wantStatusCode)
			}

			if response.StatusCode == http.StatusTemporaryRedirect && response.Header.Get("Location") != tt.wantLocationHeader {
				t.Errorf("response Location header does not match: got %v, want %v",
					response.Header.Get("Location"), tt.wantLocationHeader)
			}
		})
	}
}

// TestPostAPIHandler тестирует обработчик создания короткого URL через JSON API.
func TestPostAPIHandler(t *testing.T) {
	args := initArgs(t)

	tests := []struct {
		name                    string
		body                    string
		wantStatusCode          int
		wantResponseContentType string
	}{
		{
			name:                    "post_json_request_for_body_https://yandex.ru",
			body:                    `{"url": "https://yandex.ru"}`,
			wantStatusCode:          http.StatusCreated,
			wantResponseContentType: "application/json",
		},
		{
			name:                    "post_json_request_for_body_https://practicum.yandex.ru",
			body:                    `{"url": "https://practicum.yandex.ru"}`,
			wantStatusCode:          http.StatusCreated,
			wantResponseContentType: "application/json",
		},
		{
			name:                    "post_json_request_for_empty_body",
			body:                    "",
			wantStatusCode:          http.StatusBadRequest,
			wantResponseContentType: "application/json",
		},
		{
			name:                    "post_request_when_body_isn't_json_type",
			body:                    "https://practicum.yandex.ru",
			wantStatusCode:          http.StatusBadRequest,
			wantResponseContentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepository, ctrl, shortenerService := setupTestEnvironment(t)
			defer ctrl.Finish()

			if tt.body != "" && tt.wantStatusCode == http.StatusCreated {
				mockRepository.EXPECT().Create(gomock.Any()).Return(nil, nil)
			}

			rw, req := sendRequest(http.MethodPost, args.BaseURL, strings.NewReader(tt.body))
			New(shortenerService, nil).PostAPI(rw, req)

			response := rw.Result()
			defer responseClose(t, response)

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("expected status: got %v, want %v", response.StatusCode, tt.wantStatusCode)
			}

			if contentType := response.Header.Get("Content-Type"); contentType != tt.wantResponseContentType {
				t.Errorf("response content type header does not match: got %v, want %v", contentType, tt.wantResponseContentType)
			}
		})
	}
}

// TestPingHandler тестирует обработчик проверки состояния сервиса.
func TestPingHandler(t *testing.T) {
	args := initArgs(t)

	tests := []struct {
		name           string
		wantStatusCode int
	}{
		{
			name:           "test_ping_handler",
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepository, ctrl, shortenerService := setupTestEnvironment(t)
			defer ctrl.Finish()

			mockRepository.EXPECT().CheckHealth().Return(nil)

			rw, req := sendRequest(http.MethodGet, args.BaseURL, nil)
			New(shortenerService, nil).Ping(rw, req)

			response := rw.Result()
			defer responseClose(t, response)

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("expected status: got %v, want %v", response.StatusCode, tt.wantStatusCode)
			}
		})
	}
}

// TestPostBatchHandler тестирует обработчик создания нескольких коротких URL.
func TestPostBatchHandler(t *testing.T) {
	tests := []struct {
		name                    string
		body                    string
		mockReturnValue         []*entity.URLItem
		mockReturnError         error
		wantStatusCode          int
		wantResponseContentType string
		expectCreateListCalled  bool
	}{
		{
			name: "post_batch_with_valid_urls",
			body: `[{"url": "https://yandex.ru"}, {"url": "https://practicum.yandex.ru"}]`,
			mockReturnValue: []*entity.URLItem{
				{ID: "1", ShortURL: "some-short-url-1"},
				{ID: "2", ShortURL: "some-short-url-2"},
			},
			mockReturnError:         nil,
			wantStatusCode:          http.StatusCreated,
			wantResponseContentType: "application/json",
			expectCreateListCalled:  true,
		},
		{
			name:                    "post_batch_with_empty_body",
			body:                    "",
			mockReturnValue:         nil,
			mockReturnError:         ErrURLIsEmpty,
			wantStatusCode:          http.StatusBadRequest,
			wantResponseContentType: "application/json",
			expectCreateListCalled:  false,
		},
		{
			name: "post_batch_with_conflicting_urls",
			body: `[{"url": "https://yandex.ru"}]`,
			mockReturnValue: []*entity.URLItem{
				{ID: "1", ShortURL: "some-short-url-1"},
			},
			mockReturnError:         storage.ErrURLAlreadyExist,
			wantStatusCode:          http.StatusConflict,
			wantResponseContentType: "application/json",
			expectCreateListCalled:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepository, ctrl, shortenerService := setupTestEnvironment(t)
			defer ctrl.Finish()

			if tt.expectCreateListCalled {
				mockRepository.EXPECT().CreateList(nil, gomock.Any()).Return(tt.mockReturnValue, tt.mockReturnError)
			} else {
				mockRepository.EXPECT().CreateList(nil, gomock.Any()).Times(0)
			}

			rw, req := sendRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(tt.body))
			New(shortenerService, nil).PostBatch(rw, req)

			response := rw.Result()
			defer responseClose(t, response)

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("expected status: got %v, want %v", response.StatusCode, tt.wantStatusCode)
			}

			if contentType := response.Header.Get("Content-Type"); contentType != tt.wantResponseContentType {
				t.Errorf("response content type header does not match: got %v, want %v", contentType, tt.wantResponseContentType)
			}
		})
	}
}

// TestGetAllHandler тестирует обработчик получения всех сокращенных URL пользователя.
func TestGetAllHandler(t *testing.T) {
	tests := []struct {
		name                    string
		userID                  string
		mockReturnValue         []*entity.URLItem
		mockReturnError         error
		wantStatusCode          int
		wantResponseContentType string
	}{
		{
			name:   "get_all_with_valid_userID",
			userID: "user123",
			mockReturnValue: []*entity.URLItem{
				{ID: "1", ShortURL: "https://short.url/1", OriginalURL: "https://original.url/1"},
				{ID: "2", ShortURL: "https://short.url/2", OriginalURL: "https://original.url/2"},
			},
			mockReturnError:         nil,
			wantStatusCode:          http.StatusOK,
			wantResponseContentType: "application/json",
		},
		{
			name:                    "get_all_with_no_urls",
			userID:                  "user123",
			mockReturnValue:         []*entity.URLItem{},
			mockReturnError:         storage.ErrUserListURL,
			wantStatusCode:          http.StatusNoContent,
			wantResponseContentType: "application/json",
		},
		{
			name:                    "get_all_with_service_error",
			userID:                  "user123",
			mockReturnValue:         nil,
			mockReturnError:         fmt.Errorf("service error"),
			wantStatusCode:          http.StatusBadRequest,
			wantResponseContentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepository, ctrl, shortenerService := setupTestEnvironment(t)
			defer ctrl.Finish()

			// Настройка ожидания вызова метода GetAllShortURL в зависимости от условий теста.
			if tt.userID != "" {
				mockRepository.EXPECT().GetAll(tt.userID).Return(tt.mockReturnValue, tt.mockReturnError)
			}

			rw, req := sendRequest(http.MethodGet, "/api/user/urls", nil)
			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDContextKey, tt.userID))

			New(shortenerService, nil).GetAll(rw, req)

			response := rw.Result()
			defer responseClose(t, response)

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("expected status: got %v, want %v", response.StatusCode, tt.wantStatusCode)
			}

			if response.Header.Get("Content-Type") != tt.wantResponseContentType {
				t.Errorf("response content type header does not match: got %v, want %v", response.Header.Get("Content-Type"), tt.wantResponseContentType)
			}

			if tt.wantStatusCode == http.StatusOK {
				body, err := io.ReadAll(response.Body)
				if err != nil {
					t.Fatalf("failed to read response body: %v", err)
				}

				var urls []*entity.URLItem
				if err := json.Unmarshal(body, &urls); err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}

				if len(urls) != len(tt.mockReturnValue) {
					t.Errorf("expected number of URLs: got %v, want %v", len(urls), len(tt.mockReturnValue))
				}
			}
		})
	}
}

// TestDeleteHandler тестирует обработчик удаления короткого URL.
func TestHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		userID         string
		expectedStatus int
		expectTask     *dto.DeleteTask
		expectError    bool
	}{
		{
			name:           "valid_request",
			body:           `["short-key-1", "short-key-2"]`,
			userID:         "user123",
			expectedStatus: http.StatusAccepted,
			expectTask: &dto.DeleteTask{
				ShortKeys: []string{"short-key-1", "short-key-2"},
				UserID:    "user123",
			},
			expectError: false,
		},
		{
			name:           "invalid_json",
			body:           `invalid-json`,
			userID:         "user123",
			expectedStatus: http.StatusBadRequest,
			expectTask:     nil,
			expectError:    true,
		},
		{
			name:           "empty_body",
			body:           ``,
			userID:         "user123",
			expectedStatus: http.StatusBadRequest,
			expectTask:     nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleteChannel := make(chan dto.DeleteTask, 1)

			h := Handler{deleteChannel: deleteChannel}

			body := strings.NewReader(tt.body)

			rw, req := sendRequest(http.MethodDelete, "/api/shorten", body)
			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDContextKey, tt.userID))

			h.Delete(rw, req)

			response := rw.Result()
			defer responseClose(t, response)

			if response.StatusCode != tt.expectedStatus {
				t.Errorf("expected status: got %v, want %v", response.StatusCode, tt.expectedStatus)
			}

			if !tt.expectError {
				select {
				case task := <-deleteChannel:
					if !reflect.DeepEqual(task, *tt.expectTask) {
						t.Errorf("expected task: got %v, want %v", task, *tt.expectTask)
					}
				default:
					t.Errorf("expected task to be sent to channel, but it was not")
				}
			}
		})
	}
}

func initArgs(t *testing.T) *config.Args {
	t.Helper()
	err := config.LoadConfig("../../../")

	if err != nil {
		t.Errorf("error read config %v", err)
	}

	serverAddress := fmt.Sprintf(":%s", viper.GetString("SERVER_ADDRESS"))
	baseURL := fmt.Sprintf("http://localhost:%s", viper.GetString("PORT"))
	fileStoragePath := "tmp/Rquxc"
	databaseDNS := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", viper.GetString("DB_USERNAME"), viper.GetString("DB_PASSWORD"), viper.GetString("DB_HOST"), viper.GetString("DB_PORT"), viper.GetString("DB_DATABASE"))

	args := []string{
		"-a", serverAddress,
		"-b", baseURL,
		"-f", fileStoragePath,
		"-d", databaseDNS,
	}

	a := config.NewArgs()
	a.ParseFlags(args)

	return a
}

func responseClose(t *testing.T, response *http.Response) {
	t.Helper()
	err := response.Body.Close()
	if err != nil {
		t.Errorf("failed to close response body: %v", err.Error())
	}
}

func sendRequest(method, url string, body io.Reader) (*httptest.ResponseRecorder, *http.Request) {
	req := httptest.NewRequest(method, url, body)
	return httptest.NewRecorder(), req
}

func withURLParam(r *http.Request, key, value string) *http.Request {
	chiCtx := chi.NewRouteContext()
	req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))
	chiCtx.URLParams.Add(key, value)
	return req
}
