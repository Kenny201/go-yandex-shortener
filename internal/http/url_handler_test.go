package http

import (
	"context"
	"fmt"
	"github.com/Kenny201/go-yandex-shortener.git/internal/app/url"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPostHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/", strings.NewReader("https://practicum.yandex.ru"))

	if err != nil {
		t.Fatalf("method not alowed: %v", err)
	}

	w := httptest.NewRecorder()

	us, err := url.NewService(url.WithMemoryRepository())

	if err != nil {
		fmt.Printf("%v", err)

		return
	}

	NewURLHandler(us).PostHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("excpected statud OK; got %v", res.StatusCode)
	}

	if ctype := w.Header().Get("Content-Type"); ctype != "text/plain" {
		t.Errorf("content type header does not match: got %v want %v",
			ctype, "text/plain")
	}

	_, err = io.ReadAll(res.Body)

	if err != nil {
		t.Fatalf("could not read response:%v", err)
	}
}

func TestGetByIDHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080", strings.NewReader("https://practicum.yandex.ru"))
	if err != nil {
		t.Fatalf("method not alowed: %v", err)
	}

	responseForPost := httptest.NewRecorder()
	us, _ := url.NewService(url.WithMemoryRepository())
	handler := NewURLHandler(us)
	handler.PostHandler(responseForPost, req)

	urlStorage := us.Ur.GetAll()

	for _, v := range urlStorage {
		req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/", nil)

		chiCtx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		chiCtx.URLParams.Add("id", v.ID())

		if err != nil {
			t.Fatalf("method not alowed: %v", err)
		}

		responseForGet := httptest.NewRecorder()
		handler.GetByIDHandler(responseForGet, req)
		res := responseForGet.Result()

		if res.StatusCode != http.StatusTemporaryRedirect {
			t.Errorf("excpected status 307; got %v", res.StatusCode)
		}

		if location := responseForGet.Header().Get("Location"); location != "https://practicum.yandex.ru" {
			t.Errorf("location header does not match: got %v want %v",
				location, "https://practicum.yandex.ru")
		}

		res.Body.Close()
	}
}
