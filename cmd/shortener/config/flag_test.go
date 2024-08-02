package config

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener/strategy"
	"github.com/Kenny201/go-yandex-shortener.git/internal/http/handler"
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
			service := initService(strategy.NewMemory(args.BaseURL))
			handler.New(service).Post(rw, r)

			res := rw.Result()
			body, err := io.ReadAll(res.Body)

			if err != nil {
				t.Fatalf("could not read response:%v", err)
			}

			assertCorrectError(t, string(body), tt.wantError)
			assertCorrectStatusCode(t, res.StatusCode, tt.wantStatusCode)

			defer responseClose(t, res)
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
			service := initService(strategy.NewMemory(args.BaseURL))
			handler.New(service).Post(rw, r)

			res := rw.Result()
			_, err := io.ReadAll(res.Body)

			if err != nil {
				t.Fatalf("could not read response:%v", err)
			}

			assertCorrectStatusCode(t, res.StatusCode, tt.wantStatusCode)
			defer responseClose(t, res)
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

func initService(strategy strategy.Strategy) *shortener.Service {
	ss := shortener.NewService()
	ss.SetStrategy(strategy)

	return ss
}

func assertCorrectError(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("error handler not correct: got %v want %v", got, want)
	}
}

func assertCorrectStatusCode(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("excpected status: got %v want %v", got, want)
	}
}
