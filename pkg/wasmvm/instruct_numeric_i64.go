package wasmvm

import (
	"encoding/binary"
	"errors"
	"math/bits"
)

// 0x42 const.i64: reads 8 octets little endian and pushes unit64 to stack
func CONST_I64(vm *VMState) error {
	const width = 1 + WIDTH_I64
	if vm.PC+width > uint64(len(vm.Memory)) {
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

// 0x7C add.i64: Pull two I64 words off stack, push I64 sum word on stack
func ADD_I64(vm *VMState) error {
	enough, collect := vm.ValueStack.HasAtLeastOfType(2, TYPE_I64)
	if !enough {
		return NewStackUnderflowErrorAndSetTrapReason(vm, "ADD_I64")
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		return NewStackCleanupErrorAndSetTrapReason(vm, "ADD_I64")
	}

	// Discard the overflow, effectively loops
	accumulator, _ := bits.Add64(collect[0].Value_I64, collect[1].Value_I64, 0)
	vm.ValueStack.PushInt64(accumulator)
	vm.PC += 1
	return nil
}

// 0x7D sub.i64: Pull two I64 words off stack, push I64 difference word on stack
func SUB_I64(vm *VMState) error {
	enough, collect := vm.ValueStack.HasAtLeastOfType(2, TYPE_I64)
	if !enough {
		return NewStackUnderflowErrorAndSetTrapReason(vm, "SUB_I64")
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		return NewStackCleanupErrorAndSetTrapReason(vm, "SUB_I64")
	}
	accumulator, _ := bits.Sub64(collect[0].Value_I64, collect[1].Value_I64, 0)

	vm.ValueStack.PushInt64(accumulator)
	vm.PC += 1
	return nil
}

func MUL_I64(vm *VMState) error {
	enough, collect := vm.ValueStack.HasAtLeastOfType(2, TYPE_I64)
	if !enough {
		return NewStackUnderflowErrorAndSetTrapReason(vm, "MUL_I64")
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		return NewStackCleanupErrorAndSetTrapReason(vm, "MUL_I64")
	}

	// While add and sub has sum/diff followed by carry/borrow
	// mul has producthi followed by productlo
	_, accumulator := bits.Mul64(collect[0].Value_I64, collect[1].Value_I64)
	vm.ValueStack.PushInt64(accumulator)
	vm.PC += 1
	return nil
}

// 0x80 div_u.i64: Pull two I64 words off stack, push I64 quotient word on stack
func DIVU_I64(vm *VMState) error {
	enough, collect := vm.ValueStack.HasAtLeastOfType(2, TYPE_I64)
	if !enough {
		return NewStackUnderflowErrorAndSetTrapReason(vm, "DIVU_I64")
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		return NewStackCleanupErrorAndSetTrapReason(vm, "DIVU_I64")
	}

	dividend := collect[0].Value_I64
	divisor := collect[1].Value_I64

	if divisor == 0 {
		vm.Trap = true
		vm.TrapReason = "DIVU_I64: Divide by Zero"
		return errors.New(vm.TrapReason)
	}

	accumulator, _ := bits.Div64(uint64(0), dividend, divisor)
	vm.ValueStack.PushInt64(accumulator)
	vm.PC += 1
	return nil
}
