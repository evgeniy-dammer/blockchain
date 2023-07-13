package core

type Instruction byte

const (
	InstructionPush Instruction = 0x0a
	instructionAdd  Instruction = 0x0b
)

// VirtualMachine
type VirtualMachine struct {
	data               []byte
	instructionPointer int
	stack              []byte
	stackPointer       int
}

// NewVirtualMachine is a constructor for the VirtualMachine
func NewVirtualMachine(data []byte) *VirtualMachine {
	return &VirtualMachine{
		data:               data,
		instructionPointer: 0,
		stack:              make([]byte, 1024),
		stackPointer:       -1,
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
	case InstructionPush:
		vm.pushStack(vm.data[vm.instructionPointer-1])
	case instructionAdd:
		a := vm.stack[0]
		b := vm.stack[1]
		c := a + b

		vm.pushStack(c)
	}

	return nil
}

// pushStack pushes the instruction into the stack
func (vm *VirtualMachine) pushStack(b byte) {
	vm.stackPointer++
	vm.stack[vm.stackPointer] = b
}
