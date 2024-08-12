package valueobject

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

var (
	ErrParseBaseURL  = errors.New("не удалось разобрать базовый URL")
	ErrSplitHostPort = errors.New("не удалось разделить хост и порт")
)

// BaseURL представляет собой структуру URL с полями схема, хост и порт.
type BaseURL struct {
	scheme string
	host   string
	port   string
}

// NewBaseURL создает новый BaseURL из данной строки хоста.
// Строка хоста должна быть в формате "scheme://host:port".
func NewBaseURL(host string) (BaseURL, error) {
	parsedURL, err := ParseBaseURL(host)

	if err != nil {
		return BaseURL{}, err
	}

	return BaseURL{
		scheme: parsedURL["scheme"],
		host:   parsedURL["host"],
		port:   parsedURL["port"],
	}, nil
}

// ToString возвращает BaseURL в виде строки в формате "scheme://host:port".
func (bu BaseURL) ToString() string {
	if bu.port == "" {
		return fmt.Sprintf("%s://%s", bu.scheme, bu.host)
	}
	return fmt.Sprintf("%s://%s:%s", bu.scheme, bu.host, bu.port)
}

// ParseBaseURL парсит переданную строку URL на компоненты схема, хост и порт.
// Входная строка должна быть в формате "scheme://host:port" или "host:port".
func ParseBaseURL(s string) (map[string]string, error) {
	parsedURL := make(map[string]string)

	// Убедитесь, что строка URL имеет схему
	if !strings.Contains(s, "://") {
		s = "://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, ErrParseBaseURL
	}

	host, port, err := net.SplitHostPort(u.Host)

	if err != nil && err.Error() == "missing port" {
		host = u.Host
		port = ""
	} else if err != nil {
		return nil, ErrSplitHostPort
	}

	parsedURL["scheme"] = u.Scheme
	parsedURL["host"] = host
	parsedURL["port"] = port

	return parsedURL, nil
}
