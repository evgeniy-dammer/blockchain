package network

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnect(t *testing.T) {
	tra := NewLocalTransport("A").(*LocalTransport)
	trb := NewLocalTransport("B").(*LocalTransport)

	tra.Connect(trb)
	trb.Connect(tra)

	assert.Equal(t, tra.peers[trb.address], trb)
	assert.Equal(t, trb.peers[tra.address], tra)
}

func TestSendMessage(t *testing.T) {
	tra := NewLocalTransport("A").(*LocalTransport)
	trb := NewLocalTransport("B").(*LocalTransport)

	tra.Connect(trb)
	trb.Connect(tra)

	msg := []byte("Hello World!")

	assert.Nil(t, tra.SendMessage(trb.Address(), msg))

	rpc := <-trb.Consume()

	buf := make([]byte, len(msg))

	n, err := rpc.Payload.Read(buf)

	assert.Nil(t, err)
	assert.Equal(t, n, len(msg))

	assert.Equal(t, buf, msg)
	assert.Equal(t, rpc.From, tra.address)
}
