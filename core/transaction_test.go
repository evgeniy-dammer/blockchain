package core

import (
	"bytes"
	"encoding/gob"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNFTTransaction(t *testing.T) {
	collectionTx := CollectionTx{
		Fee:      200,
		MetaData: []byte("The beginning of a new collection"),
	}

	privKey := crypto.GeneratePrivateKey()
	tx := &Transaction{
		Type:    TxTypeCollection,
		TxInner: collectionTx,
	}
	tx.Sign(privKey)

	buf := new(bytes.Buffer)
	assert.Nil(t, gob.NewEncoder(buf).Encode(tx))

	txDecoded := &Transaction{}
	assert.Nil(t, gob.NewDecoder(buf).Decode(txDecoded))
	assert.Equal(t, tx, txDecoded)
}

func TestTransaction_Sign(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	transaction := &Transaction{
		Data: []byte("foo"),
	}

	assert.Nil(t, transaction.Sign(privateKey))
	assert.NotNil(t, transaction.Signature)
}

func TestTransaction_Verify(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	transaction := &Transaction{
		Data: []byte("foo"),
	}

	assert.Nil(t, transaction.Sign(privateKey))
	assert.Nil(t, transaction.Verify())

	otherPrivateKey := crypto.GeneratePrivateKey()
	transaction.From = otherPrivateKey.PublicKey()

	assert.NotNil(t, transaction.Verify())
}

func TestTransaction_EncodeDecode(t *testing.T) {
	tx := randomTransactionWithSignature(t)
	buf := &bytes.Buffer{}

	assert.Nil(t, tx.Encode(NewGobTransactionEncoder(buf)))

	txDecoded := new(Transaction)

	assert.Nil(t, txDecoded.Decode(NewGobTransactionDecoder(buf)))
	assert.Equal(t, tx, txDecoded)
}

func randomTransactionWithSignature(t *testing.T) *Transaction {
	privateKey := crypto.GeneratePrivateKey()

	tx := Transaction{
		Data: []byte("foo"),
	}

	assert.Nil(t, tx.Sign(privateKey))

	return &tx
}
