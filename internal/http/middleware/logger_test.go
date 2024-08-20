package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggerMiddleware(t *testing.T) {
	// Создаем буфер для перехвата логов
	var buf bytes.Buffer

	// Создаем новый логгер с выводом в буфер
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	slog.SetDefault(logger) // Устанавливаем его как дефолтный логгер

	// Создаем тестовый обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	// Оборачиваем наш тестовый обработчик в middleware Logger
	loggedHandler := Logger(handler)

	// Создаем тестовый HTTP-запрос
	req := httptest.NewRequest(http.MethodGet, "/test-uri", nil)
	rec := httptest.NewRecorder()
	response := rec.Result()
	// Закрываем тело ответа
	defer response.Body.Close()

	// Вызываем обработчик
	loggedHandler.ServeHTTP(rec, req)

	// Проверяем, что статус-код и тело ответа соответствуют ожидаемым
	if status := response.StatusCode; status != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, status)
	}

	expectedResponse := "Hello, World!"
	if rec.Body.String() != expectedResponse {
		t.Errorf("expected body %q, got %q", expectedResponse, rec.Body.String())
	}

	// Проверяем содержимое логов
	logs := buf.String()
	if !bytes.Contains([]byte(logs), []byte("Request Info")) {
		t.Error("Expected 'Request Info' log entry")
	}

	if !bytes.Contains([]byte(logs), []byte("Response Info")) {
		t.Error("Expected 'Response Info' log entry")
	}

	if !bytes.Contains([]byte(logs), []byte("uri")) || !bytes.Contains([]byte(logs), []byte("/test-uri")) {
		t.Error("Expected log entry with correct URI")
	}

	if !bytes.Contains([]byte(logs), []byte("method")) || !bytes.Contains([]byte(logs), []byte(http.MethodGet)) {
		t.Error("Expected log entry with correct method")
	}

	if !bytes.Contains([]byte(logs), []byte("status")) || !bytes.Contains([]byte(logs), []byte("200")) {
		t.Error("Expected log entry with correct status code")
	}
}
