package core

import (
	"crypto/sha256"
	"github.com/evgeniy-dammer/blockchain/types"
)

// Hasher
type Hasher[T any] interface {
	Hash(T) types.Hash
}

// BlockHasher
type BlockHasher struct {
}

// Hash a block's Header
func (BlockHasher) Hash(header *Header) types.Hash {
	hash := sha256.Sum256(header.Bytes())

	return types.Hash(hash)
}
