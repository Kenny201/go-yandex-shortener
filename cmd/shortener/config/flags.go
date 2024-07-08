package config

import (
	"flag"
	"os"
	"strconv"
)

var Args struct {
	ServerAddress string
	BaseURL       string
}

func ParseFlags() error {
	flag.StringVar(&Args.ServerAddress, "a", ":8080", "server address host:port")
	flag.StringVar(&Args.BaseURL, "b", "http://localhost:8080", "Result net address host:port")
	flag.Parse()

	err := setArgsFromEnv()

	if err != nil {
		return err
	}

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
