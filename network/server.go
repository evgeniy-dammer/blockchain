package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/evgeniy-dammer/blockchain/api"
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/evgeniy-dammer/blockchain/types"
	"github.com/go-kit/log"
	"net"
	"os"
	"sync"
	"time"
)

var defaultBlockTime = time.Second * 5

// ServerOptions
type ServerOptions struct {
	APIListenAddr string
	SeedNodes     []string
	ListenAddr    string
	TCPTransport  *TCPTransport
	ID            string
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	BlockTime     time.Duration
	PrivateKey    *crypto.PrivateKey
}

// Server
type Server struct {
	TCPTransport *TCPTransport
	peerCh       chan *TCPPeer
	mu           sync.RWMutex
	peerMap      map[net.Addr]*TCPPeer
	options      ServerOptions
	memoryPool   *TransactionPool
	chain        *core.Blockchain
	isValidator  bool
	rpcCh        chan RPC
	quitCh       chan struct{}
	txChan       chan *core.Transaction
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
		options.Logger = log.With(options.Logger, "addr", options.ID)
	}

	chain, err := core.NewBlockchain(options.Logger, genesisBlock())
	if err != nil {
		return nil, err
	}

	// Channel being used to communicate between the JSON RPC server
	// and the node that will process this message.
	txChan := make(chan *core.Transaction)

	// Only boot up the API server if the config has a valid port number.
	if len(options.APIListenAddr) > 0 {
		apiServerCfg := api.ServerConfig{
			Logger:     options.Logger,
			ListenAddr: options.APIListenAddr,
		}
		apiServer := api.NewServer(apiServerCfg, chain, txChan)
		go apiServer.Start()

		options.Logger.Log("msg", "JSON API server running", "port", options.APIListenAddr)
	}

	peerCh := make(chan *TCPPeer)
	tr := NewTCPTransport(options.ListenAddr, peerCh)

	server := &Server{
		TCPTransport: tr,
		peerCh:       peerCh,
		peerMap:      make(map[net.Addr]*TCPPeer),
		options:      options,
		chain:        chain,
		memoryPool:   NewTransactionPool(1000),
		isValidator:  options.PrivateKey != nil,
		rpcCh:        make(chan RPC),
		quitCh:       make(chan struct{}, 1),
		txChan:       txChan,
	}

	server.TCPTransport.peerCh = peerCh

	if server.options.RPCProcessor == nil {
		server.options.RPCProcessor = server
	}

	if server.isValidator {
		go server.validatorLoop()
	}

	return server, nil
}

func (s *Server) bootstrapNetwork() {
	for _, addr := range s.options.SeedNodes {
		fmt.Println("trying to connect to ", addr)

		go func(addr string) {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				fmt.Printf("could not connect to %+v\n", conn)
				return
			}

			s.peerCh <- &TCPPeer{
				conn: conn,
			}
		}(addr)
	}
}

