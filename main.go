package main

import (
	"github.com/evgeniy-dammer/blockchain/network"
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
			remoteTransport.SendMessage(localTransport.Address(), []byte("Hello World!"))
			time.Sleep(1 * time.Second)
		}
	}()

	options := network.ServerOptions{Transports: []network.Transport{localTransport}}

	// creating and starting server
	server := network.NewServer(options)
	server.Start()
}
