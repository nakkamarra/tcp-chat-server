package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nakkamarra/tcp-chat-server/server"
)

const defaultPort = ":8080"

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	listener, listenerErr := net.Listen("tcp", defaultPort)
	if listenerErr != nil {
		log.Fatalf("couldnt listen on port %s: %s\n", defaultPort, listenerErr)
	}

	s := server.New(listener, sigChan)
	log.Printf("listening on port %s...\n", defaultPort)
	log.Fatalf("%s\n", s.ListenAndServe())
}
