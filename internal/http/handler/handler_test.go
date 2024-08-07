package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Kenny201/go-yandex-shortener.git/cmd/shortener/config"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

const (
	serverAddress = ":8080"
	baseURL       = "http://localhost:8080"
	URL           = "http://localhost:8080"
)

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
		t.Run("test_for_memory_strategy_"+tt.name, func(t *testing.T) {
			args := initArgs(serverAddress, baseURL, "")
			rw, r := sendRequest(http.MethodPost, URL, strings.NewReader(tt.body))
			repository := storage.NewMemoryShortenerRepository(args.BaseURL)
			service := shortener.NewService(repository)
			New(service).Post(rw, r)

			response := rw.Result()
			defer responseClose(t, response)

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("excpected status: got %v want %v", response.StatusCode, tt.wantStatusCode)
			}

			if response.Header.Get("Content-Type") != tt.wantResponseContentType {
				t.Errorf("response content type header does not match: got %v wantResponseContentType %v",
					response.Header.Get("Content-Type"), tt.wantResponseContentType)
			}
		})

		t.Run("test_for_file_strategy_"+tt.name, func(t *testing.T) {
			args := initArgs(serverAddress, baseURL, "urls.txt")
			rw, r := sendRequest(http.MethodPost, URL, strings.NewReader(tt.body))
			repository, err := storage.NewFile(args.BaseURL, args.FileStoragePath)

			if err != nil {
				t.Errorf("failed create repository file error: %v", err.Error())
			}

			service := shortener.NewService(repository)

			New(service).Post(rw, r)

			response := rw.Result()
			defer responseClose(t, response)

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("excpected status: got %v want %v", response.StatusCode, tt.wantStatusCode)
			}

			if response.Header.Get("Content-Type") != tt.wantResponseContentType {
				t.Errorf("response content type header does not match: got %v wantResponseContentType %v",
					response.Header.Get("Content-Type"), tt.wantResponseContentType)
			}
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
			name:               "redirect_for_body_https://yandex.ru",
			body:               "https://yandex.ru",
			wantLocationHeader: "https://yandex.ru",
			wantStatusCode:     http.StatusTemporaryRedirect,
		},
		{
			name:               "redirect_for_body_https://practicum.yandex.ru",
			body:               "https://practicum.yandex.ru",
			wantLocationHeader: "https://practicum.yandex.ru",
			wantStatusCode:     http.StatusTemporaryRedirect,
		},
		{
			name:           "id_not_found",
			body:           "https://yandex.ru",
			id:             "sdsds",
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run("test_for_memory_strategy_"+tt.name, func(t *testing.T) {
			args := initArgs(serverAddress, baseURL, "")
			rw, req := sendRequest(http.MethodPost, URL, strings.NewReader(tt.body))

			repository := storage.NewMemoryShortenerRepository(args.BaseURL)
			service := shortener.NewService(repository)

			handler := New(service)
			handler.Post(rw, req)
			responsePost := rw.Result()
			defer responseClose(t, responsePost)

			// Получает shortKey из строки сокращённого url
			shortKey := getShortKeyFromShortedURL(t, responsePost.Body)

			req = httptest.NewRequest(http.MethodGet, URL, nil)

			if tt.id != "" {
				req = withURLParam(req, "id", tt.id)
			} else {
				req = withURLParam(req, "id", shortKey)
			}

			w := httptest.NewRecorder()
			handler.Get(w, req)
			responseGet := w.Result()
			defer responseClose(t, responseGet)

			if responseGet.StatusCode != tt.wantStatusCode {
				t.Errorf("excpected status: got %v want %v", responseGet.StatusCode, tt.wantStatusCode)
			}

			if responseGet.Header.Get("Location") != tt.wantLocationHeader {
				t.Errorf("location header does not match: got %v want %v",
					responseGet.Header.Get("Location"), tt.wantLocationHeader)
			}
		})

		t.Run("test_for_file strategy_"+tt.name, func(t *testing.T) {
			args := initArgs(serverAddress, baseURL, "urls_get.txt")
			rw, req := sendRequest(http.MethodPost, URL, strings.NewReader(tt.body))

			repository, err := storage.NewFile(args.BaseURL, args.FileStoragePath)

			if err != nil {
				t.Errorf("failed create repository file error: %v", err.Error())
			}

			service := shortener.NewService(repository)

			handler := New(service)
			handler.Post(rw, req)
			responsePost := rw.Result()
			defer responseClose(t, responsePost)

			// Получает shortKey из строки сокращённого url
			shortKey := getShortKeyFromShortedURL(t, responsePost.Body)

			req = httptest.NewRequest(http.MethodGet, URL, nil)

			if tt.id != "" {
				req = withURLParam(req, "id", tt.id)
			} else {
				req = withURLParam(req, "id", shortKey)
			}

			w := httptest.NewRecorder()
			handler.Get(w, req)
			response := w.Result()
			defer responseClose(t, response)

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("excpected status: got %v want %v", response.StatusCode, tt.wantStatusCode)
			}

			if response.Header.Get("Location") != tt.wantLocationHeader {
				t.Errorf("location header does not match: got %v want %v",
					response.Header.Get("Location"), tt.wantLocationHeader)
			}
		})
	}
}

