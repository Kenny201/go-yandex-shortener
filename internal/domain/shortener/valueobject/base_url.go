package valueobject

import (
	"fmt"
	"strconv"
	"strings"
)

type BaseURL struct {
	scheme string
	host   string
	port   int
}

func NewBaseURL(host string) (BaseURL, error) {
	scheme, hp, err := ParseBaseURL(host)
	var port int

	if err != nil {
		return BaseURL{}, err
	}

	if len(hp) == 2 {
		port, err = strconv.Atoi(hp[1])
	}

	if err != nil {
		return BaseURL{}, err
	}

	return BaseURL{scheme, hp[0], port}, nil
}

func (sa BaseURL) ToString() string {
	if sa.port == 0 {
		return fmt.Sprintf("%s://%s", sa.scheme, sa.host)
	}

	return fmt.Sprintf("%s://%s:%d", sa.scheme, sa.host, sa.port)
}

func (sa BaseURL) Host() string {
	return sa.host
}

func (sa BaseURL) Port() int {
	return sa.port
}

func ParseBaseURL(s string) (string, []string, error) {
	host := strings.Split(s, "://")
	var hp []string
	var scheme string

	if len(host) == 1 {
		scheme = "http"
		hp = strings.Split(host[0], ":")
	} else {
		scheme = host[0]
		hp = strings.Split(host[1], ":")
	}

	return scheme, hp, nil
}
