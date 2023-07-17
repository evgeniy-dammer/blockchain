package core

import (
	"fmt"
	"github.com/go-kit/log"
	"sync"
)

// Blockchain
type Blockchain struct {
	logger        log.Logger
	store         Storage
	lock          sync.RWMutex
	headers       []*Header
	validator     Validator
	contractState *State
}

// NewBlockchain is a constructor for the Blockchain
func NewBlockchain(logger log.Logger, genesis *Block) (*Blockchain, error) {
	blockchain := &Blockchain{
		headers:       []*Header{},
		store:         NewMemoryStore(),
		logger:        logger,
		contractState: NewState(),
	}

	blockchain.validator = NewBlockValidator(blockchain)

	err := blockchain.addBlockWithoutValidation(genesis)

	return blockchain, err
}

// SetValidator sets the validator fot the blockchain
func (bc *Blockchain) SetValidator(validator Validator) {
	bc.validator = validator
}

// AddBlock  validates a block and adds it into blockchain
func (bc *Blockchain) AddBlock(block *Block) error {
	if err := bc.validator.ValidateBlock(block); err != nil {
		return err
	}

	for _, transaction := range block.Transactions {
		bc.logger.Log("msg", "executing code", "len", len(transaction.Data), "hash", transaction.Hash(&TransactionHasher{}))

		virtualMachine := NewVirtualMachine(transaction.Data, bc.contractState)

		if err := virtualMachine.Run(); err != nil {
			return err
		}
	}

	if err := bc.addBlockWithoutValidation(block); err != nil {
		return err
	}

	return nil
}

// GetHeader returns blockchain's Header with given height
func (bc *Blockchain) GetHeader(height uint32) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("given height %d is too high", height)
	}

	bc.lock.RLock()
	defer bc.lock.RUnlock()

	return bc.headers[height], nil
}

// HasBlock checks if blockchain has block with given height
func (bc *Blockchain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

// Height returns a blockchain's height
func (bc *Blockchain) Height() uint32 {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	return uint32(len(bc.headers) - 1)
}

// addBlockWithoutValidation adds a block into blockchain without validation
func (bc *Blockchain) addBlockWithoutValidation(block *Block) error {
	bc.lock.RLock()

	bc.headers = append(bc.headers, block.Header)

	bc.lock.RUnlock()

	bc.logger.Log(
		"msg", "adding new block",
		"height", block.Header.Height,
		"hash", block.Hash(BlockHasher{}),
		"transactions", len(block.Transactions),
	)

	return bc.store.Put(block)
}
