package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/evgeniy-dammer/blockchain/core"
	"github.com/rs/zerolog/log"
	"io"
)

const (
	MessageTypeTransaction MessageType = 0x1
	MessageTypeBlock       MessageType = 0x2
	MessageTypeGetBlocks   MessageType = 0x3
)

// RPC
type RPC struct {
	From    NetworkAddress
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
	From NetworkAddress
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
	default:
		return nil, fmt.Errorf("invalid message type %x", message.Type)
	}
}

// RPCProcessor
type RPCProcessor interface {
	ProcessMessage(message *DecodedMessage) error
}