// Start starts the Server
func (s *Server) Start() {
	s.TCPTransport.Start()

	time.Sleep(time.Second * 1)

	s.bootstrapNetwork()

	s.options.Logger.Log("msg", "accepting TCP connection on", "addr", s.options.ListenAddr, "id", s.options.ID)

LOOP:
	for {
		select {
		case peer := <-s.peerCh:
			s.peerMap[peer.conn.RemoteAddr()] = peer

			go peer.readLoop(s.rpcCh)

			if err := s.sendGetStatusMessage(peer); err != nil {
				s.options.Logger.Log("err", err)
				continue
			}

			s.options.Logger.Log("msg", "peer added to the server", "outgoing", peer.Outgoing, "addr", peer.conn.RemoteAddr())
		case tx := <-s.txChan:
			if err := s.processTransaction(tx); err != nil {
				s.options.Logger.Log("process TX error", err)
			}
		case rpc := <-s.rpcCh:
			message, err := s.options.RPCDecodeFunc(rpc)
			if err != nil {
				s.options.Logger.Log("RPC error", err)
				continue
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
	case *BlocksMessage:
		return s.processBlocksMessage(message.From, t)
	}

	return nil
}

// processGetBlocksMessage
func (s *Server) processGetBlocksMessage(from net.Addr, data *GetBlocksMessage) error {
	s.options.Logger.Log("msg", "received getBlocks message", "from", from)

	var (
		blocks    = []*core.Block{}
		ourHeight = s.chain.Height()
	)

	if data.To == 0 {
		for i := int(data.From); i <= int(ourHeight); i++ {
			block, err := s.chain.GetBlock(uint32(i))
			if err != nil {
				return err
			}

			blocks = append(blocks, block)
		}
	}

	blocksMsg := &BlocksMessage{
		Blocks: blocks,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(blocksMsg); err != nil {
		return err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	msg := NewMessage(MessageTypeBlocks, buf.Bytes())
	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
	}

	return peer.Send(msg.Bytes())
}

func (s *Server) processBlocksMessage(from net.Addr, data *BlocksMessage) error {
	s.options.Logger.Log("msg", "received BLOCKS!!!!!!!!", "from", from)

	for _, block := range data.Blocks {
		fmt.Printf("BlOCK with %+v\n", block.Header)
		if err := s.chain.AddBlock(block); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) processStatusMessage(from net.Addr, data *StatusMessage) error {
	s.options.Logger.Log("msg", "received STATUS message", "from", from)

	if data.CurrentHeight <= s.chain.Height() {
		s.options.Logger.Log("msg", "cannot sync blockHeight to low", "ourHeight", s.chain.Height(), "theirHeight", data.CurrentHeight, "addr", from)
		return nil
	}

	go s.requestBlocksLoop(from)

	return nil
}

// processGetStatusMessage
func (s *Server) processGetStatusMessage(from net.Addr, data *GetStatusMessage) error {
	s.options.Logger.Log("msg", "received getStatus message", "from", from)

	statusMessage := &StatusMessage{
		CurrentHeight: s.chain.Height(),
		ID:            s.options.ID,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(statusMessage); err != nil {
		return err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
	}

	msg := NewMessage(MessageTypeStatus, buf.Bytes())

	return peer.Send(msg.Bytes())
}

// sendGetStatusMessage normally Transport which is our own transport should do the trick.
func (s *Server) sendGetStatusMessage(peer *TCPPeer) error {
	var (
		getStatusMsg = new(GetStatusMessage)
		buf          = new(bytes.Buffer)
	)
	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeGetStatus, buf.Bytes())
	return peer.Send(msg.Bytes())
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

// TODO: Find a way to make sure we dont keep syncing when we are at the highest
// block height in the network.
func (s *Server) requestBlocksLoop(peer net.Addr) error {
	ticker := time.NewTicker(3 * time.Second)

	for {
		ourHeight := s.chain.Height()

		s.options.Logger.Log("msg", "requesting new blocks", "requesting height", ourHeight+1)

		// In this case we are 100% sure that the node has blocks heigher than us.
		getBlocksMessage := &GetBlocksMessage{
			From: ourHeight + 1,
			To:   0,
		}

		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(getBlocksMessage); err != nil {
			return err
		}

		s.mu.RLock()
		defer s.mu.RUnlock()

		msg := NewMessage(MessageTypeGetBlocks, buf.Bytes())
		peer, ok := s.peerMap[peer]
		if !ok {
			return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
		}

		if err := peer.Send(msg.Bytes()); err != nil {
			s.options.Logger.Log("error", "failed to send to peer", "err", err, "peer", peer)
		}

		<-ticker.C
	}
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
	s.mu.RLock()
	defer s.mu.RUnlock()

	for netAddr, peer := range s.peerMap {
		if err := peer.Send(payload); err != nil {
			fmt.Printf("peer send error => addr %s [err: %s]\n", netAddr, err)
		}
	}

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

// broadcastTransaction encodes a transaction and broadcasts the message
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

	privKey := crypto.GeneratePrivateKey()
	if err := b.Sign(privKey); err != nil {
		panic(err)
	}

	return b
}
