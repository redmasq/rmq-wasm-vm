package wasmvm

import (
	"encoding/binary"
	"errors"
	"math/bits"
)

// 0x43 const.i32: reads 4 octets little endian and pushes unit32 to stack
func CONST_I32(vm *VMState) error {
	const width = 1 + WIDTH_I32
	if vm.PC+width >= uint64(len(vm.Memory)) {
		vm.Trap = true
		vm.TrapReason = "CONST_I32: Out of bounds"
		return errors.New(vm.TrapReason)
	}
	span1 := vm.PC + 1
	span2 := span1 + WIDTH_I32

	val := binary.LittleEndian.Uint32(vm.Memory[span1:span2])
	vm.ValueStack.PushInt32(val)
	vm.PC += width
	return nil
}

// 0x6A add.i32: Pull two I32 words off stack, push I32 sum word on stack
func ADD_I32(vm *VMState) error {
	enough, collect := vm.ValueStack.HasAtLeastOfType(2, TYPE_I32)
	if !enough {
		return NewStackUnderflowErrorAndSetTrapReason(vm, "ADD_I32")
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		return NewStackCleanupErrorAndSetTrapReason(vm, "ADD_I32")
	}
	accumulator, _ := bits.Add32(collect[0].Value_I32, collect[1].Value_I32, 0)
	vm.ValueStack.PushInt32(accumulator)
	vm.PC += 1
	return nil
}

// 0x6B sub.i32: Pull two I32 words off stack, push I32 difference word on stack
func SUB_I32(vm *VMState) error {
	enough, collect := vm.ValueStack.HasAtLeastOfType(2, TYPE_I32)
	if !enough {
		return NewStackUnderflowErrorAndSetTrapReason(vm, "SUB_I32")
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		return NewStackCleanupErrorAndSetTrapReason(vm, "SUB_I32")
	}
	accumulator, _ := bits.Sub32(collect[0].Value_I32, collect[1].Value_I32, 0)
	vm.ValueStack.PushInt32(accumulator)
	vm.PC += 1
	return nil
}

// 0x6C mul.i32: Pull two I32 words off stack, push I32 product word on stack
func MUL_I32(vm *VMState) error {
	enough, collect := vm.ValueStack.HasAtLeastOfType(2, TYPE_I32)
	if !enough {
		return NewStackUnderflowErrorAndSetTrapReason(vm, "MUL_I32")
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		return NewStackCleanupErrorAndSetTrapReason(vm, "MUL_I32")
	}

	// While add and sub has sum/diff followed by carry/borrow
	// mul has producthi followed by productlo
	_, accumulator := bits.Mul32(collect[0].Value_I32, collect[1].Value_I32)
	vm.ValueStack.PushInt32(accumulator)
	vm.PC += 1
	return nil
}
