package gopeers

import (
	"fmt"
	"net"
)

func getNetworkInfo(cidr string) ([]string, string) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}
	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	myIp, err := getMyIPInNetwork(ipnet)
	if err != nil {
		panic(err)
	}

	// remove network address and broadcast address
	return ips[1 : len(ips)-1], myIp
}

func getMyIPInNetwork(network *net.IPNet) (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil && network.Contains(ipnet.IP) {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("local IP address not found")
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func deleteByValue(s []string, val string) []string {
	for i, v := range s {
		if v == val {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
