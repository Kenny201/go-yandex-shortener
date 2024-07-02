package handlers

import (
	"fmt"
	"github.com/Kenny201/go-yandex-shortener.git/internal/storage"
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

	postHandler(w, req)

	res := w.Result()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("excpected statud OK; got %v", res.StatusCode)
	}

	if ctype := w.Header().Get("Content-Type"); ctype != "text/plain" {
		t.Errorf("content type header does not match: got %v want %v",
			ctype, "text-plain")
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
	postHandler(responseForPost, req)
	urlStorage := *storage.GetStorage()

	for k := range urlStorage {
		host := fmt.Sprintf("http://localhost:8080/%v", k)
		req, err := http.NewRequest(http.MethodGet, host, nil)

		if err != nil {
			t.Fatalf("method not alowed: %v", err)
		}

		responseForGet := httptest.NewRecorder()
		Handler(responseForGet, req)
		res := responseForGet.Result()

		if res.StatusCode != http.StatusTemporaryRedirect {
			t.Errorf("excpected status 307; got %v", res.StatusCode)
		}

		if location := responseForGet.Header().Get("Location"); location != "https://practicum.yandex.ru" {
			t.Errorf("location header does not match: got %v want %v",
				location, "https://practicum.yandex.ru")
		}
	}
}
