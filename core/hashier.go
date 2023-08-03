package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"github.com/evgeniy-dammer/blockchain/types"
)

// Hasher
type Hasher[T any] interface {
	Hash(T) types.Hash
}

// BlockHasher
type BlockHasher struct{}

// Hash hashes a block's Header
func (BlockHasher) Hash(header *Header) types.Hash {
	hash := sha256.Sum256(header.Bytes())

	return types.Hash(hash)
}

// TransactionHasher
type TransactionHasher struct{}

// Hash hashes a transaction's data
func (TransactionHasher) Hash(transaction *Transaction) types.Hash {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(transaction); err != nil {
		panic(err)
	}

	return types.Hash(sha256.Sum256(buf.Bytes()))
}
