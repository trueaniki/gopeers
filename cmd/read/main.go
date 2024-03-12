package main

import (
	"fmt"

	"github.com/trueaniki/gopeers"
)

func main() {
	locals := gopeers.PingSweep("192.168.100.0/24")

	p := gopeers.NewPeer()
	defer p.GracefulExit()
	p.Discover(locals)
	p.Listen()

	data := <-p.ReadChan
	fmt.Println(string(data))
}
