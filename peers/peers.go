package peers

import (
	"fmt"
	"net"
	"os"
	"sync"
)

type List struct {
	pl map[net.Addr]net.Conn
	mu sync.RWMutex
}

func NewList() *List {
	return &List{
		pl: make(map[net.Addr]net.Conn),
		mu: sync.RWMutex{},
	}
}

func (l *List) Add(peer net.Conn) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pl[peer.RemoteAddr()] = peer
}

func (l *List) Remove(peer net.Conn) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.pl, peer.RemoteAddr())
}

func (l *List) WriteToPeers(from net.Addr, content []byte) {
	l.mu.RLock()
	for addr, peer := range l.pl {
		if addr == from {
			continue
		}
		bytesWritten, writeErr := peer.Write(content)
		if writeErr != nil {
			fmt.Fprintf(os.Stdout, "failed to write to peer %s: %s", peer.RemoteAddr(), writeErr)
			continue
		}
		if bytesWritten < 1 {
			fmt.Fprintf(os.Stdout, "failed to write to peer %s: %s", peer.RemoteAddr(), writeErr)
			continue
		}
	}
	l.mu.RUnlock()
}
