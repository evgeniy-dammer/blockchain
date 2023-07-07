package network

import (
	"bytes"
	"fmt"
	"sync"
)

// LocalTransport
type LocalTransport struct {
	address   NetworkAddress
	consumeCh chan RPC
	lock      sync.RWMutex
	peers     map[NetworkAddress]*LocalTransport
}

// NewLocalTransport constructor for the NewLocalTransport
func NewLocalTransport(address NetworkAddress) Transport {
	return &LocalTransport{
		address:   address,
		consumeCh: make(chan RPC, 1024),
		peers:     make(map[NetworkAddress]*LocalTransport),
	}
}

// Consume returns transports consume channel
func (t *LocalTransport) Consume() <-chan RPC {
	return t.consumeCh
}

// Connect connects one transport to another
func (t *LocalTransport) Connect(transport Transport) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.peers[transport.Address()] = transport.(*LocalTransport)

	return nil
}

// SendMessage sends message from one transport to another
func (t *LocalTransport) SendMessage(to NetworkAddress, payload []byte) error {
	t.lock.RLock()
	defer t.lock.RUnlock()

	peer, ok := t.peers[to]

	if !ok {
		return fmt.Errorf("%s: could not send message to %s", t.address, to)
	}

	peer.consumeCh <- RPC{
		From:    t.address,
		Payload: bytes.NewReader(payload),
	}

	return nil
}

// Address returns transports network address
func (t *LocalTransport) Address() NetworkAddress {
	return t.address
}
