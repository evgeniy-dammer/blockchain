package network

/*func TestLocalTransport_Connect(t *testing.T) {
	tra := NewLocalTransport("A").(*LocalTransport)
	trb := NewLocalTransport("B").(*LocalTransport)

	tra.Connect(trb)
	trb.Connect(tra)

	assert.Equal(t, tra.peers[trb.address], trb)
	assert.Equal(t, trb.peers[tra.address], tra)
}

func TestLocalTransport_SendMessage(t *testing.T) {
	tra := NewLocalTransport("A").(*LocalTransport)
	trb := NewLocalTransport("B").(*LocalTransport)

	tra.Connect(trb)
	trb.Connect(tra)

	msg := []byte("Hello World!")

	assert.Nil(t, tra.SendMessage(trb.Address(), msg))

	rpc := <-trb.Consume()

	b, err := io.ReadAll(rpc.Payload)

	assert.Nil(t, err)
	assert.Equal(t, b, msg)
	assert.Equal(t, rpc.From, tra.address)
}

func TestLocalTransport_Broadcast(t *testing.T) {
	tra := NewLocalTransport("A").(*LocalTransport)
	trb := NewLocalTransport("B").(*LocalTransport)
	trc := NewLocalTransport("C").(*LocalTransport)

	tra.Connect(trb)
	tra.Connect(trc)

	msg := []byte("Hello World!")

	assert.Nil(t, tra.Broadcast(msg))

	rpcb := <-trb.Consume()

	b, err := io.ReadAll(rpcb.Payload)
	assert.Nil(t, err)
	assert.Equal(t, b, msg)

	rpcc := <-trc.Consume()

	c, err := io.ReadAll(rpcc.Payload)
	assert.Nil(t, err)
	assert.Equal(t, c, msg)
}*/
