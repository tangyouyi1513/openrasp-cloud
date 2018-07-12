package tools

import (
	"net"
)

func GetMacAddrs() (macs []string, err error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return
	}
	macs = make([]string, 0, len(interfaces))
	for _, inter := range interfaces {
		mac := inter.HardwareAddr
		if len(mac) != 0 {
			macs = append(macs, mac.String())
		}
	}
	return
}
