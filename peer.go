package gopeers

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

const (
	PORT_S = "34759"
	PORT   = 34759
)

type Peer struct {
	m     sync.RWMutex
	Peers map[string]struct{}

	WriteChan chan []byte
	ReadChan  chan []byte

	laddr *net.TCPAddr
}

// Tries to dial the ip by TCP
func (p *Peer) tryConnect(ip string) {
	conn, err := net.DialTCP("tcp",
		nil,
		&net.TCPAddr{IP: net.ParseIP(ip), Port: 34759},
	)
	// _, err := net.Dial("tcp", ip+":"+PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	// Add the peer to the set
	p.m.Lock()
	p.Peers[ip+":"+PORT_S] = struct{}{}
	p.m.Unlock()
}

func (p *Peer) muxWrite() {
	for msg := range p.WriteChan {
		fmt.Println("Got message", string(msg))
		// Send the message to all peers
		for peer := range p.Peers {
			fmt.Println("Sending to", peer)
			conn, err := net.DialTCP("tcp",
				nil,
				&net.TCPAddr{IP: net.ParseIP(peer), Port: PORT},
			)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Sending: ", string(msg))
			_, err = conn.Write(msg)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (p *Peer) muxRead() {
	fmt.Println("Waiting for connections")
	ln, err := net.Listen("tcp", ":"+PORT_S)
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

		// Port can differ, so changing it to the standard one
		peerAddr := strings.Split(conn.RemoteAddr().String(), ":")[0] + ":" + PORT_S
		// Add new peer to the set
		if _, ok := p.Peers[peerAddr]; !ok {
			p.m.Lock()
			p.Peers[conn.RemoteAddr().String()] = struct{}{}
			p.m.Unlock()
		}
		// Read from the peer
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			continue
		}
		// Pass the message to the channel
		p.ReadChan <- buf[:n]
		conn.Close()
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

	peer.laddr = &net.TCPAddr{IP: nil, Port: 34759}

	wg := sync.WaitGroup{}
	wg.Add(len(ipList))
	// Find devices that are using gopeers
	for _, ip := range ipList {
		go func(ip string) {
			peer.tryConnect(ip)
			wg.Done()
		}(ip)
	}
	wg.Wait()

	return peer
}
