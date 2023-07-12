package core

import (
	"crypto/elliptic"
	"encoding/gob"
	"io"
)

func init() {
	gob.Register(elliptic.P256())
}

// Decoder
type Decoder[T any] interface {
	Decode(T) error
}

// GobTransactionDecoder
type GobTransactionDecoder struct {
	r io.Reader
}

// NewGobTransactionDecoder is a constructor for the GobTransactionDecoder
func NewGobTransactionDecoder(r io.Reader) *GobTransactionDecoder {
	//gob.Register(elliptic.P256())

	return &GobTransactionDecoder{r: r}
}

// Decode decodes the transaction
func (e *GobTransactionDecoder) Decode(transaction *Transaction) error {
	return gob.NewDecoder(e.r).Decode(transaction)
}

// GobBlockDecoder
type GobBlockDecoder struct {
	r io.Reader
}

// NewGobBlockDecoder is a constructor for the GobBlockDecoder
func NewGobBlockDecoder(r io.Reader) *GobBlockDecoder {
	return &GobBlockDecoder{
		r: r,
	}
}

// Decode decodes the block
func (dec *GobBlockDecoder) Decode(b *Block) error {
	return gob.NewDecoder(dec.r).Decode(b)
}
