package core

import (
	"encoding/gob"
	"fmt"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/evgeniy-dammer/blockchain/types"
	"math/rand"
)

type TxType byte

const (
	TxTypeCollection TxType = iota // 0x0
	TxTypeMint                     // 0x01
)

type CollectionTx struct {
	Fee      int64
	MetaData []byte
}

type MintTx struct {
	Fee             int64
	NFT             types.Hash
	Collection      types.Hash
	MetaData        []byte
	CollectionOwner crypto.PublicKey
	Signature       crypto.Signature
}

// Transaction
type Transaction struct {
	Type      TxType
	TxInner   any    // Only used for native NFT logic
	Data      []byte // Any arbitrary data for the VM
	To        crypto.PublicKey
	Value     uint64
	From      crypto.PublicKey
	Signature *crypto.Signature
	Nonce     int64
	hash      types.Hash // cached version of transaction data hash
}

// NewTransaction is a constructor for a Transaction
func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data:  data,
		Nonce: rand.Int63n(1000000000000000),
	}
}

// Hash returns a transactions hash
func (t *Transaction) Hash(hasher Hasher[*Transaction]) types.Hash {
	if t.hash.IsZero() {
		t.hash = hasher.Hash(t)
	}

	return t.hash
}

// Sign signs a Transaction data
func (t *Transaction) Sign(privateKey crypto.PrivateKey) error {
	signature, err := privateKey.Sign(t.Data)
	if err != nil {
		return err
	}

	t.From = privateKey.PublicKey()
	t.Signature = signature

	return nil
}

// Verify verifies a Transaction signature
func (t *Transaction) Verify() error {
	if t.Signature == nil {
		return fmt.Errorf("transaction has no signature")
	}

	if !t.Signature.Verify(t.From, t.Data) {
		return fmt.Errorf("invalid transaction signature")
	}

	return nil
}

// Encode encodes the transaction
func (t *Transaction) Encode(encoder Encoder[*Transaction]) error {
	return encoder.Encode(t)
}

// Decode decodes the transaction
func (t *Transaction) Decode(decoder Decoder[*Transaction]) error {
	return decoder.Decode(t)
}

func init() {
	gob.Register(CollectionTx{})
	gob.Register(MintTx{})
}
