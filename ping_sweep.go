package gopeers

import (
	"net"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// PingSweep returns a slice of strings containing the IP addresses of the
// hosts that are alive in the network. It uses ICMP echo requests to check
// if the hosts are alive.
// The network parameter must be in the CIDR notation, for example "192.168.0.0/24".
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
		return false
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
			return false
		}
		if addr.String() != ip {
			continue
		}

		rm, err := icmp.ParseMessage(1, rb[:n])
		if err != nil {
			return false
		}

		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			return true
		default:
			return false
		}
	}
}
