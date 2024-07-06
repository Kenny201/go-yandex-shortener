package config

import (
	"errors"
	"strings"
)

func ParseServerAddress(s string) ([]string, error) {
	host := strings.Split(s, "://")
	var hp []string

	if len(host) != 2 {
		hp = strings.Split(host[0], ":")
	} else {
		hp = strings.Split(host[1], ":")
	}

	if len(hp) != 2 {
		return nil, errors.New("need address in server address form host:port")
	}

	return hp, nil
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

	if len(hp) != 2 {
		return "", url, errors.New("need address in base url form host:port")
	}

	return protocol, hp, nil
}
