package wasmvm

import (
	"errors"
)

type Instruction func(*VMState) error

func defaultInstructionMap() map[uint8]Instruction {
	return map[uint8]Instruction{
		0x00: NOP,
		0x01: ADD8, // Example additional instruction
		0x0B: END,  // End of function
	}
}

// NOP: No Operation
func NOP(vm *VMState) error {
	vm.PC++
	return nil
}

// ADD8: Add next two bytes, store in third byte (mem[PC+1] + mem[PC+2] -> mem[PC+3])
func ADD8(vm *VMState) error {
	if vm.PC+3 >= uint64(len(vm.Memory)) {
		vm.Trap = true
		vm.TrapReason = "ADD8: Out of bounds"
		return errors.New(vm.TrapReason)
	}
	a := vm.Memory[vm.PC+1]
	b := vm.Memory[vm.PC+2]
	vm.Memory[vm.PC+3] = a + b
	vm.PC += 4
	return nil
}

func END(vm *VMState) error {
	// Stub
	// TODO: Add handling when there are entries on the call stack for the execution frame
	vm.Trap = true
	vm.TrapReason = "End of final function"
	return nil
}
