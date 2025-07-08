package wasmvm

import (
	"encoding/binary"
	"errors"
	"math/bits"
)

// 0x42 const.i64: reads 8 octets little endian and pushes unit64 to stack
func CONST_I64(vm *VMState) error {
	const width = 1 + WIDTH_I64
	if vm.PC+width >= uint64(len(vm.Memory)) {
		vm.Trap = true
		vm.TrapReason = "CONST_I64: Out of bounds"
		return errors.New(vm.TrapReason)
	}
	span1 := vm.PC + 1
	span2 := span1 + WIDTH_I64

	val := binary.LittleEndian.Uint64(vm.Memory[span1:span2])
	vm.ValueStack.PushInt64(val)
	vm.PC += width
	return nil
}

// 0x7C add.i32: Pull two I64 words off stack, push I64 sum word on stack
func ADD_I64(vm *VMState) error {
	enough, collect := vm.ValueStack.HasAtLeastOfType(2, TYPE_I64)
	if !enough {
		vm.Trap = true
		vm.TrapReason = "ADD_I64: Stack Underflow"
		return errors.New(vm.TrapReason)
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		vm.Trap = true
		vm.TrapReason = "ADD_I64: Stack Cleanup Error"
		return errors.New(vm.TrapReason)
	}

	// Discard the overflow, effectively loops
	accumalator, _ := bits.Add64(collect[0].Value_I64, collect[1].Value_I64, 0)
	vm.ValueStack.PushInt64(accumalator)
	vm.PC += 1
	return nil
}

// 0x7D sub.i32: Pull two I64 words off stack, push I64 difference word on stack
func SUB_I64(vm *VMState) error {
	enough, collect := vm.ValueStack.HasAtLeastOfType(2, TYPE_I64)
	if !enough {
		vm.Trap = true
		vm.TrapReason = "SUB_I64: Stack Underflow"
		return errors.New(vm.TrapReason)
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		vm.Trap = true
		vm.TrapReason = "SUB_I64: Stack Cleanup Error"
		return errors.New(vm.TrapReason)
	}
	accumalator, _ := bits.Sub64(collect[0].Value_I64, collect[1].Value_I64, 0)

	vm.ValueStack.PushInt64(accumalator)
	vm.PC += 1
	return nil
}
