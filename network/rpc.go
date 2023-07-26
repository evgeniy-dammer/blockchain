package network

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/rs/zerolog/log"
	"io"
	"net"
)

const (
	MessageTypeTransaction MessageType = 0x1
	MessageTypeBlock       MessageType = 0x2
	MessageTypeGetBlocks   MessageType = 0x3
	MessageTypeStatus      MessageType = 0x4
	MessageTypeGetStatus   MessageType = 0x5
	MessageTypeBlocks      MessageType = 0x6
)

func init() {
	gob.Register(elliptic.P256())
}

// RPC
type RPC struct {
	From    net.Addr
	Payload io.Reader
}

// MessageType
type MessageType byte

// Message
type Message struct {
	Type MessageType
	Data []byte
}

// NewMessage is a constructor for the Message
func NewMessage(messageType MessageType, data []byte) *Message {
	return &Message{Type: messageType, Data: data}
}

// Bytes returns the message as a slice of bytes
func (m *Message) Bytes() []byte {
	buf := &bytes.Buffer{}

	if err := gob.NewEncoder(buf).Encode(m); err != nil {
		return nil
	}

	return buf.Bytes()
}

// DecodedMessage
type DecodedMessage struct {
	From net.Addr
	Data any
}

// RPCDecodeFunc
type RPCDecodeFunc func(RPC) (*DecodedMessage, error)

// DefaultRPCDecodeFunc returns a decoded message fetched from peers
func DefaultRPCDecodeFunc(rpc RPC) (*DecodedMessage, error) {
	message := Message{}

	if err := gob.NewDecoder(rpc.Payload).Decode(&message); err != nil {
		return nil, fmt.Errorf("failed to decode message from %s: %s", rpc.From, err)
	}

	log.Info().Msgf("receives a new message from %s with %x type", rpc.From, message.Type)

	switch message.Type {
	case MessageTypeTransaction:
		transaction := new(core.Transaction)

		if err := transaction.Decode(core.NewGobTransactionDecoder(bytes.NewReader(message.Data))); err != nil {
			return nil, err
		}

		return &DecodedMessage{
			From: rpc.From,
			Data: transaction,
		}, nil
	case MessageTypeBlock:
		block := new(core.Block)
		if err := block.Decode(core.NewGobBlockDecoder(bytes.NewReader(message.Data))); err != nil {
			return nil, err
		}

		return &DecodedMessage{
			From: rpc.From,
			Data: block,
		}, nil
	case MessageTypeGetStatus:
		return &DecodedMessage{
			From: rpc.From,
			Data: &GetStatusMessage{},
		}, nil

	case MessageTypeStatus:
		statusMessage := new(StatusMessage)
		if err := gob.NewDecoder(bytes.NewReader(message.Data)).Decode(statusMessage); err != nil {
			return nil, err
		}

		return &DecodedMessage{
			From: rpc.From,
			Data: statusMessage,
		}, nil
	case MessageTypeGetBlocks:
		getBlocks := new(GetBlocksMessage)
		if err := gob.NewDecoder(bytes.NewReader(message.Data)).Decode(getBlocks); err != nil {
			return nil, err
		}

		return &DecodedMessage{
			From: rpc.From,
			Data: getBlocks,
		}, nil

	case MessageTypeBlocks:
		blocks := new(BlocksMessage)
		if err := gob.NewDecoder(bytes.NewReader(message.Data)).Decode(blocks); err != nil {
			return nil, err
		}

		return &DecodedMessage{
			From: rpc.From,
			Data: blocks,
		}, nil

	default:
		return nil, fmt.Errorf("invalid message type %x", message.Type)
	}
}

// RPCProcessor
type RPCProcessor interface {
	ProcessMessage(message *DecodedMessage) error
}
