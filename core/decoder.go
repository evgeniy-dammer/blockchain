package core

import (
	"encoding/gob"
	"io"
)

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
