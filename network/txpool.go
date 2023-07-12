package network

import (
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/evgeniy-dammer/blockchain/types"
	"sync"
)

// TransactionPool
type TransactionPool struct {
	all       *TransactionSortedMap
	pending   *TransactionSortedMap
	maxLength int // The max length of the total pool of transactions. When the pool is full we will prune the oldest transaction
}

// NewTransactionPool is a constructor for a TransactionPool
func NewTransactionPool(maxLength int) *TransactionPool {
	return &TransactionPool{
		all:       NewTransactionSortedMap(),
		pending:   NewTransactionSortedMap(),
		maxLength: maxLength,
	}
}

// Add adds the transaction to the pool
func (p *TransactionPool) Add(transaction *core.Transaction) {
	// prune the oldest transaction that is sitting in the all pool
	if p.all.Count() == p.maxLength {
		oldest := p.all.First()
		p.all.Remove(oldest.Hash(core.TransactionHasher{}))
	}

	if !p.all.Contains(transaction.Hash(core.TransactionHasher{})) {
		p.all.Add(transaction)
		p.pending.Add(transaction)
	}
}

// Contains check if all pool contains hash
func (p *TransactionPool) Contains(hash types.Hash) bool {
	return p.all.Contains(hash)
}

// Pending return transactions from pending pool
func (p *TransactionPool) Pending() []*core.Transaction {
	return p.pending.transactions.Data
}

// ClearPending flushes pending pull
func (p *TransactionPool) ClearPending() {
	p.pending.Clear()
}

// PendingCount returns count of pending transactions
func (p *TransactionPool) PendingCount() int {
	return p.pending.Count()
}

// TransactionSortedMap
type TransactionSortedMap struct {
	lock         sync.RWMutex
	lookup       map[types.Hash]*core.Transaction
	transactions *types.List[*core.Transaction]
}

// NewTransactionSortedMap is a constructor for the TransactionSortedMap
func NewTransactionSortedMap() *TransactionSortedMap {
	return &TransactionSortedMap{
		lookup:       make(map[types.Hash]*core.Transaction),
		transactions: types.NewList[*core.Transaction](),
	}
}

// First returns the first transaction
func (t *TransactionSortedMap) First() *core.Transaction {
	t.lock.RLock()
	defer t.lock.RUnlock()

	first := t.transactions.Get(0)

	return t.lookup[first.Hash(core.TransactionHasher{})]
}

// Get returns the transaction with given hash
func (t *TransactionSortedMap) Get(hash types.Hash) *core.Transaction {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.lookup[hash]
}

// Add adds the transaction into the sorted map
func (t *TransactionSortedMap) Add(transaction *core.Transaction) {
	hash := transaction.Hash(core.TransactionHasher{})

	t.lock.Lock()
	defer t.lock.Unlock()

	if _, ok := t.lookup[hash]; !ok {
		t.lookup[hash] = transaction
		t.transactions.Insert(transaction)
	}
}

// Remove removes the transaction from sorted map
func (t *TransactionSortedMap) Remove(hash types.Hash) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.transactions.Remove(t.lookup[hash])
	delete(t.lookup, hash)
}

// Count returns a count of sorted map elements
func (t *TransactionSortedMap) Count() int {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return len(t.lookup)
}

// Contains checks if sorted map contains the hash
func (t *TransactionSortedMap) Contains(hash types.Hash) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	_, ok := t.lookup[hash]

	return ok
}

// Clear clears the sorted map
func (t *TransactionSortedMap) Clear() {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.lookup = make(map[types.Hash]*core.Transaction)
	t.transactions.Clear()
}
