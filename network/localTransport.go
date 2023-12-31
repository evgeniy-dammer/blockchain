package network

import (
	"bytes"
	"fmt"
	"net"
	"sync"
)

// LocalTransport
type LocalTransport struct {
	address   net.Addr
	consumeCh chan RPC
	lock      sync.RWMutex
	peers     map[net.Addr]*LocalTransport
}

// NewLocalTransport constructor for the NewLocalTransport
func NewLocalTransport(address net.Addr) Transport {
	return &LocalTransport{
		address:   address,
		consumeCh: make(chan RPC, 1024),
		peers:     make(map[net.Addr]*LocalTransport),
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
func (t *LocalTransport) SendMessage(to net.Addr, payload []byte) error {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if t.address == to {
		return nil
	}

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

// Broadcast sends message to all peers
func (t *LocalTransport) Broadcast(payload []byte) error {
	for _, peer := range t.peers {
		if err := t.SendMessage(peer.address, payload); err != nil {
			return err
		}
	}

	return nil
}

// Address returns transports network address
func (t *LocalTransport) Address() net.Addr {
	return t.address
}
