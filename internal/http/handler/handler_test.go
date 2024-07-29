package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Kenny201/go-yandex-shortener.git/cmd/shortener/config"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener/strategy"
	"github.com/go-chi/chi/v5"
)

func TestPostHandler(t *testing.T) {
	tests := []struct {
		name                    string
		body                    string
		wantResponseContentType string
		wantStatusCode          int
	}{
		{
			name:                    "post request for body https://yandex.ru",
			body:                    "https://yandex.ru",
			wantResponseContentType: "text/plain",
			wantStatusCode:          http.StatusCreated,
		},
		{
			name:                    "post request for body https://practicum.yandex.ru",
			body:                    "https://practicum.yandex.ru",
			wantResponseContentType: "text/plain",
			wantStatusCode:          http.StatusCreated,
		},
		{
			name:                    "post request for empty body",
			body:                    "",
			wantResponseContentType: "text/plain; charset=utf-8",
			wantStatusCode:          http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := config.NewArgs()
			args.SetArgs(":8080", "http://localhost:8080", "urls.txt")

			req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			strg := strategy.NewMemory(args.BaseURL)
			ss := shortener.NewService()
			ss.SetStrategy(strg)

			New(ss).Post(w, req)

			res := w.Result()

			if res.StatusCode != tt.wantStatusCode {
				t.Errorf("excpected status %v; got %v", res.StatusCode, tt.wantStatusCode)
			}

			if ctype := res.Header.Get("Content-Type"); ctype != tt.wantResponseContentType {
				t.Errorf("response content type header does not match: got %v wantResponseContentType %v",
					ctype, tt.wantResponseContentType)
			}

			defer func() {
				err := res.Body.Close()
				if err != nil {
					t.Errorf("failed to close response body: %v", err.Error())
				}
			}()
		})
	}
}

func TestGetHandler(t *testing.T) {
	tests := []struct {
		name               string
		body               string
		id                 string
		wantLocationHeader string
		wantStatusCode     int
	}{
		{
			name:               "redirect for body https://yandex.ru",
			body:               "https://yandex.ru",
			wantLocationHeader: "https://yandex.ru",
			wantStatusCode:     http.StatusTemporaryRedirect,
		},
		{
			name:               "redirect for body https://practicum.yandex.ru",
			body:               "https://practicum.yandex.ru",
			wantLocationHeader: "https://practicum.yandex.ru",
			wantStatusCode:     http.StatusTemporaryRedirect,
		},
		{
			name:           "id not found",
			body:           "https://yandex.ru",
			id:             "sdsds",
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Установка аргументов командной строки
			args := config.NewArgs()
			args.SetArgs(":8080", "http://localhost:8080", "urls.txt")

			req := httptest.NewRequest(http.MethodPost, "http://localhost:8080", strings.NewReader(tt.body))
			responseForPost := httptest.NewRecorder()

			strg := strategy.NewMemory(args.BaseURL)
			ss := shortener.NewService()
			ss.SetStrategy(strg)

			handler := New(ss)
			handler.Post(responseForPost, req)

			urlStorage, _ := ss.GetAll()

			for _, v := range urlStorage {
				req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/", nil)

				chiCtx := chi.NewRouteContext()
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

				if tt.id != "" {
					chiCtx.URLParams.Add("id", tt.id)
				} else {
					chiCtx.URLParams.Add("id", v.ShortKey)
				}

				responseForGet := httptest.NewRecorder()
				handler.Get(responseForGet, req)
				res := responseForGet.Result()

				if res.StatusCode != tt.wantStatusCode {
					t.Errorf("excpected status: got %v want %v", res.StatusCode, tt.wantStatusCode)
				}

				if location := res.Header.Get("Location"); location != tt.wantLocationHeader {
					t.Errorf("location header does not match: got %v want %v",
						location, tt.wantLocationHeader)
				}

				err := res.Body.Close()

				if err != nil {
					t.Errorf("failed to close response body: %v", err.Error())
				}
			}
		})
	}
}

func TestPostAPIHandler(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		wantStatusCode int
	}{
		{
			name:           "post json request for body https://yandex.ru",
			body:           `{"url": "https://yandex.ru"}`,
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "post json request for body https://practicum.yandex.ru",
			body:           `{"url": "https://practicum.yandex.ru"}`,
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "post json request for empty body",
			body:           "",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "post request when body isn`t json type",
			body:           "https://practicum.yandex.ru",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := config.NewArgs()
			args.SetArgs(":8080", "http://localhost:8080", "urls.txt")

			req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			strg := strategy.NewFile(args.BaseURL, args.FileStoragePath)
			ss := shortener.NewService()
			ss.SetStrategy(strg)

			New(ss).PostAPI(w, req)

			res := w.Result()

			if res.StatusCode != tt.wantStatusCode {
				t.Errorf("excpected status %v; got %v", res.StatusCode, tt.wantStatusCode)
			}

			defer func() {
				err := res.Body.Close()
				if err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()
		})
	}
}
