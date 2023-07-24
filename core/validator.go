package core

import (
	"errors"
	"fmt"
)

var ErrBlockKnown = errors.New("block already known")

// Validator interface
type Validator interface {
	ValidateBlock(block *Block) error
}

// BlockValidator
type BlockValidator struct {
	blockchain *Blockchain
}

// NewBlockValidator is a constructor for the BlockValidator
func NewBlockValidator(blockchain *Blockchain) *BlockValidator {
	return &BlockValidator{blockchain: blockchain}
}

// ValidateBlock validates and verifies a block
func (bv *BlockValidator) ValidateBlock(block *Block) error {
	if bv.blockchain.HasBlock(block.Header.Height) {
		return ErrBlockKnown
	}

	if block.Header.Height != bv.blockchain.Height()+1 {
		return fmt.Errorf("block %s too high", block.Hash(BlockHasher{}))
	}

	prevHeader, err := bv.blockchain.GetHeader(block.Header.Height - 1)
	if err != nil {
		return err
	}

	hash := BlockHasher{}.Hash(prevHeader)

	if hash != block.Header.PreviousBlockHash {
		return fmt.Errorf("the hash of the previous block %s is invalid", block.Header.PreviousBlockHash)
	}

	if err := block.Verify(); err != nil {
		return err
	}

	return nil
}
