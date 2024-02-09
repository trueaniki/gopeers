package gopeers

import (
	"fmt"
	"net"
	"sync"
)

const (
	PORT = "34759"
)

type Peer struct {
	m     sync.RWMutex
	Conns []net.Conn

	WriteChan chan []byte
	ReadChan  chan []byte
}

// Tries to dial the ip by TCP
func (p *Peer) tryConnect(ip string) {
	conn, err := net.Dial("tcp", ip+":"+PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	p.m.Lock()
	p.Conns = append(p.Conns, conn)
	p.m.Unlock()
}

func (p *Peer) listen() {
	ln, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		p.m.Lock()
		p.Conns = append(p.Conns, conn)
		p.m.Unlock()
	}
}

func (p *Peer) muxWrite() {
	for {
		select {
		case msg := <-p.WriteChan:
			fmt.Println(len(p.Conns))
			for _, conn := range p.Conns {
				fmt.Println("Sending message")
				_, err := conn.Write(msg)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}

func (p *Peer) muxRead() {
	for {
		for _, conn := range p.Conns {
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				continue
			}
			p.ReadChan <- buf[:n]
		}
	}
}

func (p *Peer) GracefulExit() {
	for _, conn := range p.Conns {
		conn.Close()
	}
}

func (p *Peer) Start() {
	go p.muxWrite()
	go p.muxRead()
}

func NewPeer(ipList []string) *Peer {
	peer := &Peer{}

	wg := sync.WaitGroup{}
	wg.Add(len(ipList))
	for _, ip := range ipList {
		go func(ip string) {
			peer.tryConnect(ip)
			wg.Done()
		}(ip)
	}
	wg.Wait()

	go peer.listen()

	return peer
}
