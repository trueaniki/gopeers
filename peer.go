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

func (p *Peer) muxWrite(conn net.Conn) {
	for {
		select {
		case msg := <-p.WriteChan:
			fmt.Println("Sending message")
			_, err := conn.Write(msg)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (p *Peer) muxRead(conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			continue
		}
		p.ReadChan <- buf[:n]
	}
}

func (p *Peer) GracefulExit() {
	p.m.Lock()
	for _, conn := range p.Conns {
		conn.Close()
	}
	p.m.Unlock()
	close(p.WriteChan)
	close(p.ReadChan)
}

func (p *Peer) Start() {
	for _, conn := range p.Conns {
		go p.muxRead(conn)
		go p.muxWrite(conn)
	}
}

func NewPeer(ipList []string) *Peer {
	peer := &Peer{}
	peer.WriteChan = make(chan []byte)
	peer.ReadChan = make(chan []byte)
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
