package main

import (
	"bytes"
	"fmt"
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/evgeniy-dammer/blockchain/network"
	"log"
	"math/rand"
	"strconv"
	"time"
)

func main() {
	// transports creating
	localTransport := network.NewLocalTransport("LOCAL")
	remoteTransportA := network.NewLocalTransport("REMOTE_A")
	remoteTransportB := network.NewLocalTransport("REMOTE_B")
	remoteTransportC := network.NewLocalTransport("REMOTE_C")

	// transports connecting
	localTransport.Connect(remoteTransportA)
	remoteTransportA.Connect(remoteTransportB)
	remoteTransportB.Connect(remoteTransportC)
	remoteTransportA.Connect(localTransport)

	initRemoteServers([]network.Transport{remoteTransportA, remoteTransportB, remoteTransportC})

	// message sending
	go func() {
		for {
			// remoteTransport.SendMessage(localTransport.Address(), []byte("Hello World!"))
			if err := sendTransaction(remoteTransportA, localTransport.Address()); err != nil {
				log.Printf("sending transaction fail: %s", err)
			}

			time.Sleep(2 * time.Second)
		}
	}()

	privateKey := crypto.GeneratePrivateKey()
	localServer := makeServer("LOCAL", localTransport, &privateKey)

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
		PrivateKey: privateKey,
		ID:         id,
		Transports: []network.Transport{transport},
	}

	server, err := network.NewServer(options)
	if err != nil {
		log.Fatal(err)
	}

	return server
}

func sendTransaction(transport network.Transport, address network.NetworkAddress) error {
	privKey := crypto.GeneratePrivateKey()
	data := []byte(strconv.FormatInt(int64(rand.Intn(1000)), 10))

	transaction := core.NewTransaction(data)
	transaction.Sign(privKey)

	buf := &bytes.Buffer{}

	if err := transaction.Encode(core.NewGobTransactionEncoder(buf)); err != nil {
		return err
	}

	message := network.NewMessage(network.MessageTypeTransaction, buf.Bytes())

	return transport.SendMessage(address, message.Bytes())
}
