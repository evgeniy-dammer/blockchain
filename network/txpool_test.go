package network

import (
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
)

func TestNewTxPool(t *testing.T) {
	pool := NewTransactionPool()

	assert.Equal(t, pool.Len(), 0)
}

func TestTxPool_Add(t *testing.T) {
	pool := NewTransactionPool()

	tx := core.NewTransaction([]byte("foo"))

	assert.Nil(t, pool.Add(tx))
	assert.Equal(t, pool.Len(), 1)

	_ = core.NewTransaction([]byte("foo"))
	assert.Equal(t, pool.Len(), 1)

	pool.Flush()
	assert.Equal(t, pool.Len(), 0)
}

func TestNewTransactionMapSorter(t *testing.T) {
	pool := NewTransactionPool()

	count := 1000

	for i := 0; i < count; i++ {
		tx := core.NewTransaction([]byte(strconv.FormatInt(int64(i), 10)))
		tx.SetFirstSeen(int64(i * rand.Intn(10000)))
		assert.Nil(t, pool.Add(tx))
	}

	assert.Equal(t, count, pool.Len())

	txx := pool.Transactions()

	for i := 0; i < len(txx)-1; i++ {
		assert.True(t, txx[i].FirstSeen() < txx[i+1].FirstSeen())
	}

	//sorter := NewTransactionMapSorter(pool.transactions)
}
