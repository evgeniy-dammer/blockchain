package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStack_Pop(t *testing.T) {
	stack := NewStack(128)

	stack.Push(1)
	stack.Push(2)

	value := stack.Pop()

	assert.Equal(t, value, 1)

	value = stack.Pop()

	assert.Equal(t, value, 2)
}

func TestStack_PushBytes(t *testing.T) {
	s := NewStack(100)
	s.Push(2)
	s.Push(0x61)
	s.Push(0x61)
}

func TestVirtualMachine_Run(t *testing.T) {
	//data := []byte{0x02, 0x0a, 0x02, 0x0a, 0x0b}

	data := []byte{0x03, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x0d, 0x05, 0x0a, 0x0f}

	contractState := NewState()
	virtualMachine := NewVirtualMachine(data, contractState)

	assert.Nil(t, virtualMachine.Run())

	valueBytes, err := contractState.Get([]byte("FOO"))
	value := deserializeInt64(valueBytes)
	assert.Nil(t, err)
	assert.Equal(t, value, int64(5))
}
