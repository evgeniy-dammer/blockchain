package main

import (
	"bytes"
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/evgeniy-dammer/blockchain/network"
	"github.com/rs/zerolog/log"
	"math/rand"
	"strconv"
	"time"
)

func main() {
	// transports creating
	localTransport := network.NewLocalTransport("LOCAL")
	remoteTransport := network.NewLocalTransport("REMOTE")

	// transports connecting
	localTransport.Connect(remoteTransport)
	remoteTransport.Connect(localTransport)

	// message sending
	go func() {
		for {
			// remoteTransport.SendMessage(localTransport.Address(), []byte("Hello World!"))
			if err := sendTransaction(remoteTransport, localTransport.Address()); err != nil {
				log.Error().Msgf("sending transaction fail: %s", err)
			}

			time.Sleep(1 * time.Second)
		}
	}()

	privateKey := crypto.GeneratePrivateKey()

	options := network.ServerOptions{
		PrivateKey: &privateKey,
		ID:         "LOCAL",
		Transports: []network.Transport{localTransport},
	}

	// creating and starting server
	server, err := network.NewServer(options)
	if err != nil {
		log.Fatal().Msgf("server error: %s", err)
	}

	server.Start()
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
