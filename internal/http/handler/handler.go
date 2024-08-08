package handler

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Kenny201/go-yandex-shortener.git/internal/app/shortener"
	"github.com/Kenny201/go-yandex-shortener.git/internal/infra/storage"
)

const (
	FailedReadRequestBody = "failed reading request body"
	FailedUnmarshall      = "failed unmarshall"
	FailedMarshall        = "failed marshall"
	RequestBodyIsEmpty    = "request body is empty"
	BadRequest            = "bad request"
	URLFieldIsEmpty       = "the url field cannot be empty"
)

var (
	ErrURLIsEmpty       = errors.New(URLFieldIsEmpty)
	ErrReadAll          = errors.New(FailedReadRequestBody)
	ErrRequestBodyEmpty = errors.New(RequestBodyIsEmpty)
)

type (
	Handler struct {
		shortenerService shortener.Shortener
	}
)

func New(ss *shortener.Shortener) Handler {
	return Handler{
		shortenerService: *ss,
	}
}

func (sh Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	url, err := sh.shortenerService.GetShortURL(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Location", url.OriginalURL)
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

	shortURL, err = sh.shortenerService.CreateShortURL(string(body))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (sh Handler) Ping(w http.ResponseWriter, r *http.Request) {
	var value interface{} = sh.shortenerService.Repository

	switch value.(type) {
	case *storage.DatabaseShortenerRepository:
		db := value.(*storage.DatabaseShortenerRepository).DB

		if err := db.Ping(); err != nil {
			db.Close()
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	default:
		log.Fatal("Can't use this strategy to work with the database!")
	}
}
