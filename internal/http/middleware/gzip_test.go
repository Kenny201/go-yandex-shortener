package middleware

import (
	"compress/gzip"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzipCompression(t *testing.T) {
	responseBody := `{"url": "https://practicum.yandex.ru"}`
	r := chi.NewRouter()
	r.Use(Gzip)

	r.Get("/getjson", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(responseBody))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []struct {
		name              string
		path              string
		expectedEncoding  string
		acceptedEncodings []string
	}{
		{
			name:              "gzip is only encoding",
			path:              "/getjson",
			acceptedEncodings: []string{"gzip"},
			expectedEncoding:  "gzip",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp, respString := testRequestWithAcceptedEncodings(t, ts, "GET", tc.path, tc.acceptedEncodings...)
			if respString != responseBody {
				t.Errorf("response text doesn't match; expected:%q, got:%q", responseBody, respString)
			}
			if got := resp.Header.Get("Content-Encoding"); got != tc.expectedEncoding {
				t.Errorf("expected encoding %q but got %q", tc.expectedEncoding, got)
			}

			defer resp.Body.Close()
		})
	}
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

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		var err error
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
	default:
		reader = resp.Body
	}

	respBody, err := io.ReadAll(reader)

	if err != nil {
		t.Fatal(err)
		return ""
	}

	defer reader.Close()

	return string(respBody)
}
