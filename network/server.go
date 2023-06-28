package network

import (
	"fmt"
	"time"
)

// ServerOptions
type ServerOptions struct {
	Transports []Transport
}

// Server
type Server struct {
	options ServerOptions
	rpcCh   chan RPC
	quitCh  chan struct{}
}

// NewServer is a constructor for the Server
func NewServer(options ServerOptions) *Server {
	return &Server{
		options: options,
		rpcCh:   make(chan RPC),
		quitCh:  make(chan struct{}, 1),
	}
}

// Start starts the Server
func (s *Server) Start() {
	s.initTransports()
	ticker := time.NewTicker(5 * time.Second)

LOOP:
	for {
		select {
		case rpc := <-s.rpcCh:
			fmt.Printf("%+v\n", rpc)
		case <-s.quitCh:
			break LOOP
		case <-ticker.C:
			fmt.Println("do stuff every X second")
		}
	}

	fmt.Println("Server shutdown...")
}

// initTransports initializes Transports
func (s *Server) initTransports() {
	for _, tr := range s.options.Transports {
		go func(tr Transport) {
			for rpc := range tr.Consume() {
				s.rpcCh <- rpc
			}
		}(tr)
	}
}
