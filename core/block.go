package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"github.com/evgeniy-dammer/blockchain/types"
	"io"
)

// byteOrder
var byteOrder = binary.LittleEndian

// Header is a header of the Block
type Header struct {
	Version       uint32
	PreviousBlock types.Hash
	Timestamp     int64
	Height        uint32
	Nonce         uint64
}

// EncodeBinary encodes the Header
func (h *Header) EncodeBinary(w io.Writer) error {
	if err := binary.Write(w, byteOrder, &h.Version); err != nil {
		return err
	}

	if err := binary.Write(w, byteOrder, &h.PreviousBlock); err != nil {
		return err
	}

	if err := binary.Write(w, byteOrder, &h.Timestamp); err != nil {
		return err
	}

	if err := binary.Write(w, byteOrder, &h.Height); err != nil {
		return err
	}

	if err := binary.Write(w, byteOrder, &h.Nonce); err != nil {
		return err
	}

	return nil
}

// DecodeBinary decodes the Header
func (h *Header) DecodeBinary(r io.Reader) error {
	if err := binary.Read(r, byteOrder, &h.Version); err != nil {
		return err
	}

	if err := binary.Read(r, byteOrder, &h.PreviousBlock); err != nil {
		return err
	}

	if err := binary.Read(r, byteOrder, &h.Timestamp); err != nil {
		return err
	}

	if err := binary.Read(r, byteOrder, &h.Height); err != nil {
		return err
	}

	if err := binary.Read(r, byteOrder, &h.Nonce); err != nil {
		return err
	}

	return nil
}

// Block is a block of transactions
type Block struct {
	Header       Header
	Transactions []Transaction
	hash         types.Hash // Cached version of the Header hash
}

// Hash hashes the Block
func (b *Block) Hash() types.Hash {
	buf := &bytes.Buffer{}

	if err := b.Header.EncodeBinary(buf); err != nil {
		return [32]uint8{}
	}

	if b.hash.IsZero() {
		b.hash = types.Hash(sha256.Sum256(buf.Bytes()))
	}

	return b.hash
}

// EncodeBinary encodes the Block
func (b *Block) EncodeBinary(w io.Writer) error {
	if err := b.Header.EncodeBinary(w); err != nil {
		return err
	}

	for _, tx := range b.Transactions {
		if err := tx.EncodeBinary(w); err != nil {
			return nil
		}
	}

	return nil
}

// DecodeBinary decodes the Block
func (b *Block) DecodeBinary(r io.Reader) error {
	if err := b.Header.DecodeBinary(r); err != nil {
		return err
	}

	for _, tx := range b.Transactions {
		if err := tx.DecodeBinary(r); err != nil {
			return nil
		}
	}

	return nil
}
