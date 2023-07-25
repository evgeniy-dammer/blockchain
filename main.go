package main

import (
	"bytes"
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/evgeniy-dammer/blockchain/network"
	"log"
	"net"
)

func main() {
	privKey := crypto.GeneratePrivateKey()

	localNode := makeServer("LOCAL_NODE", &privKey, ":3000", []string{":4000"})
	go localNode.Start()

	remoteNode := makeServer("REMOTE_NODE", nil, ":4000", []string{":5000"})
	go remoteNode.Start()

	remoteNodeB := makeServer("REMOTE_NODE_B", nil, ":5000", nil)
	go remoteNodeB.Start()

	// time.Sleep(1 * time.Second)

	// tcpTester()

	select {}
}

func makeServer(id string, privateKey *crypto.PrivateKey, addr string, seedNodes []string) *network.Server {
	options := network.ServerOptions{
		SeedNodes:  seedNodes,
		ListenAddr: addr,
		PrivateKey: privateKey,
		ID:         id,
	}

	server, err := network.NewServer(options)
	if err != nil {
		log.Fatal(err)
	}

	return server
}

func tcpTester() {
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		panic(err)

	}

	privKey := crypto.GeneratePrivateKey()
	//data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}

	transaction := core.NewTransaction(data)
	transaction.Sign(privKey)

	buf := &bytes.Buffer{}

	if err = transaction.Encode(core.NewGobTransactionEncoder(buf)); err != nil {
		panic(err)
	}

	message := network.NewMessage(network.MessageTypeTransaction, buf.Bytes())

	_, err = conn.Write(message.Bytes())
	if err != nil {
		panic(err)
	}

}
