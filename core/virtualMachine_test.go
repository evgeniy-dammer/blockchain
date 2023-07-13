package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVirtualMachine_Run(t *testing.T) {
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	virtualMachine := NewVirtualMachine(data)

	assert.Nil(t, virtualMachine.Run())

	assert.Equal(t, byte(3), virtualMachine.stack[virtualMachine.stackPointer])
}
