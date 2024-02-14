package gopeers

import (
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func PingSweep(network string) []string {
	ips, myIp := getNetworkInfo(network)
	m := sync.Mutex{}
	var alive []string

	wg := new(sync.WaitGroup)
	wg.Add(len(ips))
	for _, ip := range ips {
		go func(ip string) {
			if ping(ip) {
				m.Lock()
				alive = append(alive, ip)
				m.Unlock()
			}
			wg.Done()
		}(ip)
	}
	wg.Wait()
	return deleteByValue(alive, myIp)
}

func ping(ip string) bool {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: 1, Seq: 1,
			Data: []byte("HELLO-R-U-THERE"),
		},
	}

	wb, err := wm.Marshal(nil)
	if err != nil {
		panic(err)
	}

	c.SetReadDeadline(time.Now().Add(1 * time.Second))
	if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(ip)}); err != nil {
		return false
	}

	rb := make([]byte, 1500)
	for {
		n, addr, err := c.ReadFrom(rb)
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				return false
			}
			panic(err)
		}
		if addr.String() != ip {
			continue
		}

		rm, err := icmp.ParseMessage(1, rb[:n])
		if err != nil {
			panic(err)
		}

		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			// fmt.Printf("got reflection from %v", rm)
			return true
		default:
			fmt.Printf("got %+v; want echo reply", rm)
			return false
		}
	}
}

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
