package wasmvm_test

import (
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
)

func TestCONST_I64(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 10,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.Memory[0] = 0x42 // Opcode
	vm.Memory[1] = 0xEF
	vm.Memory[2] = 0xCD
	vm.Memory[3] = 0xAB
	vm.Memory[4] = 0x89
	vm.Memory[5] = 0x67
	vm.Memory[6] = 0x45
	vm.Memory[7] = 0x23
	vm.Memory[8] = 0x01
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, success := vm.ValueStack.Pop()
	assert.True(t, success)
	assert.Equal(t, uint64(9), vm.PC)
	assert.Equal(t, uint64(0x0123456789ABCDEF), val.Value_I64)

}

func TestCONST_I64_OOB(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 4,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)

	vm.Memory[0] = 0x42 // Opcode
	vm.Memory[1] = 0x78
	vm.Memory[2] = 0x56
	vm.PC = 0
	err = vm.Step()
	assert.Error(t, err)
	assert.Equal(t, int(0), vm.ValueStack.Size())
	assert.True(t, vm.Trap)
	assert.Equal(t, "CONST_I64: Out of bounds", vm.TrapReason)

}

func TestADD_I64_NotEnoughStack(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	// opcode
	vm.Memory[0] = 0x7C
	vm.PC = 0
	err = vm.Step()
	assert.Error(t, err)
	assert.Equal(t, int(0), vm.ValueStack.Size())
	assert.True(t, vm.Trap)
	assert.Equal(t, "ADD_I64: Stack Underflow", vm.TrapReason)
}

func TestADD_I64_SmallNumbers(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)

	vm.ValueStack.PushInt64(5)
	vm.ValueStack.PushInt64(7)
	// opcode
	vm.Memory[0] = 0x7C
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, success := vm.ValueStack.Pop()
	assert.True(t, success)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, uint64(12), val.Value_I64)
	assert.Equal(t, int(0), vm.ValueStack.Size())
}

func TestADD_I64_OverflowWrap(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt64(^uint64(0)) // I didn't feel like typing a bunch of f's
	vm.ValueStack.PushInt64(2)
	// opcode
	vm.Memory[0] = 0x7C
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, success := vm.ValueStack.Pop()
	assert.True(t, success)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, uint64(1), val.Value_I64)
	assert.Equal(t, int(0), vm.ValueStack.Size())
}

func TestSUB_I64_NotEnoughStack(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	// opcode
	vm.Memory[0] = 0x7D
	vm.PC = 0
	err = vm.Step()
	assert.Error(t, err)
	assert.Equal(t, int(0), vm.ValueStack.Size())
	assert.True(t, vm.Trap)
	assert.Equal(t, "SUB_I64: Stack Underflow", vm.TrapReason)
}

func TestSUB_I64_SmallNumbers(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt64(7)
	vm.ValueStack.PushInt64(5)
	// opcode
	vm.Memory[0] = 0x7D
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, success := vm.ValueStack.Pop()
	assert.True(t, success)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, uint64(2), val.Value_I64)
	assert.Equal(t, int(0), vm.ValueStack.Size())
}

func TestSUB_I64_OverflowWrap(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt64(1)
	vm.ValueStack.PushInt64(2)
	// opcode
	vm.Memory[0] = 0x7D
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, success := vm.ValueStack.Pop()
	assert.True(t, success)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, ^uint64(0), val.Value_I64) // Short hand for all f;s
	assert.Equal(t, int(0), vm.ValueStack.Size())
}
