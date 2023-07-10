package network

import (
	"bytes"
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/evgeniy-dammer/blockchain/types"
	"github.com/go-kit/log"
	"os"
	"time"
)

var defaultBlockTime = time.Second * 5

// ServerOptions
type ServerOptions struct {
	ID            string
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	Transports    []Transport
	BlockTime     time.Duration
	PrivateKey    *crypto.PrivateKey
}

// Server
type Server struct {
	options     ServerOptions
	memoryPool  *TransactionPool
	chain       *core.Blockchain
	isValidator bool
	rpcCh       chan RPC
	quitCh      chan struct{}
}

// NewServer is a constructor for the Server
func NewServer(options ServerOptions) (*Server, error) {
	if options.BlockTime == time.Duration(0) {
		options.BlockTime = defaultBlockTime
	}

	if options.RPCDecodeFunc == nil {
		options.RPCDecodeFunc = DefaultRPCDecodeFunc
	}

	if options.Logger == nil {
		options.Logger = log.NewLogfmtLogger(os.Stderr)
		options.Logger = log.With(options.Logger, "ID", options.ID)
	}

	chain, err := core.NewBlockchain(genesisBlock())
	if err != nil {
		return nil, err
	}

	server := &Server{
		options:     options,
		memoryPool:  NewTransactionPool(),
		chain:       chain,
		isValidator: options.PrivateKey != nil,
		rpcCh:       make(chan RPC),
		quitCh:      make(chan struct{}, 1),
	}

	if server.options.RPCProcessor == nil {
		server.options.RPCProcessor = server
	}

	if server.isValidator {
		go server.validatorLoop()
	}

	return server, nil
}

// Start starts the Server
func (s *Server) Start() {
	s.initTransports()

LOOP:
	for {
		select {
		case rpc := <-s.rpcCh:
			message, err := s.options.RPCDecodeFunc(rpc)

			if err != nil {
				s.options.Logger.Log("error", err)
			}

			if err = s.options.RPCProcessor.ProcessMessage(message); err != nil {
				s.options.Logger.Log("error", err)
			}
		case <-s.quitCh:
			break LOOP
		}
	}

	s.options.Logger.Log("msg", "server shutdown...")
}

func (s *Server) validatorLoop() {
	ticker := time.NewTicker(s.options.BlockTime)

	s.options.Logger.Log("msg", "starting validator loop...", "blocktime", s.options.BlockTime)

	for {
		<-ticker.C
		s.createNewBlock()
	}
}

// ProcessMessage checks message type and process it
func (s *Server) ProcessMessage(message *DecodedMessage) error {
	switch t := message.Data.(type) {
	case *core.Transaction:
		return s.processTransaction(t)
	}

	return nil
}

// processTransaction handles new transaction from network and adds it into memory pool
func (s *Server) processTransaction(transaction *core.Transaction) error {
	hash := transaction.Hash(core.TransactionHasher{})

	if s.memoryPool.Has(hash) {
		return nil
	}

	if err := transaction.Verify(); err != nil {
		return err
	}

	transaction.SetFirstSeen(time.Now().UnixNano())

	s.options.Logger.Log("msg", "adding new transaction to mempool", "hash", hash, "mempoolLen", s.memoryPool.Len())

	go func() {
		if err := s.broadcastTransaction(transaction); err != nil {
			s.options.Logger.Log("error", err)
		}
	}()

	return s.memoryPool.Add(transaction)
}

// broadcast broadcasts a payload to all transports
func (s *Server) broadcast(payload []byte) error {
	for _, transport := range s.options.Transports {
		if err := transport.Broadcast(payload); err != nil {
			return err
		}
	}

	return nil
}

// broadcastTransaction encodes transaction and broadcasts the message
func (s *Server) broadcastTransaction(transaction *core.Transaction) error {
	buf := &bytes.Buffer{}

	if err := transaction.Encode(core.NewGobTransactionEncoder(buf)); err != nil {
		return err
	}

	message := NewMessage(MessageTypeTransaction, buf.Bytes())

	return s.broadcast(message.Bytes())
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

// createNewBlock creates a new block
func (s *Server) createNewBlock() error {
	s.options.Logger.Log("msg", "creating a new block...")

	currentHeader, err := s.chain.GetHeader(s.chain.Height())
	if err != nil {
		return err
	}

	block, err := core.NewBlockFromPreviousHeader(currentHeader, nil)
	if err != nil {
		return err
	}

	if err = block.Sign(*s.options.PrivateKey); err != nil {
		return err
	}

	if err = s.chain.AddBlock(block); err != nil {
		return err
	}

	return nil
}

// genesisBlock returns a genesis block
func genesisBlock() *core.Block {
	header := &core.Header{
		Version:   1,
		DataHash:  types.Hash{},
		Timestamp: time.Now().UnixNano(),
		Height:    0,
	}

	return core.NewBlock(header, nil)
}
