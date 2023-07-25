package network

import "net"

// NetworkAddress
type NetworkAddress string

// Transport interface
type Transport interface {
	Consume() <-chan RPC
	Connect(transport Transport) error
	SendMessage(to net.Addr, payload []byte) error
	Broadcast([]byte) error
	Address() net.Addr
}
