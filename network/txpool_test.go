package network

import (
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/stretchr/testify/assert"
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
