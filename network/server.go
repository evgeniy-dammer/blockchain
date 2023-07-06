package network

import (
	"fmt"
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"log"
	"time"
)

var defaultBlockTime = time.Second * 5

// ServerOptions
type ServerOptions struct {
	Transports []Transport
	BlockTime  time.Duration
	PrivateKey *crypto.PrivateKey
}

// Server
type Server struct {
	options     ServerOptions
	blockTime   time.Duration
	memoryPool  *TransactionPool
	isValidator bool
	rpcCh       chan RPC
	quitCh      chan struct{}
}

// NewServer is a constructor for the Server
func NewServer(options ServerOptions) *Server {
	if options.BlockTime == time.Duration(0) {
		options.BlockTime = defaultBlockTime
	}

	return &Server{
		options:     options,
		blockTime:   options.BlockTime,
		memoryPool:  NewTransactionPool(),
		isValidator: options.PrivateKey != nil,
		rpcCh:       make(chan RPC),
		quitCh:      make(chan struct{}, 1),
	}
}

// Start starts the Server
func (s *Server) Start() {
	s.initTransports()
	ticker := time.NewTicker(s.blockTime)

LOOP:
	for {
		select {
		case rpc := <-s.rpcCh:
			fmt.Printf("%+v\n", rpc)
		case <-s.quitCh:
			break LOOP
		case <-ticker.C:
			if s.isValidator {
				s.createNewBlock()
			}
		}
	}

	fmt.Println("Server shutdown...")
}

// handleTransaction handles new transaction from network and adds it into memory pool
func (s *Server) handleTransaction(transaction *core.Transaction) error {
	if err := transaction.Verify(); err != nil {
		return err
	}

	hash := transaction.Hash(core.TransactionHasher{})

	if s.memoryPool.Has(hash) {
		log.Printf("transaction with hash %s is already in mempool", hash)

		return nil
	}

	log.Printf("adding new transaction with hash %s into mempool", hash)

	return s.memoryPool.Add(transaction)
}

// createNewBlock creates a new block
func (s *Server) createNewBlock() error {
	fmt.Println("creating a new block...")

	return nil
}

// initTransports initializes Transports
func (s *Server) initTransports() {
	for _, tr := range s.options.Transports {
		go func(tr Transport) {
			for rpc := range tr.Consume() {
				s.rpcCh <- rpc
			}
		}(tr)
	}
}
