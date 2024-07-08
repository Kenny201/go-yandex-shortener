package valueobject

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ServerAddress struct {
	scheme string
	host   string
	port   int
}

func NewServerAddress(host string) (ServerAddress, error) {
	scheme, hp, err := ParseServerAddress(host)

	if err != nil {
		return ServerAddress{}, err
	}

	port, err := strconv.Atoi(hp[1])

	if err != nil {
		return ServerAddress{}, err
	}

	return ServerAddress{scheme, hp[0], port}, nil
}

func (sa ServerAddress) ToString() string {
	return fmt.Sprintf("%s:%d", sa.host, sa.port)
}

func (sa ServerAddress) Host() string {
	return sa.host
}

func (sa ServerAddress) Port() int {
	return sa.port
}

func ParseServerAddress(s string) (string, []string, error) {
	host := strings.Split(s, "://")
	var hp []string
	var scheme string

	if len(host) != 2 {
		scheme = "http"
		hp = strings.Split(host[0], ":")
	} else {
		scheme = host[0]
		hp = strings.Split(host[1], ":")
	}

	if len(hp) != 2 {
		return "", nil, errors.New("need address in server address form host:port")
	}

	return scheme, hp, nil
}
