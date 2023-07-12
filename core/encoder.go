package core

import (
	"crypto/elliptic"
	"encoding/gob"
	"io"
)

func init() {
	gob.Register(elliptic.P256())
}

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

// GobBlockEncoder
type GobBlockEncoder struct {
	w io.Writer
}

// NewGobBlockEncoder is a constructor for the GobBlockEncoder
func NewGobBlockEncoder(w io.Writer) *GobBlockEncoder {
	return &GobBlockEncoder{
		w: w,
	}
}

// Encode encodes the block
func (enc *GobBlockEncoder) Encode(b *Block) error {
	return gob.NewEncoder(enc.w).Encode(b)
}
