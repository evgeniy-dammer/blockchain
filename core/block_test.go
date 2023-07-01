package core

import (
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/evgeniy-dammer/blockchain/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// randomBlock returns a random Block
func randomBlock(height uint32) *Block {
	header := &Header{
		Version:           1,
		PreviousBlockHash: types.RandomHash(),
		Timestamp:         time.Now().UnixNano(),
		Height:            height,
	}

	transaction := Transaction{
		Data: []byte("foo"),
	}

	return NewBlock(header, []Transaction{transaction})
}

func TestBlock_Sign(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	block := randomBlock(0)

	assert.Nil(t, block.Sign(privateKey))
	assert.NotNil(t, block.Signature)
}

func TestBlock_Verify(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	block := randomBlock(0)

	assert.Nil(t, block.Sign(privateKey))
	assert.Nil(t, block.Verify())

	otherPrivateKey := crypto.GeneratePrivateKey()
	block.Validator = otherPrivateKey.PublicKey()

	assert.NotNil(t, block.Verify())

	block.Header.Height = 100

	assert.NotNil(t, block.Verify())
}
