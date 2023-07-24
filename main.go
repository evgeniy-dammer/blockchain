package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/evgeniy-dammer/blockchain/network"
	"log"
	"time"
)

var transports = []network.Transport{
	network.NewLocalTransport("LOCAL"),
	// network.NewLocalTransport("REMOTE_B"),
	// network.NewLocalTransport("REMOTE_C"),
}

func main() {
	initRemoteServers(transports)
	localNode := transports[0]
	trLate := network.NewLocalTransport("LATE_NODE")
	// remoteNodeA := transports[1]
	// remoteNodeC := transports[3]

	go func() {
		time.Sleep(7 * time.Second)
		lateServer := makeServer(string(trLate.Address()), trLate, nil)
		go lateServer.Start()
	}()

	privKey := crypto.GeneratePrivateKey()

	localServer := makeServer("LOCAL", localNode, &privKey)
	localServer.Start()
}

func initRemoteServers(transports []network.Transport) {
	for i := 0; i < len(transports); i++ {
		id := fmt.Sprintf("REMOTE_%d", i)
		s := makeServer(id, transports[i], nil)

		go s.Start()
	}
}

func makeServer(id string, transport network.Transport, privateKey *crypto.PrivateKey) *network.Server {
	options := network.ServerOptions{
		Transport:  transport,
		PrivateKey: privateKey,
		ID:         id,
		Transports: transports,
	}

	server, err := network.NewServer(options)
	if err != nil {
		log.Fatal(err)
	}

	return server
}

func sendGetStatusMessage(tr network.Transport, to network.NetworkAddress) error {
	var (
		getStatusMsg = new(network.GetStatusMessage)
		buf          = new(bytes.Buffer)
	)

	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}
	msg := network.NewMessage(network.MessageTypeGetStatus, buf.Bytes())

	return tr.SendMessage(to, msg.Bytes())
}

func sendTransaction(transport network.Transport, address network.NetworkAddress) error {
	privKey := crypto.GeneratePrivateKey()
	//data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}

	transaction := core.NewTransaction(data)
	transaction.Sign(privKey)

	buf := &bytes.Buffer{}

	if err := transaction.Encode(core.NewGobTransactionEncoder(buf)); err != nil {
		return err
	}

	message := network.NewMessage(network.MessageTypeTransaction, buf.Bytes())

	return transport.SendMessage(address, message.Bytes())
}
