package core

import (
	"github.com/evgeniy-dammer/blockchain/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newBlockchainWithGenesis(t *testing.T) *Blockchain {
	blockchain, err := NewBlockchain(randomBlock(0, types.Hash{}))

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
	block := randomBlockWithSignature(t, uint32(1), getPreviousBlockHash(t, blockchain, uint32(1)))

	assert.Nil(t, blockchain.AddBlock(block))
}

func TestBlockchain_GetHeader(t *testing.T) {
	blockchain := newBlockchainWithGenesis(t)

	assert.NotNil(t, blockchain.AddBlock(randomBlockWithSignature(t, 3, getPreviousBlockHash(t, blockchain, uint32(1)))))
}

func TestBlockchain_AddBlockTooHigh(t *testing.T) {
	blockchain := newBlockchainWithGenesis(t)

	block := randomBlockWithSignature(t, uint32(1), getPreviousBlockHash(t, blockchain, uint32(1)))
	assert.Nil(t, blockchain.AddBlock(block))
 
	header, err := blockchain.GetHeader(block.Header.Height)
	assert.Nil(t, err)

	assert.Equal(t, header, block.Header)
}

func getPreviousBlockHash(t *testing.T, blockchain *Blockchain, height uint32) types.Hash {
	prevHeader, err := blockchain.GetHeader(height - 1)

	assert.Nil(t, err)

	return BlockHasher{}.Hash(prevHeader)
}
