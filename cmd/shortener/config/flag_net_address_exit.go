package config

import (
	"strconv"
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
	protocol, hp, err := ParseBaseURL(s)

	if err != nil {
		return err
	}

	a.Scheme = protocol
	a.Host = hp[0]
	a.Port, err = strconv.Atoi(hp[1])

	if err != nil {
		return err
	}

	return nil
}
