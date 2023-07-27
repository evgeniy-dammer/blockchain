package core

import "encoding/binary"

type Instruction byte

const (
	InstructionPushInt  Instruction = 0x0a
	InstructionAdd      Instruction = 0x0b
	InstructionPushByte Instruction = 0x0c
	InstructionPack     Instruction = 0x0d
	InstructionSub      Instruction = 0x0e
	InstructionStore    Instruction = 0x0f
	InstructionGet      Instruction = 0xae
	InstructionMul      Instruction = 0xea
	InstructionDiv      Instruction = 0xfd
)

// Stack
type Stack struct {
	data         []any
	stackPointer int
}

// NewStack is a constructor for the Stack
func NewStack(size int) *Stack {
	return &Stack{
		data:         make([]any, size),
		stackPointer: 0,
	}
}

// Push pushes the given value at the end of the stack
func (s *Stack) Push(value any) {
	s.data[s.stackPointer] = value
	s.stackPointer++
}

// Pop pops the value from the start of the stack
func (s *Stack) Pop() any {
	value := s.data[0]
	s.data = append(s.data[:0], s.data[1:]...)
	s.stackPointer--

	return value
}

// VirtualMachine
type VirtualMachine struct {
	data               []byte
	instructionPointer int
	stack              *Stack
	contractState      *State
}

// NewVirtualMachine is a constructor for the VirtualMachine
func NewVirtualMachine(data []byte, contractState *State) *VirtualMachine {
	return &VirtualMachine{
		data:               data,
		instructionPointer: 0,
		stack:              NewStack(128),
		contractState:      contractState,
	}
}

// Run runs the virtual machine
func (vm *VirtualMachine) Run() error {
	for {
		instruction := Instruction(vm.data[vm.instructionPointer])

		if err := vm.Exec(instruction); err != nil {
			return err
		}

		vm.instructionPointer++

		if vm.instructionPointer > len(vm.data)-1 {
			break
		}
	}

	return nil
}

// Exec executes the instruction
func (vm *VirtualMachine) Exec(instruction Instruction) error {
	switch instruction {
	case InstructionPushInt:
		vm.stack.Push(int(vm.data[vm.instructionPointer-1]))
	case InstructionPushByte:
		vm.stack.Push(byte(vm.data[vm.instructionPointer-1]))
	case InstructionPack:
		n := vm.stack.Pop().(int)
		b := make([]byte, n)

		for i := 0; i < n; i++ {
			b[i] = vm.stack.Pop().(byte)
		}

		vm.stack.Push(b)
	case InstructionSub:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a - b

		vm.stack.Push(c)
	case InstructionAdd:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a + b

		vm.stack.Push(c)
	case InstructionStore:
		var (
			key             = vm.stack.Pop().([]byte)
			value           = vm.stack.Pop()
			serializedValue []byte
		)

		switch v := value.(type) {
		case int:
			serializedValue = serializeInt64(int64(v))
		default:
			panic("TODO: unknown type")
		}

		vm.contractState.Put(key, serializedValue)
	case InstructionGet:
		key := vm.stack.Pop().([]byte)

		value, err := vm.contractState.Get(key)
		if err != nil {
			return err
		}

		vm.stack.Push(value)
	case InstructionMul:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a * b

		vm.stack.Push(c)
	case InstructionDiv:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a / b

		vm.stack.Push(c)
	}

	return nil
}

func serializeInt64(value int64) []byte {
	buf := make([]byte, 8)

	binary.LittleEndian.PutUint64(buf, uint64(value))

	return buf
}

func deserializeInt64(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
}
