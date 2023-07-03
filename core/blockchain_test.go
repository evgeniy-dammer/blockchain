package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func newBlockchainWithGenesis(t *testing.T) *Blockchain {
	blockchain, err := NewBlockchain(randomBlock(0))

	assert.Nil(t, err)

	return blockchain
}

func TestNewBlockchain(t *testing.T) {
	blockchain := newBlockchainWithGenesis(t)

	assert.NotNil(t, blockchain.validator)
	assert.Equal(t, blockchain.Height(), uint32(0))
}

func TestBlockchain_HasBlock(t *testing.T) {
	blockchain := newBlockchainWithGenesis(t)

	assert.True(t, blockchain.HasBlock(0))
}

func TestBlockchain_AddBlock(t *testing.T) {
	blockchain := newBlockchainWithGenesis(t)
	block := randomBlockWithSignature(t, uint32(1))

	assert.Nil(t, blockchain.AddBlock(block))
}
