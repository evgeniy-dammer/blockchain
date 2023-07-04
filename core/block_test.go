package core

import (
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/evgeniy-dammer/blockchain/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// randomBlock returns a random Block
func randomBlock(height uint32, prevBlockHash types.Hash) *Block {
	header := &Header{
		Version:           1,
		PreviousBlockHash: prevBlockHash,
		Timestamp:         time.Now().UnixNano(),
		Height:            height,
	}

	return NewBlock(header, []Transaction{})
}

func randomBlockWithSignature(t *testing.T, height uint32, prevBlockHash types.Hash) *Block {
	privateKey := crypto.GeneratePrivateKey()
	block := randomBlock(height, prevBlockHash)

	tx := randomTransactionWithSignature(t)
	block.AddTransaction(tx)

	assert.Nil(t, block.Sign(privateKey))

	return block
}

func TestBlock_Sign(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	block := randomBlock(0, types.Hash{})

	assert.Nil(t, block.Sign(privateKey))
	assert.NotNil(t, block.Signature)
}

func TestBlock_Verify(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	block := randomBlock(0, types.Hash{})

	assert.Nil(t, block.Sign(privateKey))
	assert.Nil(t, block.Verify())

	otherPrivateKey := crypto.GeneratePrivateKey()
	block.Validator = otherPrivateKey.PublicKey()

	assert.NotNil(t, block.Verify())

	block.Header.Height = 100

	assert.NotNil(t, block.Verify())
}
