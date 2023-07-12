package util

import (
	"math/rand"
	"testing"
	"time"

	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/evgeniy-dammer/blockchain/types"
	"github.com/stretchr/testify/assert"
)

// RandomBytes returns a random slice of bytes
func RandomBytes(size int) []byte {
	token := make([]byte, size)
	rand.Read(token)
	return token
}

// RandomHash returns a random hash
func RandomHash() types.Hash {
	return types.HashFromBytes(RandomBytes(32))
}

// NewRandomTransaction returns a new random transaction without signature
func NewRandomTransaction(size int) *core.Transaction {
	return core.NewTransaction(RandomBytes(size))
}

// NewRandomTransactionWithSignature returns a new random transaction with signature.
func NewRandomTransactionWithSignature(t *testing.T, privKey crypto.PrivateKey, size int) *core.Transaction {
	tx := NewRandomTransaction(size)
	assert.Nil(t, tx.Sign(privKey))
	return tx
}

// NewRandomBlock returns a random block without signature
func NewRandomBlock(t *testing.T, height uint32, prevBlockHash types.Hash) *core.Block {
	txSigner := crypto.GeneratePrivateKey()
	tx := NewRandomTransactionWithSignature(t, txSigner, 100)
	header := &core.Header{
		Version:           1,
		PreviousBlockHash: prevBlockHash,
		Height:            height,
		Timestamp:         time.Now().UnixNano(),
	}
	b, err := core.NewBlock(header, []*core.Transaction{tx})
	assert.Nil(t, err)
	dataHash, err := core.CalculateDataHash(b.Transactions)
	assert.Nil(t, err)
	b.Header.DataHash = dataHash

	return b
}

// NewRandomBlockWithSignature returns a new random block without signature
func NewRandomBlockWithSignature(t *testing.T, pk crypto.PrivateKey, height uint32, prevHash types.Hash) *core.Block {
	b := NewRandomBlock(t, height, prevHash)
	assert.Nil(t, b.Sign(pk))

	return b
}
