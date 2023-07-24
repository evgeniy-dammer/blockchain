package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
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
	Transport     Transport
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
		options.Logger = log.With(options.Logger, "addr", options.Transport.Address())
	}

	chain, err := core.NewBlockchain(options.Logger, genesisBlock())
	if err != nil {
		return nil, err
	}

	server := &Server{
		options:     options,
		memoryPool:  NewTransactionPool(1000),
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

	server.boostrapNodes()

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
				if err != core.ErrBlockKnown {
					s.options.Logger.Log("error", err)
				}
			}
		case <-s.quitCh:
			break LOOP
		}
	}

	s.options.Logger.Log("msg", "server shutdown...")
}

// validatorLoop runs a creating new block loop if node is validator
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
	case *core.Block:
		return s.processBlock(t)
	case *GetStatusMessage:
		return s.processGetStatusMessage(message.From, t)
	case *StatusMessage:
		return s.processStatusMessage(message.From, t)
	case *GetBlocksMessage:
		return s.processGetBlocksMessage(message.From, t)
	}

	return nil
}

// processGetBlocksMessage
func (s *Server) processGetBlocksMessage(from NetworkAddress, data *GetBlocksMessage) error {
	panic("here")
	fmt.Printf("got get blocks message => %+v\n", data)

	return nil
}

// processStatusMessage
func (s *Server) processStatusMessage(from NetworkAddress, data *StatusMessage) error {
	if data.CurrentHeight <= s.chain.Height() {
		s.options.Logger.Log("msg", "cannot sync blockHeight to low", "ourHeight", s.chain.Height(), "theirHeight", data.CurrentHeight, "addr", from)
		return nil
	}

	// In this case we are 100% sure that the node has blocks higher than us.
	getBlocksMessage := &GetBlocksMessage{
		From: s.chain.Height(),
		To:   0,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(getBlocksMessage); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeGetBlocks, buf.Bytes())

	return s.options.Transport.SendMessage(from, msg.Bytes())
}

// processGetStatusMessage
func (s *Server) processGetStatusMessage(from NetworkAddress, data *GetStatusMessage) error {
	fmt.Printf("=> received Getstatus msg from %s => %+v\n", from, data)

	statusMessage := &StatusMessage{
		CurrentHeight: s.chain.Height(),
		ID:            s.options.ID,
	}

	buf := new(bytes.Buffer)

	if err := gob.NewEncoder(buf).Encode(statusMessage); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeStatus, buf.Bytes())

	return s.options.Transport.SendMessage(from, msg.Bytes())
}

// sendGetStatusMessage normally Transport which is our own transport should do the trick.
func (s *Server) sendGetStatusMessage(tr Transport) error {
	var (
		getStatusMsg = new(GetStatusMessage)
		buf          = new(bytes.Buffer)
	)

	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeGetStatus, buf.Bytes())

	if err := s.options.Transport.SendMessage(tr.Address(), msg.Bytes()); err != nil {
		return err
	}

	return nil
}

// processTransaction handles new transaction from network and adds it into memory pool
func (s *Server) processTransaction(transaction *core.Transaction) error {
	hash := transaction.Hash(core.TransactionHasher{})

	if s.memoryPool.Contains(hash) {
		return nil
	}

	if err := transaction.Verify(); err != nil {
		return err
	}

	//s.options.Logger.Log("msg", "adding new transaction to mempool", "hash", hash, "mempoolPending", s.memoryPool.PendingCount())

	go func() {
		if err := s.broadcastTransaction(transaction); err != nil {
			s.options.Logger.Log("error", err)
		}
	}()

	s.memoryPool.Add(transaction)

	return nil
}

// processBlock adds block to servers chain and broadcasts the block
func (s *Server) processBlock(b *core.Block) error {
	if err := s.chain.AddBlock(b); err != nil {
		return err
	}

	go s.broadcastBlock(b)

	return nil
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

// boostrapNodes connects to transports and sends GetStatusMessage
func (s *Server) boostrapNodes() {
	for _, tr := range s.options.Transports {
		if s.options.Transport.Address() != tr.Address() {
			if err := s.options.Transport.Connect(tr); err != nil {
				s.options.Logger.Log("error", "could not connect to remote", "err", err)
			}
			s.options.Logger.Log("msg", "connect to remote", "we", s.options.Transport.Address(), "addr", tr.Address())

			// Send the getStatusMessage so we can sync (if needed)
			if err := s.sendGetStatusMessage(tr); err != nil {
				s.options.Logger.Log("error", "sendGetStatusMessage", "err", err)
			}
		}
	}
}

// broadcastBlock encodes a block and broadcasts the message
func (s *Server) broadcastBlock(block *core.Block) error {
	buf := &bytes.Buffer{}

	if err := block.Encode(core.NewGobBlockEncoder(buf)); err != nil {
		return err
	}

	message := NewMessage(MessageTypeBlock, buf.Bytes())

	return s.broadcast(message.Bytes())
}

// broadcastTransaction encodes a transaction and broadcasts the message
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
	currentHeader, err := s.chain.GetHeader(s.chain.Height())
	if err != nil {
		return err
	}

	// For now, we are going to use all transactions, that are in the memory pool.
	// Later on when we know the internal structure of our transaction we will
	// implement some kind of complexity function to determine how many transactions
	// can be included to the block
	transactions := s.memoryPool.Pending()

	block, err := core.NewBlockFromPreviousHeader(currentHeader, transactions)
	if err != nil {
		return err
	}

	if err = block.Sign(*s.options.PrivateKey); err != nil {
		return err
	}

	if err = s.chain.AddBlock(block); err != nil {
		return err
	}

	s.memoryPool.ClearPending()

	go s.broadcastBlock(block)

	return nil
}

// genesisBlock returns a genesis block
func genesisBlock() *core.Block {
	header := &core.Header{
		Version:   1,
		DataHash:  types.Hash{},
		Timestamp: 000000,
		Height:    0,
	}

	b, _ := core.NewBlock(header, nil)
	return b
}
