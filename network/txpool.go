package network

import (
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/evgeniy-dammer/blockchain/types"
)

// TransactionPool
type TransactionPool struct {
	transactions map[types.Hash]*core.Transaction
}

// NewTransactionPool is a constructor for a TransactionPool
func NewTransactionPool() *TransactionPool {
	return &TransactionPool{
		transactions: make(map[types.Hash]*core.Transaction),
	}
}

// Add adds the transaction to the pool, the caller is responsible checking if the transaction already exists
func (p *TransactionPool) Add(transaction *core.Transaction) error {
	hash := transaction.Hash(core.TransactionHasher{})

	p.transactions[hash] = transaction

	return nil
}

// Has checks if transaction is already in memory pool
func (p *TransactionPool) Has(hash types.Hash) bool {
	_, ok := p.transactions[hash]

	return ok
}

// Len returns a length of memory pool
func (p *TransactionPool) Len() int {
	return len(p.transactions)
}

// Flush flushes the memory pool
func (p *TransactionPool) Flush() {
	p.transactions = make(map[types.Hash]*core.Transaction)
}
