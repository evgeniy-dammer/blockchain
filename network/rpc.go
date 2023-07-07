package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/evgeniy-dammer/blockchain/core"
	"io"
)

const (
	MessageTypeTransaction MessageType = 0x1
	MessageTypeBlock
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

// RPCHandler
type RPCHandler interface {
	HandleRPC(rpc RPC) error
}

// DefaultRPCHandler
type DefaultRPCHandler struct {
	processor RPCProcessor
}

// NewDefaultRPCHandler is a constructor for the DefaultRPCHandler
func NewDefaultRPCHandler(processor RPCProcessor) *DefaultRPCHandler {
	return &DefaultRPCHandler{processor: processor}
}

// HandleRPC handles the rpc, decodes id and forvards for processing
func (h *DefaultRPCHandler) HandleRPC(rpc RPC) error {
	message := Message{}

	if err := gob.NewDecoder(rpc.Payload).Decode(&message); err != nil {
		return fmt.Errorf("failed to decode message from %s: %s", rpc.From, err)
	}

	switch message.Type {
	case MessageTypeTransaction:
		transaction := new(core.Transaction)

		if err := transaction.Decode(core.NewGobTransactionDecoder(bytes.NewReader(message.Data))); err != nil {
			return err
		}

		return h.processor.ProcessTransaction(rpc.From, transaction)
	default:
		return fmt.Errorf("invalid message type %x", message.Type)
	}
}

// RPCProcessor
type RPCProcessor interface {
	ProcessTransaction(address NetworkAddress, transaction *core.Transaction) error
}
