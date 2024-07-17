package http

import (
	"context"
	"github.com/Kenny201/go-yandex-shortener.git/cmd/shortener/config"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
			args.SetArgs(":8080", "http://localhost:8080")

			req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/", strings.NewReader(tt.body))

			if err != nil {
				t.Fatalf("method not alowed: %v", err)
			}

			w := httptest.NewRecorder()

			ss := shortener.NewService(args.BaseURL, storage.NewRepositoryMemory())

			NewShortenerHandler(ss).PostHandler(w, req)

			res := w.Result()

			if res.StatusCode != tt.wantStatusCode {
				t.Errorf("excpected status %v; got %v", tt.wantStatusCode, res.StatusCode)
			}

			if ctype := res.Header.Get("Content-Type"); ctype != tt.wantResponseContentType {
				t.Errorf("response content type header does not match: got %v wantResponseContentType %v",
					ctype, tt.wantResponseContentType)
			}

			_, err = io.ReadAll(res.Body)

			defer func() {
				err := res.Body.Close()
				if err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()

			if err != nil {
				t.Fatalf("could not read response:%v", err)
			}
		})
	}
}

func TestGetByIDHandler(t *testing.T) {
	tests := []struct {
		name               string
		body               string
		wantLocationHeader string
		wantStatusCode     int
	}{
		{
			name:               "redirect for body https://yandex.ru",
			body:               "https://yandex.ru",
			wantLocationHeader: "https://yandex.ru",
			wantStatusCode:     http.StatusOK,
		},
		{
			name:               "redirect for body https://practicum.yandex.ru",
			body:               "https://practicum.yandex.ru",
			wantLocationHeader: "https://practicum.yandex.ru",
			wantStatusCode:     http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Установка аргументов командной строки
			args := config.NewArgs()
			args.SetArgs(":8080", "http://localhost:8080")

			req, err := http.NewRequest(http.MethodPost, "http://localhost:8080", strings.NewReader(tt.body))

			if err != nil {
				t.Fatalf("method not alowed: %v", err)
			}

			responseForPost := httptest.NewRecorder()
			ss := shortener.NewService(args.BaseURL, storage.NewRepositoryMemory())
			handler := NewShortenerHandler(ss)
			handler.PostHandler(responseForPost, req)

			urlStorage := ss.Sr.GetAll()

			for _, v := range urlStorage {
				req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/", nil)

				chiCtx := chi.NewRouteContext()
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
				chiCtx.URLParams.Add("id", v.ID())

				if err != nil {
					t.Fatalf("method not alowed: %v", err)
				}

				responseForGet := httptest.NewRecorder()
				handler.GetByIDHandler(responseForGet, req)
				res := responseForGet.Result()

				if res.StatusCode != http.StatusTemporaryRedirect {
					t.Errorf("excpected status 307; got %v", res.StatusCode)
				}

				if location := res.Header.Get("Location"); location != tt.wantLocationHeader {
					t.Errorf("location header does not match: got %v want %v",
						location, tt.wantLocationHeader)
				}

				err = res.Body.Close()

				if err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}
		})
	}
}
