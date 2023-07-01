package core

import (
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
	transaction.PublicKey = otherPrivateKey.PublicKey()

	assert.NotNil(t, transaction.Verify())
}
