package valueobject

import (
	"fmt"
	"strconv"
	"strings"
)

type OriginalURL struct {
	scheme string
	host   string
	port   int
}

func NewOriginalURL(originalURL string) (OriginalURL, error) {
	scheme, hp, err := ParseOriginalURL(originalURL)
	var port int

	if len(hp) == 2 {
		port, err = strconv.Atoi(hp[1])
	}

	if err != nil {
		return OriginalURL{}, err
	}

	return OriginalURL{scheme, hp[0], port}, nil
}

func (ou OriginalURL) Scheme() string {
	return ou.scheme
}

func (ou OriginalURL) Host() string {
	return ou.host
}

func (ou OriginalURL) Port() int {
	return ou.port
}

func (ou OriginalURL) ToString() string {
	if ou.port != 0 {
		return fmt.Sprintf("%s://%s:%d", ou.scheme, ou.host, ou.port)
	}

	return fmt.Sprintf("%s://%s", ou.scheme, ou.host)
}

func ParseOriginalURL(s string) (string, []string, error) {
	var protocol string
	var host string

	url := strings.Split(s, "://")

	if len(url) != 2 {
		protocol = "http"
		host = url[0]
	} else {
		protocol = url[0]
		host = url[1]
	}

	hp := strings.Split(host, ":")

	return protocol, hp, nil
}
