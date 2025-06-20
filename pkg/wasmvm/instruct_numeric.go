package wasmvm

import (
	"encoding/binary"
	"errors"
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
	span2 := span1 + WIDTH_F32

	val := binary.LittleEndian.Uint32(vm.Memory[span1:span2])
	vm.ValueStack.PushInt32(val)
	vm.PC += width
	return nil
}

// 0x6A add.i32: Pull two I32 words off stack, push I32 sum word on stack
func ADD_I32(vm *VMState) error {
	enough, collect := vm.ValueStack.HasAtLeastOfType(2, TYPE_I32)
	if !enough {
		vm.Trap = true
		vm.TrapReason = "ADD_I32: Stack Underflow"
		return errors.New(vm.TrapReason)
	}

	if !vm.ValueStack.Drop(2, true) {
		vm.Trap = true
		vm.TrapReason = "ADD_I32: Stack Cleanup Error"
		return errors.New(vm.TrapReason)
	}
	accumalator := uint32(uint64(collect[0].Value_I32) + uint64(collect[1].Value_I32))
	vm.ValueStack.PushInt32(accumalator)
	vm.PC += 1
	return nil
}
