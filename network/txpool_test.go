package network

import (
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/evgeniy-dammer/blockchain/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTxMaxLength(t *testing.T) {
	p := NewTransactionPool(1)
	p.Add(util.NewRandomTransaction(10))
	assert.Equal(t, 1, p.all.Count())

	p.Add(util.NewRandomTransaction(10))
	p.Add(util.NewRandomTransaction(10))
	p.Add(util.NewRandomTransaction(10))
	tx := util.NewRandomTransaction(100)
	p.Add(tx)
	assert.Equal(t, 1, p.all.Count())
	assert.True(t, p.Contains(tx.Hash(core.TransactionHasher{})))
}

func TestTxPoolAdd(t *testing.T) {
	p := NewTransactionPool(11)
	n := 10

	for i := 1; i <= n; i++ {
		tx := util.NewRandomTransaction(100)
		p.Add(tx)
		// cannot add twice
		p.Add(tx)

		assert.Equal(t, i, p.PendingCount())
		assert.Equal(t, i, p.pending.Count())
		assert.Equal(t, i, p.all.Count())
	}
}

func TestTxPoolMaxLength(t *testing.T) {
	maxLen := 10
	p := NewTransactionPool(maxLen)
	n := 100
	txx := []*core.Transaction{}

	for i := 0; i < n; i++ {
		tx := util.NewRandomTransaction(100)
		p.Add(tx)

		if i > n-(maxLen+1) {
			txx = append(txx, tx)
		}
	}

	assert.Equal(t, p.all.Count(), maxLen)
	assert.Equal(t, len(txx), maxLen)

	for _, tx := range txx {
		assert.True(t, p.Contains(tx.Hash(core.TransactionHasher{})))
	}
}

func TestTxSortedMapFirst(t *testing.T) {
	m := NewTransactionSortedMap()
	first := util.NewRandomTransaction(100)
	m.Add(first)
	m.Add(util.NewRandomTransaction(10))
	m.Add(util.NewRandomTransaction(10))
	m.Add(util.NewRandomTransaction(10))
	m.Add(util.NewRandomTransaction(10))
	assert.Equal(t, first, m.First())
}

func TestTxSortedMapAdd(t *testing.T) {
	m := NewTransactionSortedMap()
	n := 100

	for i := 0; i < n; i++ {
		tx := util.NewRandomTransaction(100)
		m.Add(tx)
		// cannot add the same twice
		m.Add(tx)

		assert.Equal(t, m.Count(), i+1)
		assert.True(t, m.Contains(tx.Hash(core.TransactionHasher{})))
		assert.Equal(t, len(m.lookup), m.transactions.Len())
		assert.Equal(t, m.Get(tx.Hash(core.TransactionHasher{})), tx)
	}

	m.Clear()
	assert.Equal(t, m.Count(), 0)
	assert.Equal(t, len(m.lookup), 0)
	assert.Equal(t, m.transactions.Len(), 0)
}

func TestTxSortedMapRemove(t *testing.T) {
	m := NewTransactionSortedMap()

	tx := util.NewRandomTransaction(100)
	m.Add(tx)
	assert.Equal(t, m.Count(), 1)

	m.Remove(tx.Hash(core.TransactionHasher{}))
	assert.Equal(t, m.Count(), 0)
	assert.False(t, m.Contains(tx.Hash(core.TransactionHasher{})))
}
