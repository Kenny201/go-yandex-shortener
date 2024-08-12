package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// TestGzipCompression проверяет, что ответ сервера корректно сжимается в формате gzip,
// если клиент запрашивает gzip-сжатие.
func TestGzipCompression(t *testing.T) {
	responseBody := `{"url": "https://practicum.yandex.ru"}` // Ожидаемое тело ответа в формате JSON.
	r := chi.NewRouter()
	r.Use(Gzip) // Подключаем middleware для сжатия ответа в формате gzip.

	// Определяем маршрут /getjson, который возвращает JSON-ответ.
	r.Get("/getjson", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseBody))
	})

	// Создаем тестовый сервер с маршрутом.
	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []struct {
		name              string   // Название теста.
		path              string   // Путь запроса.
		expectedEncoding  string   // Ожидаемое значение заголовка Content-Encoding.
		acceptedEncodings []string // Список поддерживаемых клиентом форматов сжатия.
	}{
		{
			name:              "gzip is only encoding", // Тестирует, когда клиент поддерживает только gzip.
			path:              "/getjson",
			acceptedEncodings: []string{"gzip"},
			expectedEncoding:  "gzip",
		},
	}

	// Выполняем тесты из списка.
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Выполняем запрос с указанными поддерживаемыми форматами сжатия.
			resp, respString := testRequestWithAcceptedEncodings(t, ts, "GET", tc.path, tc.acceptedEncodings...)

			// Проверяем, что тело ответа соответствует ожидаемому.
			if respString != responseBody {
				t.Errorf("response text doesn't match; expected:%q, got:%q", responseBody, respString)
			}

			// Проверяем, что сервер отправил ответ с ожидаемым форматом сжатия.
			if got := resp.Header.Get("Content-Encoding"); got != tc.expectedEncoding {
				t.Errorf("expected encoding %q but got %q", tc.expectedEncoding, got)
			}

			defer resp.Body.Close()
		})
	}
}

// testRequestWithAcceptedEncodings выполняет HTTP-запрос к тестовому серверу с заданными заголовками Accept-Encoding.
func testRequestWithAcceptedEncodings(t *testing.T, ts *httptest.Server, method, path string, encodings ...string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil) // Создаем новый HTTP-запрос.
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	// Если указаны поддерживаемые форматы сжатия, добавляем их в заголовок запроса.
	if len(encodings) > 0 {
		encodingsString := strings.Join(encodings, ",")
		req.Header.Set("Accept-Encoding", encodingsString)
	}

	// Выполняем запрос к серверу.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	// Декодируем тело ответа, если оно сжато.
	respBody := decodeResponseBody(t, resp)

	defer resp.Body.Close()

	return resp, respBody
}

// decodeResponseBody декодирует сжатый ответ сервера в исходный текст.
func decodeResponseBody(t *testing.T, resp *http.Response) string {
	var reader io.ReadCloser
	var err error

	// Если ответ сжат, используем gzip-ридер для его декодирования.
	reader, err = gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	// Читаем и возвращаем декодированное тело ответа.
	respBody, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
		return ""
	}

	defer reader.Close()

	return string(respBody)
}
