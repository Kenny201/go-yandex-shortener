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
	if serverAddr := os.Getenv("SERVER_ADDRESS"); serverAddr != "" {
		hp, err := ParseServerAddress(serverAddr)

		if err != nil {
			return err
		}

		Args.NetAddressEntrance.Host = hp[0]
		Args.NetAddressEntrance.Port, err = strconv.Atoi(hp[1])

		if err != nil {
			return err
		}
	}

	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		protocol, hp, err := ParseBaseURL(baseURL)

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
