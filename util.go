package gopeers

import (
	"fmt"
	"net"
	"os"
)

func getLanNetwork() (*net.IPNet, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet, nil
			}
		}
	}
	return nil, fmt.Errorf("no network found")
}

func getMyIp(ipNet *net.IPNet) string {
	return ipNet.IP.To4().String()
}

func getBroadcastIp(ipNet *net.IPNet) string {
	ip := ipNet.IP.To4()
	mask := ipNet.Mask

	broadcast := net.IPv4(0, 0, 0, 0).To4()
	for i := range ip {
		broadcast[i] = ip[i] | ^mask[i]
	}

	return broadcast.String()
}
