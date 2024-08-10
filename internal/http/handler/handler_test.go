package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Kenny201/go-yandex-shortener.git/cmd/shortener/config"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/closer"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

func TestPostHandler(t *testing.T) {
	var args = initArgs(t)

	var repositoryMemory = storage.NewMemoryShortenerRepository(args.BaseURL)
	var repositoryFile, _ = storage.NewFileShortenerRepository(args.BaseURL, args.FileStoragePath)

	reps := []shortener.Repository{
		repositoryMemory,
		repositoryFile,
	}

	tests := []struct {
		name                    string
		body                    string
		repositories            []shortener.Repository
		wantResponseContentType string
		wantStatusCode          int
	}{
		{
			name:                    "post_request_for_body_https://yandex.ru",
			body:                    "https://yandex.ru",
			repositories:            reps,
			wantResponseContentType: "text/plain",
			wantStatusCode:          http.StatusCreated,
		},
		{
			name:                    "post_request_for_body_https://practicum.yandex.ru",
			body:                    "https://practicum.yandex.ru",
			repositories:            reps,
			wantResponseContentType: "text/plain",
			wantStatusCode:          http.StatusCreated,
		},
		{
			name:                    "post_request_for_empty_body",
			body:                    "",
			repositories:            reps,
			wantResponseContentType: "text/plain; charset=utf-8",
			wantStatusCode:          http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, repository := range tt.repositories {
				rw, r := sendRequest(http.MethodPost, args.BaseURL, strings.NewReader(tt.body))
				linkShortener := shortener.New(repository)
				New(linkShortener).Post(rw, r)

				response := rw.Result()
				defer responseClose(t, response)

				if response.StatusCode != tt.wantStatusCode {
					t.Errorf("excpected status: got %v want %v", response.StatusCode, tt.wantStatusCode)
				}

				if response.Header.Get("Content-Type") != tt.wantResponseContentType {
					t.Errorf("response content type header does not match: got %v wantResponseContentType %v",
						response.Header.Get("Content-Type"), tt.wantResponseContentType)
				}
			}
		})

	}
}

func TestGetHandler(t *testing.T) {
	var args = initArgs(t)

	var repositoryMemory = storage.NewMemoryShortenerRepository(args.BaseURL)
	var repositoryFile, _ = storage.NewFileShortenerRepository(args.BaseURL, args.FileStoragePath)

	reps := []shortener.Repository{
		repositoryMemory,
		repositoryFile,
	}

	tests := []struct {
		name               string
		body               string
		id                 string
		repositories       []shortener.Repository
		wantLocationHeader string
		wantStatusCode     int
	}{
		{
			name:               "redirect_for_body_https://yandex.ru",
			body:               "https://yandex.ru",
			repositories:       reps,
			wantLocationHeader: "https://yandex.ru",
			wantStatusCode:     http.StatusTemporaryRedirect,
		},
		{
			name:               "redirect_for_body_https://practicum.yandex.ru",
			body:               "https://practicum.yandex.ru",
			repositories:       reps,
			wantLocationHeader: "https://practicum.yandex.ru",
			wantStatusCode:     http.StatusTemporaryRedirect,
		},
		{
			name:           "id_not_found",
			body:           "https://yandex.ru",
			id:             "sdsds",
			repositories:   reps,
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, repository := range tt.repositories {
				rw, req := sendRequest(http.MethodPost, args.BaseURL, strings.NewReader(tt.body))
				linkShortener := shortener.New(repository)

				handler := New(linkShortener)
				handler.Post(rw, req)
				responsePost := rw.Result()
				defer responseClose(t, responsePost)

				// Получает shortKey из строки сокращённого url
				shortKey := getShortKeyFromShortedURL(t, responsePost.Body)

				req = httptest.NewRequest(http.MethodGet, args.BaseURL, nil)

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
			}
		})
	}
}

func TestPostAPIHandler(t *testing.T) {
	var args = initArgs(t)

	var repositoryMemory = storage.NewMemoryShortenerRepository(args.BaseURL)
	var repositoryFile, _ = storage.NewFileShortenerRepository(args.BaseURL, args.FileStoragePath)

	reps := []shortener.Repository{
		repositoryMemory,
		repositoryFile,
	}

	tests := []struct {
		name                    string
		body                    string
		repositories            []shortener.Repository
		wantStatusCode          int
		wantResponseContentType string
	}{
		{
			name:                    "post json request for body https://yandex.ru",
			body:                    `{"url": "https://yandex.ru"}`,
			repositories:            reps,
			wantStatusCode:          http.StatusCreated,
			wantResponseContentType: "application/json",
		},
		{
			name:                    "post json request for body https://practicum.yandex.ru",
			body:                    `{"url": "https://practicum.yandex.ru"}`,
			repositories:            reps,
			wantStatusCode:          http.StatusCreated,
			wantResponseContentType: "application/json",
		},
		{
			name:                    "post json request for empty body",
			body:                    "",
			repositories:            reps,
			wantStatusCode:          http.StatusBadRequest,
			wantResponseContentType: "application/json",
		},
		{
			name:                    "post request when body isn`t json type",
			body:                    "https://practicum.yandex.ru",
			repositories:            reps,
			wantStatusCode:          http.StatusBadRequest,
			wantResponseContentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, repository := range tt.repositories {
				rw, req := sendRequest(http.MethodPost, args.BaseURL, strings.NewReader(tt.body))

				linkShortener := shortener.New(repository)

				New(linkShortener).PostAPI(rw, req)

				response := rw.Result()
				defer responseClose(t, response)

				if response.StatusCode != tt.wantStatusCode {
					t.Errorf("excpected status: got %v want %v", response.StatusCode, tt.wantStatusCode)
				}

				if response.Header.Get("Content-Type") != tt.wantResponseContentType {
					t.Errorf("response content type header does not match: got %v wantResponseContentType %v",
						response.Header.Get("Content-Type"), tt.wantResponseContentType)
				}
			}
		})
	}
}

func TestPingHandler(t *testing.T) {
	closer.New()
	args := initArgs(t)

	repositoryDB, err := storage.NewDatabaseShortenerRepository(args.BaseURL, args.DatabaseDNS)

	if err != nil {
		t.Errorf("%v with databaseDNS: %s", err.Error(), args.DatabaseDNS)
	}

	tests := []struct {
		name                    string
		body                    string
		repository              *storage.DatabaseShortenerRepository
		wantStatusCode          int
		wantResponseContentType string
	}{
		{
			name:           "test ping handler",
			repository:     repositoryDB,
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rw, req := sendRequest(http.MethodGet, args.BaseURL, strings.NewReader(tt.body))
			linkShortener := shortener.New(tt.repository)

			New(linkShortener).Ping(rw, req)

			response := rw.Result()
			defer responseClose(t, response)

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("excpected status: got %v want %v", response.StatusCode, tt.wantStatusCode)
			}
		})
	}
}

func initArgs(t *testing.T) *config.Args {
	t.Helper()
	conf, err := config.LoadConfig("../../../")

	if err != nil {
		t.Errorf("error read config %v", err)
	}

	serverAddress := fmt.Sprintf(":%s", conf.Port)
	baseURL := fmt.Sprintf("http://localhost:%s", conf.Port)
	fileStoragePath := "tmp/Rquxc"
	databaseDNS := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", conf.DBUsername, conf.DBPassword, conf.DBHost, conf.DBPort, conf.DBDatabase)

	args := []string{
		"-a", serverAddress,
		"-b", baseURL,
		"-f", fileStoragePath,
		"-d", databaseDNS,
	}

	a := config.NewArgs(conf)
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
