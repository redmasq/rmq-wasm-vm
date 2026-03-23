package wasmvm

import (
	"encoding/binary"
	"math"
	"math/bits"
)

// 0x42 const.i64: reads 8 octets little endian and pushes unit64 to stack
func CONST_I64(vm *VMState) error {
	const width = 1 + WidthI64
	if vm.PC+width > uint64(len(vm.Memory)) {
		return vm.SetTrapError(&TrapError{
			Type:    TrapProgramCounterOutOfBounds,
			Op:      "CONST_I64",
			PC:      vm.PC,
			Message: "CONST_I64: Out of bounds",
			Meta: map[string]uint64{
				"width":      width,
				"memory_len": uint64(len(vm.Memory)),
			},
		})
	}
	span1 := vm.PC + 1
	span2 := span1 + WidthI64

	val := binary.LittleEndian.Uint64(vm.Memory[span1:span2])
	vm.ValueStack.PushInt64(val)
	vm.PC += width
	return nil
}

// 0x7C add.i64: Pull two I64 words off stack, push I64 sum word on stack
func ADD_I64(vm *VMState) error {
	enough, collect := vm.ValueStack.HasAtLeastOfType(2, TYPE_I64)
	if !enough {
		return NewStackUnderflowErrorAndSetTrap(vm, "ADD_I64")
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		return NewStackCleanupErrorAndSetTrap(vm, "ADD_I64")
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
		return NewStackUnderflowErrorAndSetTrap(vm, "SUB_I64")
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		return NewStackCleanupErrorAndSetTrap(vm, "SUB_I64")
	}
	accumulator, _ := bits.Sub64(collect[0].Value_I64, collect[1].Value_I64, 0)

	vm.ValueStack.PushInt64(accumulator)
	vm.PC += 1
	return nil
}

func MUL_I64(vm *VMState) error {
	enough, collect := vm.ValueStack.HasAtLeastOfType(2, TYPE_I64)
	if !enough {
		return NewStackUnderflowErrorAndSetTrap(vm, "MUL_I64")
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		return NewStackCleanupErrorAndSetTrap(vm, "MUL_I64")
	}

	// While add and sub has sum/diff followed by carry/borrow
	// mul has producthi followed by productlo
	_, accumulator := bits.Mul64(collect[0].Value_I64, collect[1].Value_I64)
	vm.ValueStack.PushInt64(accumulator)
	vm.PC += 1
	return nil
}

// 0x7F div_s.i64: Pull two I64 words off stack, push I64 quotient word on stack (signed)
func DIVS_I64(vm *VMState) error {
	enough, collect := vm.ValueStack.HasAtLeastOfType(2, TYPE_I64)
	if !enough {
		return NewStackUnderflowErrorAndSetTrap(vm, "DIVS_I64")
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		return NewStackCleanupErrorAndSetTrap(vm, "DIVS_I64")
	}

	dividend := collect[0].Value_I64
	divisor := collect[1].Value_I64
	sdividend := int64(dividend)
	sdivisor := int64(divisor)

	if sdivisor == 0 {
		return vm.SetTrapError(&TrapError{
			Type:    TrapDivideByZero,
			Op:      "DIVS_I64",
			PC:      vm.PC,
			Message: "DIVS_I64: Divide by Zero",
		})
	}

	if sdividend == math.MinInt64 && sdivisor == -1 {
		return vm.SetTrapError(&TrapError{
			Type:    TrapSignedDivisionOverflow,
			Op:      "DIVS_I64",
			PC:      vm.PC,
			Message: "DIVS_I64: Signed Division Overflow",
		})
	}

	// There is no wider signed integer available for a direct int64 divide here,
	// so we split the operation into three steps: determine the sign from the
	// original operands, convert each operand to its unsigned magnitude without
	// overflowing on MinInt64, then reuse bits.Div64 for the magnitude division
	// and convert the quotient back to two's-complement form if the result is
	// supposed to be negative.
	negative := (sdividend < 0) != (sdivisor < 0)

	var absDividend uint64
	if sdividend < 0 {
		absDividend = uint64(-(sdividend + 1))
		absDividend++
	} else {
		absDividend = uint64(sdividend)
	}

	var absDivisor uint64
	if sdivisor < 0 {
		absDivisor = uint64(-(sdivisor + 1))
		absDivisor++
	} else {
		absDivisor = uint64(sdivisor)
	}

	accumulator, _ := bits.Div64(0, absDividend, absDivisor)
	if negative {
		accumulator = ^accumulator + 1
	}

	vm.ValueStack.PushInt64(accumulator)
	vm.PC += 1
	return nil
}

// 0x80 div_u.i64: Pull two I64 words off stack, push I64 quotient word on stack
func DIVU_I64(vm *VMState) error {
	enough, collect := vm.ValueStack.HasAtLeastOfType(2, TYPE_I64)
	if !enough {
		return NewStackUnderflowErrorAndSetTrap(vm, "DIVU_I64")
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		return NewStackCleanupErrorAndSetTrap(vm, "DIVU_I64")
	}

	dividend := collect[0].Value_I64
	divisor := collect[1].Value_I64

	if divisor == 0 {
		return vm.SetTrapError(&TrapError{
			Type:    TrapDivideByZero,
			Op:      "DIVU_I64",
			PC:      vm.PC,
			Message: "DIVU_I64: Divide by Zero",
		})
	}

	accumulator, _ := bits.Div64(uint64(0), dividend, divisor)
	vm.ValueStack.PushInt64(accumulator)
	vm.PC += 1
	return nil
}
