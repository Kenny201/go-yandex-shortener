package config

import (
	"flag"
)

var Args struct {
	NetAddressEntrance netAddressEntrance
	NetAddressExit     netAddressExit
}

func ParseFlags() {
	_ = flag.Value(&Args.NetAddressEntrance)
	_ = flag.Value(&Args.NetAddressExit)

	flag.Var(&Args.NetAddressEntrance, "a", "Net address host:port")
	flag.Var(&Args.NetAddressExit, "b", "Result net address host:port")
	flag.Parse()

	setDefaultValueOnEntrance()
	setDefaultValueOnExit()
}

func setDefaultValueOnEntrance() {
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
