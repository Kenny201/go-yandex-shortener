package config

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http/handler"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

const (
	URL = "http://localhost:8080"
)

func TestFlagsWithError(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		args           map[string]string
		wantError      string
		wantStatusCode int
	}{
		{
			name: "incorrect base_url:not set scheme into argument shortener_base_url",
			body: "https://yandex.ru",
			args: map[string]string{
				"shortener_server_address": "http://localhost:8080",
				"shortener_base_url":       "://localhost:8080",
			},
			wantError:      "failed to parse base url\n",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "incorrect base_url:not set port into argument shortener_base_url",
			body: "https://practicum.yandex.ru",
			args: map[string]string{
				"shortener_server_address": "http://localhost:8080",
				"shortener_base_url":       "http://localhost",
			},
			wantError:      "failed to split host and port\n",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := initArgs(tt.args["shortener_server_address"], tt.args["shortener_base_url"], "")
			rw, r := sendRequest(http.MethodPost, URL, strings.NewReader(tt.body))

			repository := storage.NewMemoryShortenerRepository(args.BaseURL)
			linkShortener := shortener.New(repository)

			handler.New(linkShortener).Post(rw, r)

			res := rw.Result()
			defer responseClose(t, res)

			body, err := io.ReadAll(res.Body)

			if err != nil {
				t.Fatalf("could not read response:%v", err)
			}

			if string(body) != tt.wantError {
				t.Errorf("error handler not correct: got %v want %v", string(body), tt.wantError)
			}

			if res.StatusCode != tt.wantStatusCode {
				t.Errorf("excpected status: got %v want %v", res.StatusCode, tt.wantStatusCode)
			}
		})
	}
}

func TestFlags(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		args           map[string]string
		wantStatusCode int
	}{
		{
			name: "set port 8080 into argument shortener_server_address",
			body: "https://yandex.ru",
			args: map[string]string{
				"shortener_server_address": "http://localhost:8080",
				"shortener_base_url":       "http://localhost:8080",
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name: "set port 8090 into argument shortener_server_address",
			body: "https://yandex.ru",
			args: map[string]string{
				"shortener_server_address": "http://localhost:8090",
				"shortener_base_url":       "http://localhost:8080",
			},
			wantStatusCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := initArgs(tt.args["shortener_server_address"], tt.args["shortener_base_url"], "")
			rw, r := sendRequest(http.MethodPost, URL, strings.NewReader(tt.body))

			repository := storage.NewMemoryShortenerRepository(args.BaseURL)
			service := shortener.New(repository)

			handler.New(service).Post(rw, r)

			res := rw.Result()
			defer responseClose(t, res)

			_, err := io.ReadAll(res.Body)

			if err != nil {
				t.Fatalf("could not read response:%v", err)
			}

			if res.StatusCode != tt.wantStatusCode {
				t.Errorf("excpected status: got %v want %v", res.StatusCode, tt.wantStatusCode)
			}
		})
	}
}

func initArgs(serverAddress, baseURL, filePath string) *Args {
	args := NewArgs()
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
