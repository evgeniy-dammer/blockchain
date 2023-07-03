package core

// Storage interface
type Storage interface {
	Put(block *Block) error
}

// MemoryStore
type MemoryStore struct{}

// NewMemoryStore is a constructor for the MemoryStore
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

// Put puts a block into memory store
func (ms *MemoryStore) Put(block *Block) error {
	return nil
}
