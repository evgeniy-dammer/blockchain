package core

import (
	"fmt"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/evgeniy-dammer/blockchain/types"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newBlockchainWithGenesis(t *testing.T) *Blockchain {
	blockchain, err := NewBlockchain(log.NewNopLogger(), randomBlock(t, 0, types.Hash{}))

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

func TestSendNativeTransferTamper(t *testing.T) {
	bc := newBlockchainWithGenesis(t)

	signer := crypto.GeneratePrivateKey()

	block := randomBlock(t, uint32(1), getPreviousBlockHash(t, bc, uint32(1)))
	assert.Nil(t, block.Sign(signer))

	privKeyBob := crypto.GeneratePrivateKey()
	privKeyAlice := crypto.GeneratePrivateKey()
	amount := uint64(100)

	accountBob := bc.accountState.CreateAccount(privKeyBob.PublicKey().Address())
	accountBob.Balance = amount

	tx := NewTransaction([]byte{})
	tx.From = privKeyBob.PublicKey()
	tx.To = privKeyAlice.PublicKey()
	tx.Value = amount
	tx.Sign(privKeyBob)

	hackerPrivKey := crypto.GeneratePrivateKey()
	tx.To = hackerPrivKey.PublicKey()

	block.AddTransaction(tx)

	assert.Nil(t, bc.AddBlock(block)) // this should fail

	//fmt.Printf("%+v\n", hackerPrivKey.PublicKey().Address())
	fmt.Printf("%+v\n", bc.accountState.accounts)
	fmt.Printf("%+v\n", privKeyAlice.PublicKey().Address())

	accountHacker, err := bc.accountState.GetAccount(hackerPrivKey.PublicKey().Address())
	fmt.Printf("%s\n", err)
	assert.NotNil(t, err)

	assert.Equal(t, uint64(0), accountHacker.Balance)

	_, err = bc.accountState.GetAccount(privKeyAlice.PublicKey().Address())
	assert.NotNil(t, err)
	// assert.Equal(t, accountAlice.Balance, amount)
}

func TestSendNativeTransferInsuffientBalance(t *testing.T) {
	bc := newBlockchainWithGenesis(t)

	signer := crypto.GeneratePrivateKey()

	block := randomBlock(t, uint32(1), getPreviousBlockHash(t, bc, uint32(1)))
	assert.Nil(t, block.Sign(signer))

	privKeyBob := crypto.GeneratePrivateKey()
	privKeyAlice := crypto.GeneratePrivateKey()
	amount := uint64(100)

	accountBob := bc.accountState.CreateAccount(privKeyBob.PublicKey().Address())
	accountBob.Balance = uint64(99)

	tx := NewTransaction([]byte{})
	tx.From = privKeyBob.PublicKey()
	tx.To = privKeyAlice.PublicKey()
	tx.Value = amount
	tx.Sign(privKeyBob)
	block.AddTransaction(tx)
	assert.NotNil(t, bc.AddBlock(block))

	_, err := bc.accountState.GetAccount(privKeyAlice.PublicKey().Address())
	assert.NotNil(t, err)
}

func TestSendNativeTransferSuccess(t *testing.T) {
	bc := newBlockchainWithGenesis(t)

	signer := crypto.GeneratePrivateKey()

	block := randomBlock(t, uint32(1), getPreviousBlockHash(t, bc, uint32(1)))
	assert.Nil(t, block.Sign(signer))

	privKeyBob := crypto.GeneratePrivateKey()
	privKeyAlice := crypto.GeneratePrivateKey()
	amount := uint64(100)

	accountBob := bc.accountState.CreateAccount(privKeyBob.PublicKey().Address())
	accountBob.Balance = amount

	tx := NewTransaction([]byte{})
	tx.From = privKeyBob.PublicKey()
	tx.To = privKeyAlice.PublicKey()
	tx.Value = amount
	tx.Sign(privKeyBob)
	block.AddTransaction(tx)
	assert.Nil(t, bc.AddBlock(block))

	accountAlice, err := bc.accountState.GetAccount(privKeyAlice.PublicKey().Address())
	assert.Nil(t, err)
	assert.Equal(t, amount, accountAlice.Balance)
}
