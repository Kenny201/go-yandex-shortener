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

func NewBaseURL(baseURL string) (BaseURL, error) {
	scheme, hp, err := ParseBaseURL(baseURL)
	var port int

	if len(hp) == 2 {
		port, err = strconv.Atoi(hp[1])
	}

	if err != nil {
		return BaseURL{}, err
	}

	return BaseURL{scheme, hp[0], port}, nil
}

func (bu BaseURL) ToString() string {
	if bu.port != 0 {
		return fmt.Sprintf("%s://%s:%d", bu.scheme, bu.host, bu.port)
	}

	return fmt.Sprintf("%s://%s", bu.scheme, bu.host)
}

func ParseBaseURL(s string) (string, []string, error) {
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
