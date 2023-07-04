package core

import (
	"fmt"
	"github.com/evgeniy-dammer/blockchain/crypto"
)

// Transaction
type Transaction struct {
	Data      []byte
	From      crypto.PublicKey
	Signature *crypto.Signature
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
