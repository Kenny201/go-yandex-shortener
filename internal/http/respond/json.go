package respond

import (
	"encoding/json"
	"net/http"

	"github.com/Kenny201/go-yandex-shortener.git/internal/utils/message"
)

func Json(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if payload == nil {
		return
	}

	data, err := json.Marshal(payload)

	if err != nil {
		Error(w, http.StatusInternalServerError, message.ErrInternalError)
		return
	}

	w.Write(data)
}
