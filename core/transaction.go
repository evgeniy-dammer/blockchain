package core

import "io"

// Transaction
type Transaction struct {
	Data []byte
}

// EncodeBinary
func (t *Transaction) EncodeBinary(w io.Writer) error {
	return nil
}

// DecodeBinary
func (t *Transaction) DecodeBinary(r io.Reader) error {
	return nil
}
