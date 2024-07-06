package config

import (
	"strconv"
)

type netAddressEntrance struct {
	Host string
	Port int
}

func (a *netAddressEntrance) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *netAddressEntrance) Set(s string) error {
	hp, err := ParseServerAddress(s)

	if err != nil {
		return err
	}

	a.Host = hp[0]
	a.Port, err = strconv.Atoi(hp[1])

	if err != nil {
		return err
	}

	return nil
}
