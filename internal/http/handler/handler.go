package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/aggregate"
	"github.com/go-chi/chi/v5"
)

const (
	NotReadRequestBody = "error reading request body"
	NotUnmarshall      = "error unmarshall"
	NotMarshall        = "error marshall"
	RequestBodyIsEmpty = "request body is empty"
	BadRequest         = "bad request"
	URLFieldIsEmpty    = "the url field cannot be empty"
)

var (
	ErrURLIsEmpty       = errors.New(URLFieldIsEmpty)
	ErrReadAll          = errors.New(NotReadRequestBody)
	ErrRequestBodyEmpty = errors.New(RequestBodyIsEmpty)
)

type (
	ErrorResponse struct {
		Code   int    `json:"code"`
		Error  string `json:"error"`
		Detail string `json:"detail,omitempty"`
	}

	ShortenerService interface {
		Put(url string) (string, error)
		Get(url string) (*aggregate.URL, error)
	}

	Handler struct {
		shortenerService ShortenerService
	}
)

func New(ss ShortenerService) Handler {
	return Handler{
		shortenerService: ss,
	}
}

func (sh Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	url, err := sh.shortenerService.Get(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Location", url.OriginalURL())
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (sh Handler) Post(w http.ResponseWriter, r *http.Request) {
	var shortURL string

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, ErrReadAll.Error(), http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		http.Error(w, ErrRequestBodyEmpty.Error(), http.StatusBadRequest)
		return
	}

	shortURL, err = sh.shortenerService.Put(string(body))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func ErrorJSONResponse(w http.ResponseWriter, code int, error string, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Code: code, Error: error, Detail: message})
}

func JSONResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.WriteHeader(statusCode)

	if payload == nil {
		return
	}

	data, err := json.Marshal(payload)

	if err != nil {
		ErrorJSONResponse(w, http.StatusBadRequest, NotMarshall, err.Error())
		return
	}

	w.Write(data)
}
