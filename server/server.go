package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"
	"unicode"

	"github.com/nakkamarra/tcp-chat-server/peers"
)

var ErrShutdown = errors.New("shutting down server")

type Message struct {
	ts       time.Time
	remoteIP net.Addr
	content  []byte
}

type Server struct {
	listener      net.Listener
	signalChannel chan os.Signal
	peers         *peers.List
	broadcast     chan Message
}

func New(listener net.Listener, sigChan chan os.Signal) *Server {
	return &Server{
		listener:      listener,
		signalChannel: sigChan,
		peers:         peers.NewList(),
		broadcast:     make(chan Message),
	}
}

func (s *Server) ListenAndServe() error {
	go func() {
		for {
			conn, connErr := s.listener.Accept()
			if connErr != nil {
				continue
			}
			go s.handleConnection(conn)
		}
	}()

	go func() {
		for m := range s.broadcast {
			go s.handleBroadcast(m)
		}
	}()

	<-s.signalChannel
	close(s.broadcast)
	return ErrShutdown
}

func (s *Server) handleBroadcast(m Message) {
	fmt.Fprintf(os.Stdout, "message was broadcasted by (%s) @ [%s]: %s", m.remoteIP, m.ts.Format(time.Kitchen), m.content)
	content := fmt.Sprintf("[%s] (%s) > %s", m.ts.Format(time.Kitchen), m.remoteIP, m.content)
	s.peers.WriteToPeers(m.remoteIP, []byte(content))
}

func (s *Server) handleConnection(c net.Conn) {
	defer c.Close()
	s.registerConnect(c)
	for {
		buf := make([]byte, 2<<12)
		reader := bufio.NewReader(c)
		bytesRead, readErr := reader.Read(buf)
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			fmt.Fprintf(os.Stdout, "failed to read from connection @ %s: %s\n", c.RemoteAddr(), readErr)
			continue
		}
		if bytesRead < 1 {
			fmt.Fprintf(os.Stdout, "read with no bytes from connection @ %s\n", c.RemoteAddr())
		}
		if bytesRead == 1 && unicode.IsSpace(rune(buf[0])) {
			continue
		}
		s.broadcast <- Message{
			ts:       time.Now(),
			remoteIP: c.RemoteAddr(),
			content:  buf,
		}
	}
	s.registerDisconnect(c)
}

func (s *Server) registerConnect(c net.Conn) {
	s.peers.Add(c)
	fmt.Fprintf(os.Stdout, "client connected: %s\n", c.RemoteAddr())
}

func (s *Server) registerDisconnect(c net.Conn) {
	s.peers.Remove(c)
	fmt.Fprintf(os.Stdout, "client disconnected: %s\n", c.RemoteAddr())
}
