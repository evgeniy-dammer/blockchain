package network

import (
	"bytes"
	"fmt"
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/rs/zerolog/log"
	"time"
)

var defaultBlockTime = time.Second * 5

// ServerOptions
type ServerOptions struct {
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
	isValidator bool
	rpcCh       chan RPC
	quitCh      chan struct{}
}

// NewServer is a constructor for the Server
func NewServer(options ServerOptions) *Server {
	if options.BlockTime == time.Duration(0) {
		options.BlockTime = defaultBlockTime
	}

	if options.RPCDecodeFunc == nil {
		options.RPCDecodeFunc = DefaultRPCDecodeFunc
	}

	server := &Server{
		options:     options,
		memoryPool:  NewTransactionPool(),
		isValidator: options.PrivateKey != nil,
		rpcCh:       make(chan RPC),
		quitCh:      make(chan struct{}, 1),
	}

	if server.options.RPCProcessor == nil {
		server.options.RPCProcessor = server
	}

	return server
}

// Start starts the Server
func (s *Server) Start() {
	s.initTransports()
	ticker := time.NewTicker(s.options.BlockTime)

LOOP:
	for {
		select {
		case rpc := <-s.rpcCh:
			message, err := s.options.RPCDecodeFunc(rpc)

			if err != nil {
				log.Error().Msgf("rpc decoding failed: %s", err)
			}

			if err = s.options.RPCProcessor.ProcessMessage(message); err != nil {
				log.Error().Msgf("message processing failed: %s", err)
			}
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
		log.Info().Msgf("transaction with hash %s is already in mempool", hash)

		return nil
	}

	if err := transaction.Verify(); err != nil {
		return err
	}

	transaction.SetFirstSeen(time.Now().UnixNano())

	log.Info().Msgf("adding new transaction with hash %s into mempool", hash)

	go func() {
		if err := s.broadcastTransaction(transaction); err != nil {
			log.Error().Msgf("transaction broadcasting failed: %s", err)
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
