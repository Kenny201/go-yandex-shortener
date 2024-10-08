package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type (
	// Берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	if r.responseData.status == 0 { // Если статус еще не был установлен
		r.responseData.status = statusCode // захватываем код статуса
		r.ResponseWriter.WriteHeader(statusCode)
	}
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{}

		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}

		next.ServeHTTP(&lw, r)

		slog.Info(
			"Request Info",
			slog.String("uri", r.RequestURI),
			slog.String("method", r.Method),
			slog.Duration("duration", time.Since(start)),
		)

		slog.Info(
			"Response Info",
			slog.Int("status", responseData.status),
			slog.Int("size", responseData.size),
		)

	})
}