func TestPostAPIHandler(t *testing.T) {
	tests := []struct {
		name                    string
		body                    string
		wantStatusCode          int
		wantResponseContentType string
	}{
		{
			name:                    "post json request for body https://yandex.ru",
			body:                    `{"url": "https://yandex.ru"}`,
			wantStatusCode:          http.StatusCreated,
			wantResponseContentType: "application/json",
		},
		{
			name:                    "post json request for body https://practicum.yandex.ru",
			body:                    `{"url": "https://practicum.yandex.ru"}`,
			wantStatusCode:          http.StatusCreated,
			wantResponseContentType: "application/json",
		},
		{
			name:                    "post json request for empty body",
			body:                    "",
			wantStatusCode:          http.StatusBadRequest,
			wantResponseContentType: "application/json",
		},
		{
			name:                    "post request when body isn`t json type",
			body:                    "https://practicum.yandex.ru",
			wantStatusCode:          http.StatusBadRequest,
			wantResponseContentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run("test_for_memory_strategy_"+tt.name, func(t *testing.T) {
			args := initArgs(serverAddress, baseURL, "")
			rw, req := sendRequest(http.MethodPost, URL, strings.NewReader(tt.body))

			repository := storage.NewMemoryShortenerRepository(args.BaseURL)
			ss := shortener.NewService(repository)

			New(ss).PostAPI(rw, req)

			response := rw.Result()
			defer responseClose(t, response)

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("excpected status: got %v want %v", response.StatusCode, tt.wantStatusCode)
			}

			if response.Header.Get("Content-Type") != tt.wantResponseContentType {
				t.Errorf("response content type header does not match: got %v wantResponseContentType %v",
					response.Header.Get("Content-Type"), tt.wantResponseContentType)
			}
		})

		t.Run("test for file strategy "+tt.name, func(t *testing.T) {
			args := initArgs(serverAddress, baseURL, "urls_post_api.txt")
			rw, req := sendRequest(http.MethodPost, URL, strings.NewReader(tt.body))

			repository, err := storage.NewFile(args.BaseURL, args.FileStoragePath)

			if err != nil {
				t.Errorf("failed create repository file error:%v", err.Error())
			}

			ss := shortener.NewService(repository)

			New(ss).PostAPI(rw, req)

			response := rw.Result()
			defer responseClose(t, response)

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("excpected status: got %v want %v", response.StatusCode, tt.wantStatusCode)
			}

			if response.Header.Get("Content-Type") != tt.wantResponseContentType {
				t.Errorf("response content type header does not match: got %v wantResponseContentType %v",
					response.Header.Get("Content-Type"), tt.wantResponseContentType)
			}
		})
	}
}

func initArgs(serverAddress, baseURL, filePath string) *config.Args {
	args := config.NewArgs()
	args.SetArgs(serverAddress, baseURL, filePath)

	return args
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

func getShortKeyFromShortedURL(t *testing.T, body io.ReadCloser) string {
	t.Helper()
	var host *url.URL

	shortURL, err := io.ReadAll(body)

	if err != nil {
		t.Errorf("failed to read response body: %v", err.Error())
	}

	host, err = url.Parse(string(shortURL))

	if err != nil {
		t.Errorf("failed to parse url string: %v", err.Error())
	}

	shortKey := strings.TrimLeft(host.RequestURI(), "/")

	return shortKey
}
