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

func TestGzipCompression(t *testing.T) {
	t.Run("gzip is only encoding", func(t *testing.T) {
		responseBody := `{"url": "https://practicum.yandex.ru"}`
		r := chi.NewRouter()
		r.Use(Gzip)

		r.Get("/getjson", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(responseBody))
		})

		ts := httptest.NewServer(r)
		defer ts.Close()

		resp, respString := testRequestWithAcceptedEncodings(t, ts, "GET", "/getjson", "gzip")

		if respString != responseBody {
			t.Errorf("response text doesn't match; expected:%q, got:%q", responseBody, respString)
		}

		if got := resp.Header.Get("Content-Encoding"); got != "gzip" {
			t.Errorf("expected encoding %q but got %q", "gzip", got)
		}

		defer resp.Body.Close()
	})
}

func testRequestWithAcceptedEncodings(t *testing.T, ts *httptest.Server, method, path string, encodings ...string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	if len(encodings) > 0 {
		encodingsString := strings.Join(encodings, ",")
		req.Header.Set("Accept-Encoding", encodingsString)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody := decodeResponseBody(t, resp)

	defer resp.Body.Close()

	return resp, respBody
}

func decodeResponseBody(t *testing.T, resp *http.Response) string {
	var reader io.ReadCloser
	var err error

	reader, err = gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	respBody, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
		return ""
	}

	defer reader.Close()

	return string(respBody)
}
