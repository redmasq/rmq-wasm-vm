package wasmvm

import (
	"encoding/binary"
	"math"
	"math/bits"
)

// 0x43 const.i32: reads 4 octets little endian and pushes unit32 to stack
func CONST_I32(vm *VMState) error {
	const width = 1 + WidthI32
	if vm.PC+width > uint64(len(vm.Memory)) {
		return vm.SetTrapError(&TrapError{
			Type:    TrapProgramCounterOutOfBounds,
			Op:      "CONST_I32",
			PC:      vm.PC,
			Message: "CONST_I32: Out of bounds",
			Meta: map[string]uint64{
				"width":      width,
				"memory_len": uint64(len(vm.Memory)),
			},
		})
	}
	span1 := vm.PC + 1
	span2 := span1 + WidthI32

	val := binary.LittleEndian.Uint32(vm.Memory[span1:span2])
	vm.ValueStack.PushInt32(val)
	vm.PC += width
	return nil
}

// 0x6A add.i32: Pull two I32 words off stack, push I32 sum word on stack
func ADD_I32(vm *VMState) error {
	enough, collect := vm.ValueStack.HasAtLeastOfType(2, TYPE_I32)
	if !enough {
		return NewStackUnderflowErrorAndSetTrap(vm, "ADD_I32")
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		return NewStackCleanupErrorAndSetTrap(vm, "ADD_I32")
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
		return NewStackUnderflowErrorAndSetTrap(vm, "SUB_I32")
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		return NewStackCleanupErrorAndSetTrap(vm, "SUB_I32")
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
		return NewStackUnderflowErrorAndSetTrap(vm, "MUL_I32")
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		return NewStackCleanupErrorAndSetTrap(vm, "MUL_I32")
	}

	// While add and sub has sum/diff followed by carry/borrow
	// mul has producthi followed by productlo
	_, accumulator := bits.Mul32(collect[0].Value_I32, collect[1].Value_I32)
	vm.ValueStack.PushInt32(accumulator)
	vm.PC += 1
	return nil
}

// 0x6D div_s.i32: Pull two I32 words off stack, push I32 quotient word on stack (signed)
func DIVS_I32(vm *VMState) error {
	enough, collect := vm.ValueStack.HasAtLeastOfType(2, TYPE_I32)
	if !enough {
		return NewStackUnderflowErrorAndSetTrap(vm, "DIVS_I32")
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		return NewStackCleanupErrorAndSetTrap(vm, "DIVS_I32")
	}

	// Cheap cast for getting signed version
	// Note that the way the stack works is non-intuitive since
	// I use a slice... The top of the stack is at the end of the array
	sdividend := int32(collect[0].Value_I32)
	sdivisor := int32(collect[1].Value_I32)

	if sdivisor == 0 {
		return vm.SetTrapError(&TrapError{
			Type:    TrapDivideByZero,
			Op:      "DIVS_I32",
			PC:      vm.PC,
			Message: "DIVS_I32: Divide by Zero",
		})
	}

	// Referencing https://webassembly.github.io/spec/core/exec/numerics.html#op-idiv-s
	// This line suggests to me that overflow should be rejected
	// Else if j_2 divided by j_2 is 2^{N-1}, then the result is undefined.
	// For now, I'll trap like Divide by Zero

	if sdividend == math.MinInt32 && sdivisor == -1 { // This should be overflow
		return vm.SetTrapError(&TrapError{
			Type:    TrapSignedDivisionOverflow,
			Op:      "DIVS_I32",
			PC:      vm.PC,
			Message: "DIVS_I32: Signed Division Overflow",
		})
	}

	// I took a gander at Knunth vol 2, p 276. It looks like since 2's compliment
	// I can just do the math as unsigned after accounting for overflow and then
	// set the sign based on the effective signs of the original
	//
	// The whole thing seemed more complicated than necessary, so I'm just doing division
	// using int64 for int32, and coercing it back to uint32.
	// This means I'll need to look at it again for div_s.i64, since I haven't found int128
	// An additional look at bits.Div64 seems to suggest it's just Knuth Algorith D
	// Or some variant there of, see page 272 of 3rd edition.
	// But here, no need to work harder than need be

	accumulator := uint32(int64(sdividend) / int64(sdivisor))

	// As usual, I have to use uint32 for the stack itself since it's sign agnostic.
	// Will need to make sure there is a unit test to ensure the high bit round trips
	vm.ValueStack.PushInt32(uint32(accumulator))
	vm.PC += 1
	return nil
}

// 0x6E div_u.i32: Pull two I32 words off stack, push I32 quotient word on stack (unsigned)
func DIVU_I32(vm *VMState) error {
	enough, collect := vm.ValueStack.HasAtLeastOfType(2, TYPE_I32)
	if !enough {
		return NewStackUnderflowErrorAndSetTrap(vm, "DIVU_I32")
	}

	// I'm not even sure how I can write an unit test for this
	// one, especially since there is no multithreading and
	// the instruction treats it as an atomic operation
	if !vm.ValueStack.Drop(2, true) {
		return NewStackCleanupErrorAndSetTrap(vm, "DIVU_I32")
	}

	dividend := collect[0].Value_I32
	divisor := collect[1].Value_I32

	if divisor == 0 {
		return vm.SetTrapError(&TrapError{
			Type:    TrapDivideByZero,
			Op:      "DIVU_I32",
			PC:      vm.PC,
			Message: "DIVU_I32: Divide by Zero",
		})
	}

	accumulator, _ := bits.Div32(uint32(0), dividend, divisor)
	vm.ValueStack.PushInt32(accumulator)
	vm.PC += 1
	return nil
}
