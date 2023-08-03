package core

import (
	"fmt"
	"github.com/evgeniy-dammer/blockchain/types"
	"github.com/go-kit/log"
	"sync"
)

// Blockchain
type Blockchain struct {
	logger          log.Logger
	store           Storage
	lock            sync.RWMutex
	headers         []*Header
	blocks          []*Block
	txStore         map[types.Hash]*Transaction
	blockStore      map[types.Hash]*Block
	accountState    *AccountState
	stateLock       sync.RWMutex
	collectionState map[types.Hash]*CollectionTx
	mintState       map[types.Hash]*MintTx
	validator       Validator
	contractState   *State
}

// NewBlockchain is a constructor for the Blockchain
func NewBlockchain(logger log.Logger, genesis *Block, ac *AccountState) (*Blockchain, error) {
	blockchain := &Blockchain{
		headers:         []*Header{},
		store:           NewMemoryStore(),
		logger:          logger,
		accountState:    ac,
		blockStore:      make(map[types.Hash]*Block),
		txStore:         make(map[types.Hash]*Transaction),
		collectionState: make(map[types.Hash]*CollectionTx),
		mintState:       make(map[types.Hash]*MintTx),
		contractState:   NewState(),
	}

	blockchain.validator = NewBlockValidator(blockchain)

	err := blockchain.addBlockWithoutValidation(genesis)

	return blockchain, err
}

func (bc *Blockchain) GetBlockByHash(hash types.Hash) (*Block, error) {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	block, ok := bc.blockStore[hash]
	if !ok {
		return nil, fmt.Errorf("block with hash (%s) not found", hash)
	}

	return block, nil
}

func (bc *Blockchain) GetTransactionByHash(hash types.Hash) (*Transaction, error) {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	tx, ok := bc.txStore[hash]
	if !ok {
		return nil, fmt.Errorf("could not find tx with hash (%s)", hash)
	}

	return tx, nil
}

// SetValidator sets the validator fot the blockchain
func (bc *Blockchain) SetValidator(validator Validator) {
	bc.validator = validator
}

// AddBlock  validates a block and adds it into blockchain
func (bc *Blockchain) AddBlock(block *Block) error {
	if err := bc.validator.ValidateBlock(block); err != nil {
		return err
	}

	bc.stateLock.Lock()
	defer bc.stateLock.Unlock()

	for _, transaction := range block.Transactions {
		// If we have data inside execute that data on the VM.
		if len(transaction.Data) > 0 {
			bc.logger.Log("msg", "executing code", "len", len(transaction.Data), "hash", transaction.Hash(&TransactionHasher{}))

			vm := NewVirtualMachine(transaction.Data, bc.contractState)
			if err := vm.Run(); err != nil {
				return err
			}
		}

		// If the txInner of the transaction is not nil we need to handle
		// the native NFT implemtation.
		if transaction.TxInner != nil {
			if err := bc.handleNativeNFT(transaction); err != nil {
				return err
			}
		}

		// Handle the native transaction here
		if transaction.Value > 0 {
			if err := bc.handleNativeTransfer(transaction); err != nil {
				return err
			}
		}
	}

	if err := bc.addBlockWithoutValidation(block); err != nil {
		return err
	}

	return nil
}

func (bc *Blockchain) handleNativeTransfer(tx *Transaction) error {
	bc.logger.Log("msg", "handle native token transfer", "from", tx.From, "to", tx.To, "value", tx.Value)

	return bc.accountState.Transfer(tx.From.Address(), tx.To.Address(), tx.Value)
}

func (bc *Blockchain) handleNativeNFT(tx *Transaction) error {
	hash := tx.Hash(TransactionHasher{})

	switch t := tx.TxInner.(type) {
	case CollectionTx:
		bc.collectionState[hash] = &t
		bc.logger.Log("msg", "created new NFT collection", "hash", hash)
	case MintTx:
		_, ok := bc.collectionState[t.Collection]
		if !ok {
			return fmt.Errorf("collection (%s) does not exist on the blockchain", t.Collection)
		}
		bc.mintState[hash] = &t

		bc.logger.Log("msg", "created new NFT mint", "NFT", t.NFT, "collection", t.Collection)
	default:
		return fmt.Errorf("unsupported tx type %v", t)
	}

	return nil
}

func (bc *Blockchain) GetBlock(height uint32) (*Block, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("given height (%d) too high", height)
	}

	bc.lock.Lock()
	defer bc.lock.Unlock()

	return bc.blocks[height], nil
}

// GetHeader returns blockchain's Header with given height
func (bc *Blockchain) GetHeader(height uint32) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("given height %d is too high", height)
	}

	bc.lock.RLock()
	defer bc.lock.RUnlock()

	return bc.headers[height], nil
}

// HasBlock checks if blockchain has block with given height
func (bc *Blockchain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

// Height returns a blockchain's height
func (bc *Blockchain) Height() uint32 {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	return uint32(len(bc.headers) - 1)
}

// addBlockWithoutValidation adds a block into blockchain without validation
func (bc *Blockchain) addBlockWithoutValidation(block *Block) error {
	bc.lock.RLock()

	bc.headers = append(bc.headers, block.Header)
	bc.blocks = append(bc.blocks, block)
	bc.blockStore[block.Hash(BlockHasher{})] = block

	for _, tx := range block.Transactions {
		bc.txStore[tx.Hash(TransactionHasher{})] = tx
	}

	bc.lock.RUnlock()

	bc.logger.Log(
		"msg", "adding new block",
		"height", block.Header.Height,
		"hash", block.Hash(BlockHasher{}),
		"transactions", len(block.Transactions),
	)

	return bc.store.Put(block)
}
