package core

import (
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/evgeniy-dammer/blockchain/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// randomBlock returns a random Block
func randomBlock(t *testing.T, height uint32, prevBlockHash types.Hash) *Block {
	privateKey := crypto.GeneratePrivateKey()
	tx := randomTransactionWithSignature(t)
	header := &Header{
		Version:           1,
		PreviousBlockHash: prevBlockHash,
		Timestamp:         time.Now().UnixNano(),
		Height:            height,
	}

	b := NewBlock(header, []Transaction{tx})

	dataHash, err := CalculateDataHash(b.Transactions)
	assert.Nil(t, err)

	b.Header.DataHash = dataHash
	assert.Nil(t, b.Sign(privateKey))

	return b
}

func TestBlock_Sign(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	block := randomBlock(t, 0, types.Hash{})

	assert.Nil(t, block.Sign(privateKey))
	assert.NotNil(t, block.Signature)
}

func TestBlock_Verify(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	block := randomBlock(t, 0, types.Hash{})

	assert.Nil(t, block.Sign(privateKey))
	assert.Nil(t, block.Verify())

	otherPrivateKey := crypto.GeneratePrivateKey()
	block.Validator = otherPrivateKey.PublicKey()

	assert.NotNil(t, block.Verify())

	block.Header.Height = 100

	assert.NotNil(t, block.Verify())
}
