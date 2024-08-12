package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

const (
	FailedReadRequestBody = "failed reading request body"
	FailedUnmarshall      = "failed to unmarshal request body"
	FailedMarshall        = "failed to marshal response"
	RequestBodyIsEmpty    = "request body is empty"
	BadRequest            = "bad request"
	URLFieldIsEmpty       = "the url field cannot be empty"
)

var (
	ErrURLIsEmpty       = errors.New(URLFieldIsEmpty)
	ErrReadAll          = errors.New(FailedReadRequestBody)
	ErrRequestBodyEmpty = errors.New(RequestBodyIsEmpty)
)

// Handler управляет HTTP-запросами, связанными с сокращением URL-адресов.
type Handler struct {
	shortenerService shortener.Shortener
}

// New создает новый экземпляр Handler с заданным сервисом сокращения URL.
func New(ss *shortener.Shortener) Handler {
	return Handler{
		shortenerService: *ss,
	}
}

// Get обрабатывает GET-запрос для получения оригинального URL по короткому ключу.
// Возвращает 302 Redirect с заголовком Location.
func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	url, err := h.shortenerService.GetShortURL(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Location", url.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// Post обрабатывает POST-запрос для создания короткого URL.
// Ожидает URL в теле запроса и возвращает короткий URL или ошибку.
func (h Handler) Post(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, ErrReadAll.Error(), http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		http.Error(w, ErrRequestBodyEmpty.Error(), http.StatusBadRequest)
		return
	}

	shortURL, err := h.shortenerService.CreateShortURL(string(body))

	if err != nil {
		if errors.Is(err, storage.ErrURLAlreadyExist) {
			http.Error(w, shortURL, http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

// Ping обрабатывает запрос для проверки состояния сервиса.
// Возвращает 200 OK если состояние сервиса в порядке, иначе 500 Internal Server Error.
func (h Handler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.shortenerService.CheckHealth(); err != nil {
		http.Error(w, "Health check failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pong"))
}
