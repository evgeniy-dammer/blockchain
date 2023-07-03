package core

import "fmt"

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
		return fmt.Errorf("chain already contains block %d with hash %s", block.Header.Height, block.Hash(BlockHasher{}))
	}

	if err := block.Verify(); err != nil {
		return err
	}

	return nil
}
