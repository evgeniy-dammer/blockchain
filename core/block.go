package core

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/evgeniy-dammer/blockchain/types"
	"io"
)

// Header is a header of the Block
type Header struct {
	Version           uint32
	DataHash          types.Hash
	PreviousBlockHash types.Hash
	Timestamp         int64
	Height            uint32
}

// Block is a block of transactions
type Block struct {
	Header       *Header
	Transactions []Transaction
	Validator    crypto.PublicKey
	Signature    *crypto.Signature
	hash         types.Hash // Cached version of the Header hash
}

// NewBlock is a constructor for the Block
func NewBlock(header *Header, transactions []Transaction) *Block {
	return &Block{
		Header:       header,
		Transactions: transactions,
	}
}

// Sign signs a Block data
func (b *Block) Sign(privateKey crypto.PrivateKey) error {
	signature, err := privateKey.Sign(b.HeaderData())
	if err != nil {
		return err
	}

	b.Validator = privateKey.PublicKey()
	b.Signature = signature

	return nil
}

// Verify verifies a Block signature
func (b *Block) Verify() error {
	if b.Signature == nil {
		return fmt.Errorf("block has no signature")
	}

	if !b.Signature.Verify(b.Validator, b.HeaderData()) {
		return fmt.Errorf("invalid block signature")
	}

	return nil
}

// Decode decodes a Block
func (b *Block) Decode(r io.Reader, decoder Decoder[*Block]) error {
	return decoder.Decode(r, b)
}

// Encode encodes a Block
func (b *Block) Encode(w io.Writer, encoder Encoder[*Block]) error {
	return encoder.Encode(w, b)
}

// Hash hashes the Block
func (b *Block) Hash(hasher Hasher[*Block]) types.Hash {
	if b.hash.IsZero() {
		b.hash = hasher.Hash(b)
	}

	return b.hash
}

// HeaderData returns a block's Header as a slice of bytes
func (b *Block) HeaderData() []byte {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(b.Header); err != nil {
		return nil
	}

	return buf.Bytes()
}
