package config

import (
	"errors"
	"strconv"
	"strings"
)

type netAddressExit struct {
	Scheme string
	Host   string
	Port   int
}

func (a *netAddressExit) String() string {
	return a.Scheme + a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *netAddressExit) Set(s string) error {
	url := strings.Split(s, "://")
	protocol, host := getProtocolAndHost(url)

	hp := strings.Split(host, ":")

	if len(hp) != 2 {
		return errors.New("need address in b form host:port")
	}

	port, err := strconv.Atoi(hp[1])

	if err != nil {
		return err
	}

	a.Scheme = protocol
	a.Host = hp[0]
	a.Port = port

	return nil
}

func getProtocolAndHost(url []string) (string, string) {
	var protocol string
	var host string

	if len(url) != 2 {
		protocol = "http"
		host = url[0]
	} else {
		protocol = url[0]
		host = url[1]
	}

	return protocol, host
}
