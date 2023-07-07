package network

// NetworkAddress
type NetworkAddress string

// Transport interface
type Transport interface {
	Consume() <-chan RPC
	Connect(transport Transport) error
	SendMessage(to NetworkAddress, payload []byte) error
	Address() NetworkAddress
}
