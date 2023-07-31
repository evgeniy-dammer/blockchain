package main

import (
	"bytes"
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/evgeniy-dammer/blockchain/network"
	"log"
	"net/http"
	"time"
)

func main() {
	privKey := crypto.GeneratePrivateKey()

	localNode := makeServer("LOCAL_NODE", &privKey, ":3000", []string{":4000"}, ":9999")
	go localNode.Start()

	remoteNode := makeServer("REMOTE_NODE", nil, ":4000", []string{":7000"}, "")
	go remoteNode.Start()

	remoteNodeB := makeServer("REMOTE_NODE_B", nil, ":7000", nil, "")
	go remoteNodeB.Start()

	go func() {
		time.Sleep(11 * time.Second)

		// tcpTester()
		lateNode := makeServer("LATE_NODE", nil, ":6000", []string{":4000"}, "")
		go lateNode.Start()
	}()

	time.Sleep(1 * time.Second)

	txSendTicker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			txSender()

			<-txSendTicker.C
		}
	}()

	select {}
}

func makeServer(id string, privateKey *crypto.PrivateKey, addr string, seedNodes []string, apiListenAddr string) *network.Server {
	options := network.ServerOptions{
		APIListenAddr: apiListenAddr,
		SeedNodes:     seedNodes,
		ListenAddr:    addr,
		PrivateKey:    privateKey,
		ID:            id,
	}

	server, err := network.NewServer(options)
	if err != nil {
		log.Fatal(err)
	}

	return server
}

func txSender() {
	privKey := crypto.GeneratePrivateKey()
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}

	transaction := core.NewTransaction(data)
	transaction.Sign(privKey)

	buf := &bytes.Buffer{}

	if err := transaction.Encode(core.NewGobTransactionEncoder(buf)); err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "http://localhost:9999/tx", buf)
	if err != nil {
		panic(err)
	}

	client := http.Client{}

	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}
}
