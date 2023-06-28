package network

// NetworkAddress
type NetworkAddress string

// RPC
type RPC struct {
	From    NetworkAddress
	Payload []byte
}

// Transport interface
type Transport interface {
	Consume() <-chan RPC
	Connect(transport Transport) error
	SendMessage(to NetworkAddress, payload []byte) error
	Address() NetworkAddress
}
