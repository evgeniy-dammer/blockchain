package core

import (
	"bytes"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
	assert.Equal(t, &tx, txDecoded)
}

func randomTransactionWithSignature(t *testing.T) Transaction {
	privateKey := crypto.GeneratePrivateKey()

	tx := Transaction{
		Data: []byte("foo"),
	}

	assert.Nil(t, tx.Sign(privateKey))

	return tx
}
