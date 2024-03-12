package main

import (
	"time"

	"github.com/trueaniki/gopeers"
)

func main() {
	locals := gopeers.PingSweep("192.168.100.0/24")

	p := gopeers.NewPeer()
	defer p.GracefulExit()
	p.Discover(locals)
	p.Listen()

	p.WriteChan <- []byte("Hello, world!")
	time.Sleep(1 * time.Second)
}
