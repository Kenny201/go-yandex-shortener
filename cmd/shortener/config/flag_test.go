package config

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	server "github.com/Kenny201/go-yandex-shortener.git/internal/http"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
			wantError:      "parse \"://localhost:8080\": missing protocol scheme\n",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "incorrect base_url:not set port into argument shortener_base_url",
			body: "https://practicum.yandex.ru",
			args: map[string]string{
				"shortener_server_address": "http://localhost:8080",
				"shortener_base_url":       "http://localhost",
			},
			wantError:      "address localhost: missing port in address\n",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := NewArgs()
			args.SetArgs(tt.args["shortener_server_address"], tt.args["shortener_base_url"])

			req, err := http.NewRequest(http.MethodPost, tt.args["shortener_server_address"], strings.NewReader(tt.body))

			if err != nil {
				t.Fatalf("method not alowed: %v", err)
			}

			w := httptest.NewRecorder()

			ss := shortener.NewService(args.BaseURL, storage.NewRepositoryMemory())

			server.NewShortenerHandler(ss).PostHandler(w, req)

			res := w.Result()
			body, err := io.ReadAll(res.Body)

			if err != nil {
				t.Fatalf("could not read response:%v", err)
			}

			if error := string(body); error != tt.wantError {
				t.Errorf("error handler not correct: got %v want %v",
					error, tt.wantError)
			}

			if res.StatusCode != tt.wantStatusCode {
				t.Errorf("excpected status: got %v want %v", res.StatusCode, tt.wantStatusCode)
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
			args := NewArgs()
			args.SetArgs(tt.args["shortener_server_address"], tt.args["shortener_base_url"])

			req, err := http.NewRequest(http.MethodPost, tt.args["shortener_server_address"], strings.NewReader(tt.body))

			if err != nil {
				t.Fatalf("method not alowed: %v", err)
			}

			w := httptest.NewRecorder()

			ss := shortener.NewService(args.BaseURL, storage.NewRepositoryMemory())

			server.NewShortenerHandler(ss).PostHandler(w, req)

			res := w.Result()

			if err != nil {
				t.Fatalf("could not read response:%v", err)
			}

			if res.StatusCode != tt.wantStatusCode {
				t.Errorf("excpected status %v; got %v", tt.wantStatusCode, res.StatusCode)
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
