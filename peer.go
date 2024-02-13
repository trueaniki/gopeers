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
	Peers map[string]struct{}

	WriteChan chan []byte
	ReadChan  chan []byte
}

// Tries to dial the ip by TCP
func (p *Peer) tryConnect(ip string) {
	_, err := net.Dial("tcp", ip+":"+PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	p.m.Lock()
	p.Peers[ip+":"+PORT] = struct{}{}
	p.m.Unlock()
}

func (p *Peer) muxWrite() {
	for msg := range p.WriteChan {
		// Send the message to all peers
		for peer := range p.Peers {
			conn, err := net.Dial("tcp", peer)
			if err != nil {
				fmt.Println(err)
				continue
			}
			_, err = conn.Write(msg)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (p *Peer) muxRead() {
	fmt.Println("Waiting for connections")
	ln, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		panic(err)
	}
	for {
		// Accept new connections
		conn, err := ln.Accept()
		fmt.Println("New connection")
		if err != nil {
			fmt.Println(err)
			continue
		}
		// Add new peer to the set
		if _, ok := p.Peers[conn.RemoteAddr().String()]; !ok {
			p.m.Lock()
			p.Peers[conn.RemoteAddr().String()] = struct{}{}
			p.m.Unlock()
		}
		// Read from the peer
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			continue
		}
		// Pass the message to the channel
		p.ReadChan <- buf[:n]
	}
}

func (p *Peer) GracefulExit() {
	close(p.WriteChan)
	close(p.ReadChan)
}

func (p *Peer) Start() {
	go p.muxRead()
	go p.muxWrite()
}

func NewPeer(ipList []string) *Peer {
	peer := &Peer{}
	peer.WriteChan = make(chan []byte)
	peer.ReadChan = make(chan []byte)
	peer.Peers = make(map[string]struct{}, 10)

	wg := sync.WaitGroup{}
	wg.Add(len(ipList))
	for _, ip := range ipList {
		go func(ip string) {
			peer.tryConnect(ip)
			wg.Done()
		}(ip)
	}
	wg.Wait()

	return peer
}
