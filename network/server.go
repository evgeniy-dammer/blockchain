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
		options.Logger = log.With(options.Logger, "ID", options.ID)
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

	// tr := s.Transports[0].(*LocalTransport)

	// fmt.Printf("%+v\n", tr.peers)
	// for _, tr := range s.Transports {
	// 	if err := s.sendGetStatusMessage(tr); err != nil {
	// 		s.Logger.Log("send get status error", err)
	// 	}
	// }

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
	}

	return nil
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

	if err := tr.SendMessage(tr.Address(), msg.Bytes()); err != nil {
		return err
	}

	return nil
}

// processStatusMessage
func (s *Server) processStatusMessage(from NetworkAddress, data *StatusMessage) error {
	fmt.Printf("=> received GetStatus response msg from %s => %+v\n", from, data)

	return nil
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

// broadcastBlock encodes a block and broadcasts the message
func (s *Server) broadcastBlock(block *core.Block) error {
	buf := &bytes.Buffer{}

	if err := block.Encode(core.NewGobBlockEncoder(buf)); err != nil {
		return err
	}

	message := NewMessage(MessageTypeBlock, buf.Bytes())

	return s.broadcast(message.Bytes())
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
