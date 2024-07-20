package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Kenny201/go-yandex-shortener.git/cmd/shortener/config"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

func TestPostWithDataJSONHandler(t *testing.T) {
	tests := []struct {
		name                    string
		body                    string
		wantResponseContentType string
		wantStatusCode          int
	}{
		{
			name:                    "post json request for body https://yandex.ru",
			body:                    "{\"url\": \"https://yandex.ru\"}",
			wantResponseContentType: "application/json",
			wantStatusCode:          http.StatusCreated,
		},
		{
			name:                    "post json request for body https://practicum.yandex.ru",
			body:                    "{\"url\": \"https://practicum.yandex.ru\"}",
			wantResponseContentType: "application/json",
			wantStatusCode:          http.StatusCreated,
		},
		{
			name:                    "post json request for empty body",
			body:                    "",
			wantResponseContentType: "text/plain; charset=utf-8",
			wantStatusCode:          http.StatusBadRequest,
		},
		{
			name:                    "post request when body isn`t json type",
			body:                    "https://practicum.yandex.ru",
			wantResponseContentType: "text/plain; charset=utf-8",
			wantStatusCode:          http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := config.NewArgs()
			args.SetArgs(":8080", "http://localhost:8080")

			req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/shorten", strings.NewReader(tt.body))

			if err != nil {
				t.Fatalf("method not alowed: %v", err)
			}

			w := httptest.NewRecorder()

			ss := shortener.NewService(args.BaseURL, storage.NewRepositoryMemory())

			New(ss).PostWithDataJSON(w, req)

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
					t.Errorf("failed to close response body: %v", err)
				}
			}()

			if err != nil {
				t.Fatalf("could not read response:%v", err)
			}
		})
	}
}
