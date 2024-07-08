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

func NewOriginalURL(baseURL string) (OriginalURL, error) {
	scheme, hp, err := ParseOriginalURL(baseURL)
	var port int

	if len(hp) == 2 {
		port, err = strconv.Atoi(hp[1])
	}

	if err != nil {
		return OriginalURL{}, err
	}

	return OriginalURL{scheme, hp[0], port}, nil
}

func (bu OriginalURL) Scheme() string {
	return bu.scheme
}

func (bu OriginalURL) Host() string {
	return bu.host
}

func (bu OriginalURL) Port() int {
	return bu.port
}

func (bu OriginalURL) ToString() string {
	if bu.port != 0 {
		return fmt.Sprintf("%s://%s:%d", bu.scheme, bu.host, bu.port)
	}

	return fmt.Sprintf("%s://%s", bu.scheme, bu.host)
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
