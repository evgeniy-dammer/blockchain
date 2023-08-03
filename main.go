package main

import (
	"bytes"
	"encoding/json"
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/evgeniy-dammer/blockchain/network"
	"github.com/evgeniy-dammer/blockchain/types"
	"github.com/evgeniy-dammer/blockchain/util"
	"log"
	"net/http"
	"time"
)

func main() {
	validatorPrivKey := crypto.GeneratePrivateKey()

	localNode := makeServer("LOCAL_NODE", &validatorPrivKey, ":3000", []string{":4000"}, ":9999")
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

	if err := sendTransaction(validatorPrivKey); err != nil {
		panic(err)
	}

	/*collectionOwnerPrivKey := crypto.GeneratePrivateKey()
	collectionHash := createCollectionTx(collectionOwnerPrivKey)

	txSendTicker := time.NewTicker(1 * time.Second)
	go func() {
		for i := 0; i < 20; i++ {
			nftMinter(collectionOwnerPrivKey, collectionHash)

			<-txSendTicker.C
		}
	}()*/

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

func sendTransaction(privKey crypto.PrivateKey) error {
	toPrivKey := crypto.GeneratePrivateKey()

	tx := core.NewTransaction(nil)
	tx.To = toPrivKey.PublicKey()
	tx.Value = 666

	if err := tx.Sign(privKey); err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTransactionEncoder(buf)); err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "http://localhost:9999/tx", buf)
	if err != nil {
		panic(err)
	}

	client := http.Client{}
	_, err = client.Do(req)

	return err
}

func createCollectionTx(privKey crypto.PrivateKey) types.Hash {
	transaction := core.NewTransaction(nil)
	transaction.TxInner = core.CollectionTx{
		Fee:      200,
		MetaData: []byte("chicken and egg collection!"),
	}

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

	return transaction.Hash(core.TransactionHasher{})
}

func nftMinter(privKey crypto.PrivateKey, collection types.Hash) {
	metaData := map[string]any{
		"power":  8,
		"health": 100,
		"color":  "green",
		"rare":   "yes",
	}

	metaBuf := new(bytes.Buffer)
	if err := json.NewEncoder(metaBuf).Encode(metaData); err != nil {
		panic(err)
	}

	tx := core.NewTransaction(nil)
	tx.TxInner = core.MintTx{
		Fee:             200,
		NFT:             util.RandomHash(),
		MetaData:        metaBuf.Bytes(),
		Collection:      collection,
		CollectionOwner: privKey.PublicKey(),
	}
	tx.Sign(privKey)

	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTransactionEncoder(buf)); err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "http://localhost:9000/tx", buf)
	if err != nil {
		panic(err)
	}

	client := http.Client{}
	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}
}
