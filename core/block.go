package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/evgeniy-dammer/blockchain/types"
	"time"
)

// Header is a header of the Block
type Header struct {
	Version           uint32
	DataHash          types.Hash
	PreviousBlockHash types.Hash
	Timestamp         int64
	Height            uint32
}

// Bytes returns a block's Header as a slice of bytes
func (h *Header) Bytes() []byte {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(h); err != nil {
		return nil
	}

	return buf.Bytes()
}

// Block is a block of transactions
type Block struct {
	Header       *Header
	Transactions []*Transaction
	Validator    crypto.PublicKey
	Signature    *crypto.Signature
	hash         types.Hash // Cached version of the Header hash
}

// NewBlock is a constructor for the Block
func NewBlock(header *Header, transactions []*Transaction) *Block {
	return &Block{
		Header:       header,
		Transactions: transactions,
	}
}

// NewBlockFromPreviousHeader is a constructor for the Block with previous hash
func NewBlockFromPreviousHeader(previousHeader *Header, transactions []*Transaction) (*Block, error) {
	dataHash, err := CalculateDataHash(transactions)
	if err != nil {
		return nil, err
	}

	header := &Header{
		Version:           1,
		Height:            previousHeader.Height + 1,
		DataHash:          dataHash,
		PreviousBlockHash: BlockHasher{}.Hash(previousHeader),
		Timestamp:         time.Now().UnixNano(),
	}

	return NewBlock(header, transactions), nil
}

// AddTransaction adds Transaction to Blockchain
func (b *Block) AddTransaction(tx *Transaction) {
	b.Transactions = append(b.Transactions, tx)
}

// Sign signs a Block data
func (b *Block) Sign(privateKey crypto.PrivateKey) error {
	signature, err := privateKey.Sign(b.Header.Bytes())
	if err != nil {
		return err
	}

	b.Validator = privateKey.PublicKey()
	b.Signature = signature

	return nil
}

// Verify verifies a Block signature and transactions
func (b *Block) Verify() error {
	if b.Signature == nil {
		return fmt.Errorf("block has no signature")
	}

	if !b.Signature.Verify(b.Validator, b.Header.Bytes()) {
		return fmt.Errorf("invalid block signature")
	}

	// verify all transactions in block
	for _, tx := range b.Transactions {
		if err := tx.Verify(); err != nil {
			return err
		}
	}

	dataHash, err := CalculateDataHash(b.Transactions)
	if err != nil {
		return err
	}

	if dataHash != b.Header.DataHash {
		return fmt.Errorf("block %s has invalid data hash", b.Hash(BlockHasher{}))
	}

	return nil
}

// Decode decodes a Block
func (b *Block) Decode(decoder Decoder[*Block]) error {
	return decoder.Decode(b)
}

// Encode encodes a Block
func (b *Block) Encode(encoder Encoder[*Block]) error {
	return encoder.Encode(b)
}

// Hash hashes the Block
func (b *Block) Hash(hasher Hasher[*Header]) types.Hash {
	if b.hash.IsZero() {
		b.hash = hasher.Hash(b.Header)
	}

	return b.hash
}

// CalculateDataHash calculates hash of given transactions
func CalculateDataHash(transactions []*Transaction) (hash types.Hash, err error) {
	buf := &bytes.Buffer{}

	for _, transaction := range transactions {
		if err = transaction.Encode(NewGobTransactionEncoder(buf)); err != nil {
			return
		}
	}

	hash = sha256.Sum256(buf.Bytes())

	return
}
