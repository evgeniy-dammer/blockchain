package core

import (
	"crypto/elliptic"
	"encoding/gob"
	"io"
)

// Encoder
type Encoder[T any] interface {
	Encode(T) error
}

// GobTransactionEncoder
type GobTransactionEncoder struct {
	W io.Writer
}

// NewGobTransactionEncoder is a constructor for the GobTransactionEncoder
func NewGobTransactionEncoder(w io.Writer) *GobTransactionEncoder {
	gob.Register(elliptic.P256())

	return &GobTransactionEncoder{W: w}
}

// Encode encodes the transaction
func (e *GobTransactionEncoder) Encode(transaction *Transaction) error {
	return gob.NewEncoder(e.W).Encode(transaction)
}
