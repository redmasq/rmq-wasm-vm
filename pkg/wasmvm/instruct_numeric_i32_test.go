package wasmvm_test

import (
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
)

func TestCONST_I32(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 6,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.Memory[0] = 0x41 // Opcode
	vm.Memory[1] = 0x78
	vm.Memory[2] = 0x56
	vm.Memory[3] = 0x34
	vm.Memory[4] = 0x12
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, success := vm.ValueStack.Pop()
	assert.True(t, success)
	assert.Equal(t, uint64(5), vm.PC)
	assert.Equal(t, uint32(0x12345678), val.Value_I32)

}

func TestCONST_I32_OOB(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 4,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)

	vm.Memory[0] = 0x41 // Opcode
	vm.Memory[1] = 0x78
	vm.Memory[2] = 0x56
	vm.PC = 0
	err = vm.Step()
	assert.Error(t, err)
	assert.Equal(t, int(0), vm.ValueStack.Size())
	assert.True(t, vm.Trap)
	assert.Equal(t, "CONST_I32: Out of bounds", vm.TrapReason)

}

func TestADD_I32_NotEnoughStack(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	// opcode
	vm.Memory[0] = 0x6A
	vm.PC = 0
	err = vm.Step()
	assert.Error(t, err)
	assert.Equal(t, int(0), vm.ValueStack.Size())
	assert.True(t, vm.Trap)
	assert.Equal(t, "ADD_I32: Stack Underflow", vm.TrapReason)
}

func TestADD_I32_SmallNumbers(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)

	vm.ValueStack.PushInt32(5)
	vm.ValueStack.PushInt32(7)
	// opcode
	vm.Memory[0] = 0x6A
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, success := vm.ValueStack.Pop()
	assert.True(t, success)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, uint32(12), val.Value_I32)
	assert.Equal(t, int(0), vm.ValueStack.Size())
}

func TestADD_I32_OverflowWrap(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt32(0xFFFFFFFF)
	vm.ValueStack.PushInt32(2)
	// opcode
	vm.Memory[0] = 0x6A
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, success := vm.ValueStack.Pop()
	assert.True(t, success)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, uint32(1), val.Value_I32)
	assert.Equal(t, int(0), vm.ValueStack.Size())
}

func TestSUB_I32_NotEnoughStack(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	// opcode
	vm.Memory[0] = 0x6B
	vm.PC = 0
	err = vm.Step()
	assert.Error(t, err)
	assert.Equal(t, int(0), vm.ValueStack.Size())
	assert.True(t, vm.Trap)
	assert.Equal(t, "SUB_I32: Stack Underflow", vm.TrapReason)
}

func TestSUB_I32_SmallNumbers(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt32(7)
	vm.ValueStack.PushInt32(5)
	// opcode
	vm.Memory[0] = 0x6B
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, success := vm.ValueStack.Pop()
	assert.True(t, success)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, uint32(2), val.Value_I32)
	assert.Equal(t, int(0), vm.ValueStack.Size())
}

func TestSUB_I32_OverflowWrap(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt32(1)
	vm.ValueStack.PushInt32(2)
	// opcode
	vm.Memory[0] = 0x6B
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, success := vm.ValueStack.Pop()
	assert.True(t, success)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, uint32(0xFFFFFFFF), val.Value_I32)
	assert.Equal(t, int(0), vm.ValueStack.Size())
}

func TestMUL_I32_NotEnoughStack(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.Memory[0] = 0x6C // MUL_I32
	vm.PC = 0
	err = vm.Step()
	assert.Error(t, err)
	assert.Equal(t, 0, vm.ValueStack.Size())
	assert.True(t, vm.Trap)
	assert.Equal(t, "MUL_I32: Stack Underflow", vm.TrapReason)
}

func TestMUL_I32_MultiplyByZero(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt32(12345)
	vm.ValueStack.PushInt32(0)
	vm.Memory[0] = 0x6C
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, ok := vm.ValueStack.Pop()
	assert.True(t, ok)
	assert.Equal(t, uint32(0), val.Value_I32)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, 0, vm.ValueStack.Size())
}

func TestMUL_I32_MultiplyByOne(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt32(1)
	vm.ValueStack.PushInt32(0x12345678)
	vm.Memory[0] = 0x6C
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, ok := vm.ValueStack.Pop()
	assert.True(t, ok)
	assert.Equal(t, uint32(0x12345678), val.Value_I32)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, 0, vm.ValueStack.Size())
}

func TestMUL_I32_PositiveTimesNegative(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt32(7)
	vm.ValueStack.PushInt32(0xFFFFFFFD) // DWORD decimal -3
	vm.Memory[0] = 0x6C
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, ok := vm.ValueStack.Pop()
	assert.True(t, ok)
	assert.Equal(t, uint32(0xFFFFFFEB), val.Value_I32) // DWORD decimal -21
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, 0, vm.ValueStack.Size())
}

func TestMUL_I32_NegativeTimesNegative(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt32(0xFFFFFFFE) // DWORD decimal -2
	vm.ValueStack.PushInt32(0xFFFFFFFC) // DWORD decimal -4
	vm.Memory[0] = 0x6C
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, ok := vm.ValueStack.Pop()
	assert.True(t, ok)
	assert.Equal(t, uint32(8), val.Value_I32)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, 0, vm.ValueStack.Size())
}

func TestMUL_I32_OverflowWrap(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt32(0xFFFFFFFF) // DWORD decimal -1
	vm.ValueStack.PushInt32(2)
	vm.Memory[0] = 0x6C
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, ok := vm.ValueStack.Pop()
	assert.True(t, ok)
	// 0xFFFFFFFF * 2 = 0xFFFFFFFE (wraps as unsigned), which as int32 is -2
	assert.Equal(t, ^uint32(0)-1, val.Value_I32)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, 0, vm.ValueStack.Size())
}

func TestDIV_I32_DivideByZero(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt32(1) // DWORD decimal -1
	vm.ValueStack.PushInt32(0)
	vm.Memory[0] = 0x6E
	vm.PC = 0
	err = vm.Step()
	assert.Error(t, err)
	assert.Equal(t, 0, vm.ValueStack.Size())
	assert.True(t, vm.Trap)
	assert.Equal(t, "DIVU_I32: Divide by Zero", vm.TrapReason)
}
