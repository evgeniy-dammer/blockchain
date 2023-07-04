package core

import (
	"fmt"
	"log"
)

// Blockchain
type Blockchain struct {
	store     Storage
	headers   []*Header
	validator Validator
}

// NewBlockchain is a constructor for the Blockchain
func NewBlockchain(genesis *Block) (*Blockchain, error) {
	blockchain := &Blockchain{headers: []*Header{}, store: NewMemoryStore()}
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

	return bc.headers[height], nil
}

// HasBlock checks if blockchain has block with given height
func (bc *Blockchain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

// Height returns a blockchain's height
func (bc *Blockchain) Height() uint32 {
	return uint32(len(bc.headers) - 1)
}

// addBlockWithoutValidation adds a block into blockchain without validation
func (bc *Blockchain) addBlockWithoutValidation(block *Block) error {
	bc.headers = append(bc.headers, block.Header)

	log.Printf("adding new block: height - %d, hash - %s", block.Header, block.Hash(BlockHasher{}))

	return bc.store.Put(block)
}
