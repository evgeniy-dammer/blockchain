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
	data := []byte{0x03, 0x0a, 0x31, 0x0c, 0x40, 0x0c, 0x40, 0x0c, 0x0d}
	virtualMachine := NewVirtualMachine(data)

	assert.Nil(t, virtualMachine.Run())

	//result := virtualMachine.stack.Pop()
	//assert.Equal(t, 4, result)
}
