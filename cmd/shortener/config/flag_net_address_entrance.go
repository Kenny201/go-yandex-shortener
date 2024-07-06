package config

import (
	"errors"
	"strconv"
	"strings"
)

type netAddressEntrance struct {
	Host string
	Port int
}

func (a *netAddressEntrance) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *netAddressEntrance) Set(s string) error {
	host := strings.Split(s, "://")
	var hp []string

	if len(host) != 2 {
		hp = strings.Split(host[0], ":")
	} else {
		hp = strings.Split(host[1], ":")
	}

	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}

	port, err := strconv.Atoi(hp[1])

	if err != nil {
		return err
	}

	a.Host = hp[0]
	a.Port = port

	return nil
}
