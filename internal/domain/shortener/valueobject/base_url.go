package valueobject

import (
	"fmt"
	"net"
	"net/url"
	"strings"
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

	if len(schemeAndHost) == 1 {
		s = fmt.Sprintf("://%s", schemeAndHost[0])
	}

	u, err := url.Parse(s)

	if err != nil {
		return parsedURL, err
	}

	host, port, err := net.SplitHostPort(u.Host)

	if err != nil {
		return parsedURL, err
	}

	parsedURL["scheme"] = u.Scheme
	parsedURL["host"] = host
	parsedURL["port"] = port

	return parsedURL, nil
}
