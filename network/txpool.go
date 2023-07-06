package network

import (
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/evgeniy-dammer/blockchain/types"
	"sort"
)

// TransactionMapSorter
type TransactionMapSorter struct {
	transactions []*core.Transaction
}

// NewTransactionMapSorter is a constructor for the TransactionMapSorter
func NewTransactionMapSorter(transactionMap map[types.Hash]*core.Transaction) *TransactionMapSorter {
	txMap := make([]*core.Transaction, len(transactionMap))

	i := 0
	for _, val := range transactionMap {
		txMap[i] = val
		i++
	}

	sorter := &TransactionMapSorter{txMap}

	sort.Sort(sorter)

	return sorter
}

// Len returns length of sorter
func (s *TransactionMapSorter) Len() int {
	return len(s.transactions)
}

// Swap swaps transactions with given indexes
func (s *TransactionMapSorter) Swap(i, j int) {
	s.transactions[i], s.transactions[j] = s.transactions[j], s.transactions[i]
}

// Less checks if one transaction is less than other
func (s *TransactionMapSorter) Less(i, j int) bool {
	return s.transactions[i].FirstSeen() < s.transactions[j].FirstSeen()
}

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

// Transactions returns transactions of pool as a slice
func (p *TransactionPool) Transactions() []*core.Transaction {
	sorter := NewTransactionMapSorter(p.transactions)

	return sorter.transactions
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
