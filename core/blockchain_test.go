package core

import (
	"github.com/evgeniy-dammer/blockchain/types"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newBlockchainWithGenesis(t *testing.T) *Blockchain {
	accountState := NewAccountState()

	blockchain, err := NewBlockchain(log.NewNopLogger(), randomBlock(t, 0, types.Hash{}), accountState)

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

func TestBlockchain_GetBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	lenBlocks := 100

	for i := 0; i < lenBlocks; i++ {
		block := randomBlock(t, uint32(i+1), getPreviousBlockHash(t, bc, uint32(i+1)))
		assert.Nil(t, bc.AddBlock(block))

		fetchedBlock, err := bc.GetBlock(block.Header.Height)
		assert.Nil(t, err)
		assert.Equal(t, fetchedBlock, block)
	}
}

func TestBlockchain_AddBlock(t *testing.T) {
	blockchain := newBlockchainWithGenesis(t)
	block := randomBlock(t, uint32(1), getPreviousBlockHash(t, blockchain, uint32(1)))

	assert.Nil(t, blockchain.AddBlock(block))
}

func TestBlockchain_GetHeader(t *testing.T) {
	blockchain := newBlockchainWithGenesis(t)

	assert.NotNil(t, blockchain.AddBlock(randomBlock(t, 3, getPreviousBlockHash(t, blockchain, uint32(1)))))
}

func TestBlockchain_AddBlockTooHigh(t *testing.T) {
	blockchain := newBlockchainWithGenesis(t)

	block := randomBlock(t, uint32(1), getPreviousBlockHash(t, blockchain, uint32(1)))
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
