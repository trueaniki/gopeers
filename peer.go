package gopeers

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"go.uber.org/zap"
)

const (
	PORT      = 34759
	HELLO_MSG = "HELLO"
)

type Peer struct {
	m     sync.RWMutex
	Peers map[string]struct{}

	WriteChan chan []byte
	ReadChan  chan []byte

	laddr *net.TCPAddr

	L *zap.Logger
}

// Tries to dial the ip by TCP
func (p *Peer) tryConnect(ip string) {
	conn, err := net.DialTCP("tcp",
		nil,
		&net.TCPAddr{IP: net.ParseIP(ip), Port: PORT},
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(HELLO_MSG))
	if err != nil {
		fmt.Println(err)
		return
	}

	// Add the peer to the set
	p.m.Lock()
	p.Peers[ip] = struct{}{}
	p.m.Unlock()
}

func (p *Peer) muxWrite() {
	for msg := range p.WriteChan {
		p.L.Info("Got message", zap.String("msg", string(msg)))
		// Send the message to all peers
		for peer := range p.Peers {
			p.L.Info("Sending to", zap.String("peer", peer))
			conn, err := net.DialTCP("tcp",
				nil,
				&net.TCPAddr{IP: net.ParseIP(peer), Port: PORT},
			)
			if err != nil {
				p.L.Error("Error while dialing", zap.String("peer", peer), zap.Error(err))
				continue
			}
			p.L.Info("Writing: ", zap.String("msg", string(msg)))
			_, err = conn.Write(msg)
			if err != nil {
				fmt.Println(err)
			}
			conn.Close()
		}
	}
}

func (p *Peer) muxRead() {
	p.L.Info("Waiting for connections...")
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(PORT))
	if err != nil {
		panic(err)
	}
	for {
		// Accept new connections
		conn, err := ln.Accept()
		p.L.Info("New connection")
		if err != nil {
			fmt.Println(err)
			continue
		}

		peerAddr := strings.Split(conn.RemoteAddr().String(), ":")[0]
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
		if string(buf[:n]) != HELLO_MSG {
			p.ReadChan <- buf[:n]
		}
		conn.Close()
	}
}

func (p *Peer) GracefulExit() {
	close(p.WriteChan)
	close(p.ReadChan)
}

func NewPeer() *Peer {
	peer := &Peer{}
	peer.WriteChan = make(chan []byte)
	peer.ReadChan = make(chan []byte)
	peer.Peers = make(map[string]struct{}, 10)

	peer.laddr = &net.TCPAddr{IP: nil, Port: 34759}

	peer.L = zap.NewNop()

	return peer
}

func (p *Peer) Discover(ipList []string) {
	wg := sync.WaitGroup{}
	wg.Add(len(ipList))
	// Find devices that are using gopeers
	for _, ip := range ipList {
		go func(ip string) {
			p.tryConnect(ip)
			wg.Done()
		}(ip)
	}
	wg.Wait()
}

func (p *Peer) ListPeers() []string {
	peers := make([]string, 0, len(p.Peers))
	for peer := range p.Peers {
		peers = append(peers, peer)
	}
	return peers
}

func (p *Peer) SetLogger(l *zap.Logger) {
	p.L = l
}

func (p *Peer) Listen() {
	go p.muxRead()
	go p.muxWrite()
}
