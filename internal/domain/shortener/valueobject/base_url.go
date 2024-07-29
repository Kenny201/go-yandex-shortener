package valueobject

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

var (
	ErrParseBaseURL  = errors.New("failed to parse base url")
	ErrSplitHostPort = errors.New("failed to split host and port")
)

type BaseURL struct {
	scheme string
	host   string
	port   string
}

func NewBaseURL(host string) (BaseURL, error) {
	parsedURL, err := ParseBaseURL(host)

	if err != nil {
		return BaseURL{}, err
	}

	return BaseURL{parsedURL["scheme"], parsedURL["host"], parsedURL["port"]}, nil
}

// ToString Преобразовать в строку формата: scheme://host:port
func (bu BaseURL) ToString() string {
	return fmt.Sprintf("%s://%s:%s", bu.scheme, bu.host, bu.port)
}

// Host Получить хост
func (bu BaseURL) Host() string {
	return bu.host
}

// Port Получить номер порта
func (bu BaseURL) Port() string {
	return bu.port
}

// ParseBaseURL Распарсить URL на схему, хост и порт
func ParseBaseURL(s string) (map[string]string, error) {
	parsedURL := make(map[string]string, 3)
	schemeAndHost := strings.Split(s, "://")
	var host, port string

	if len(schemeAndHost) == 1 {
		s = fmt.Sprintf("://%s", schemeAndHost[0])
	}

	u, err := url.Parse(s)

	if err != nil {
		return parsedURL, ErrParseBaseURL
	}

	host, port, err = net.SplitHostPort(u.Host)

	if err != nil {
		return parsedURL, ErrSplitHostPort
	}

	parsedURL["scheme"] = u.Scheme
	parsedURL["host"] = host
	parsedURL["port"] = port

	return parsedURL, nil
}
