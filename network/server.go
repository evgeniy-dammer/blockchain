package network

import (
	"bytes"
	"fmt"
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/evgeniy-dammer/blockchain/crypto"
	"github.com/evgeniy-dammer/blockchain/types"
	"github.com/go-kit/log"
	"net"
	"os"
	"time"
)

var defaultBlockTime = time.Second * 5

// ServerOptions
type ServerOptions struct {
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
	peerMap      map[net.Addr]*TCPPeer
	options      ServerOptions
	memoryPool   *TransactionPool
	chain        *core.Blockchain
	isValidator  bool
	rpcCh        chan RPC
	quitCh       chan struct{}
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
			// TODO: add mutex PLZ!!!
			s.peerMap[peer.conn.RemoteAddr()] = peer

			go peer.readLoop(s.rpcCh)
			fmt.Printf("new peer => %+v\n", peer)
		case rpc := <-s.rpcCh:
			message, err := s.options.RPCDecodeFunc(rpc)
			if err != nil {
				s.options.Logger.Log("error", err)
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
		//return s.processGetStatusMessage(message.From, t)
	case *StatusMessage:
		//return s.processStatusMessage(message.From, t)
	case *GetBlocksMessage:
		return s.processGetBlocksMessage(message.From, t)
	}

	return nil
}

// processGetBlocksMessage
func (s *Server) processGetBlocksMessage(from net.Addr, data *GetBlocksMessage) error {
	fmt.Printf("got get blocks message => %+v\n", data)

	return nil
}

// processStatusMessage
/*func (s *Server) processStatusMessage(from net.Addr, data *StatusMessage) error {
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
}*/

// processGetStatusMessage
/*func (s *Server) processGetStatusMessage(from net.Addr, data *GetStatusMessage) error {
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
} */

// sendGetStatusMessage normally Transport which is our own transport should do the trick.
/*func (s *Server) sendGetStatusMessage(tr Transport) error {
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
}*/

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
	return b
}
