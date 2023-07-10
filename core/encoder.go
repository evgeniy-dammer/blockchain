package core

import (
	"encoding/gob"
	"io"
)

// Encoder
type Encoder[T any] interface {
	Encode(T) error
}

// GobTransactionEncoder
type GobTransactionEncoder struct {
	w io.Writer
}

// NewGobTransactionEncoder is a constructor for the GobTransactionEncoder
func NewGobTransactionEncoder(w io.Writer) *GobTransactionEncoder {
	//gob.Register(elliptic.P256())

	return &GobTransactionEncoder{w: w}
}

// Encode encodes the transaction
func (e *GobTransactionEncoder) Encode(transaction *Transaction) error {
	return gob.NewEncoder(e.w).Encode(transaction)
}
