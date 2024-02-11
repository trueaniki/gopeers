package gopeers

import (
	"net"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func PingSweep() {

}

func ping(ip string) {
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

	if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP("8.8.8.8")}); err != nil {
		panic(err)
	}

	rb := make([]byte, 1500)
	n, _, err := c.ReadFrom(rb)
	if err != nil {
		panic(err)
	}

	rm, err := icmp.ParseMessage(1, rb[:n])
	if err != nil {
		panic(err)
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		println("got reflection from %v", rm)
	default:
		println("got %+v; want echo reply", rm)
	}
}
