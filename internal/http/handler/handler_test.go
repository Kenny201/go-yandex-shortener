package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"

	"github.com/Kenny201/go-yandex-shortener.git/cmd/shortener/config"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/mocks"
)

func TestPostHandler(t *testing.T) {
	var args = initArgs(t)

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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Создаем mock для интерфейса DB
			mockRepository := mocks.NewMockRepository(ctrl)

			mockRepository.EXPECT().Create(tt.body).Return("some-short-url", nil)

			rw, r := sendRequest(http.MethodPost, args.BaseURL, strings.NewReader(tt.body))
			linkShortener := shortener.New(mockRepository)
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
		})

	}
}

func TestGetHandler(t *testing.T) {
	var args = initArgs(t)

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
			id:             "nonexistent-id",
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Создаем mock для интерфейса Repository
			mockRepository := mocks.NewMockRepository(ctrl)

			var shortURL string
			if tt.wantStatusCode != http.StatusNotFound {
				// Ожидаем, что метод Create вернет короткий URL
				shortURL = "some-short-url"
				mockRepository.EXPECT().Create(tt.body).Return("http://localhost::8080/"+shortURL, nil)
				fmt.Println(shortURL)

				// Ожидаем, что метод Get вернет сущность URL с исходным URL
				mockRepository.EXPECT().Get(shortURL).Return(&entity.URL{
					ID:          "some-id",
					ShortKey:    shortURL,
					OriginalURL: tt.body,
				}, nil)
			} else {
				// Ожидаем, что метод Get вернет ошибку "not found"
				mockRepository.EXPECT().Get(tt.id).Return(nil, fmt.Errorf("not found"))
			}

			// Отправляем запрос на создание короткого URL
			rw, req := sendRequest(http.MethodPost, "http://localhost"+args.ServerAddress, strings.NewReader(tt.body))
			linkShortener := shortener.New(mockRepository)
			New(linkShortener).PostAPI(rw, req)

			// Получаем результат создания
			createResponse := rw.Result()
			defer responseClose(t, createResponse)

			if createResponse.StatusCode != http.StatusCreated {
				t.Errorf("expected status: got %v want %v", createResponse.StatusCode, http.StatusCreated)
			}

			// Теперь отправляем запрос на получение созданного короткого URL
			rw, req = sendRequest(http.MethodGet, "http://localhost"+args.ServerAddress+"/some-short-url", nil)
			New(linkShortener).Get(rw, req)

			response := rw.Result()
			defer responseClose(t, response)

			if response.StatusCode != tt.wantStatusCode {
				t.Errorf("expected status: got %v want %v", response.StatusCode, tt.wantStatusCode)
			}

			if response.StatusCode == http.StatusTemporaryRedirect {
				if response.Header.Get("Location") != tt.wantLocationHeader {
					t.Errorf("response Location header does not match: got %v want %v",
						response.Header.Get("Location"), tt.wantLocationHeader)
				}
			}
		})

	}
}

func TestPostAPIHandler(t *testing.T) {
	var args = initArgs(t)

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
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Создаем mock для интерфейса DB
			mockRepository := mocks.NewMockRepository(ctrl)

			// Парсим тело JSON
			var requestBody map[string]string
			if tt.body != "" && tt.wantStatusCode == http.StatusCreated {
				err := json.Unmarshal([]byte(tt.body), &requestBody)
				if err == nil {
					originalURL := requestBody["url"]

					// Убедитесь, что вы ожидаете вызова Create с правильным аргументом
					mockRepository.EXPECT().Create(originalURL).Return("some-short-url", nil)
				}
			}

			rw, req := sendRequest(http.MethodPost, args.BaseURL, strings.NewReader(tt.body))

			linkShortener := shortener.New(mockRepository)

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
		})
	}
}

func TestPingHandler(t *testing.T) {
	args := initArgs(t)

	tests := []struct {
		name                    string
		body                    string
		wantStatusCode          int
		wantResponseContentType string
	}{
		{
			name:           "test ping handler",
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Создаем mock для интерфейса DB
			mockRepository := mocks.NewMockRepository(ctrl)

			// Настраиваем ожидания для метода Ping
			mockRepository.EXPECT().CheckHealth().Return(nil)

			rw, req := sendRequest(http.MethodGet, args.BaseURL, strings.NewReader(tt.body))
			linkShortener := shortener.New(mockRepository)

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

	host, err = url.Parse(strings.TrimSuffix(string(shortURL), "\n"))

	if err != nil {
		t.Errorf("failed to parse url string: %v", err.Error())
	}

	if host != nil {
		return strings.TrimLeft(host.RequestURI(), "/")
	}

	return ""
}
