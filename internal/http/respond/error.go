package respond

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Kenny201/go-yandex-shortener.git/internal/utils/message"
)

func Error(w http.ResponseWriter, statusCode int, error *message.ErrResponse) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(statusCode)

	if error == nil {
		w.Write([]byte("{}"))
		return
	}

	data, err := json.Marshal(error)

	if err != nil {
		slog.Error(
			fmt.Sprintf("error marshalling error response with detail message: %s", error.ErrorText),
		)
		return
	}

	w.Write(data)
}
