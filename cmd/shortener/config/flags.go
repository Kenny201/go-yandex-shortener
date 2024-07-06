package config

import (
	"flag"
	"os"
	"strconv"
)

var Args struct {
	NetAddressEntrance netAddressEntrance
	NetAddressExit     netAddressExit
}

func ParseFlags() error {
	_ = flag.Value(&Args.NetAddressEntrance)
	_ = flag.Value(&Args.NetAddressExit)

	flag.Var(&Args.NetAddressEntrance, "a", "Net address host:port")
	flag.Var(&Args.NetAddressExit, "b", "Result net address host:port")
	flag.Parse()

	err := setArgsFromEnv()

	if err != nil {
		return err
	}

	setDefaultValueEntrance()
	setDefaultValueOnExit()

	return nil
}

func setArgsFromEnv() error {
	if envServerAddr := os.Getenv("SERVER_ADDRESS"); envServerAddr != "" {
		hp, err := ParseServerAddress(envServerAddr)

		if err != nil {
			return err
		}

		Args.NetAddressEntrance.Host = hp[0]
		Args.NetAddressEntrance.Port, err = strconv.Atoi(hp[1])

		if err != nil {
			return err
		}
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		protocol, hp, err := ParseBaseURL(envBaseURL)

		if err != nil {
			return err
		}

		Args.NetAddressExit.Scheme = protocol
		Args.NetAddressExit.Host = hp[0]
		Args.NetAddressExit.Port, err = strconv.Atoi(hp[1])

		if err != nil {
			return err
		}
	}

	return nil
}

func setDefaultValueEntrance() {
	if Args.NetAddressEntrance.Port == 0 && Args.NetAddressEntrance.Host == "" {
		Args.NetAddressEntrance.Host = "localhost"
		Args.NetAddressEntrance.Port = 8080
	}
}

func setDefaultValueOnExit() {
	if Args.NetAddressExit.Port == 0 && Args.NetAddressExit.Host == "" {
		Args.NetAddressExit.Scheme = "http"
		Args.NetAddressExit.Host = "localhost"
		Args.NetAddressExit.Port = 8080
	}
}
