package message

import "net/http"

type ErrResponse struct {
	HTTPStatusCode int    `json:"-"`
	StatusText     string `json:"error"`
	ErrorText      string `json:"detail,omitempty"`
}

var (
	ErrInternalError = &ErrResponse{HTTPStatusCode: http.StatusBadRequest, StatusText: "Error Internal", ErrorText: ""}
	ErrBadRequest    = &ErrResponse{HTTPStatusCode: http.StatusBadRequest, StatusText: "Bad Request", ErrorText: ""}
	ErrEmptyURL      = &ErrResponse{HTTPStatusCode: http.StatusBadRequest, StatusText: "Bad Request", ErrorText: "URL is empty"}
)
