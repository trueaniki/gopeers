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
