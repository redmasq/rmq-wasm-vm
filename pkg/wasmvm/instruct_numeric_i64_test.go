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
	vm.Memory[0] = wasmvm.OP_CONST_I64 // Opcode
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

	vm.Memory[0] = wasmvm.OP_CONST_I64 // Opcode
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
	vm.Memory[0] = wasmvm.OP_ADD_I64
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
	vm.Memory[0] = wasmvm.OP_ADD_I64
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
	vm.Memory[0] = wasmvm.OP_ADD_I64
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
	vm.Memory[0] = wasmvm.OP_SUB_I64
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
	vm.Memory[0] = wasmvm.OP_SUB_I64
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
	vm.Memory[0] = wasmvm.OP_SUB_I64
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, success := vm.ValueStack.Pop()
	assert.True(t, success)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, ^uint64(0), val.Value_I64) // Short hand for all f;s
	assert.Equal(t, int(0), vm.ValueStack.Size())
}

func TestMUL_I64_NotEnoughStack(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.Memory[0] = wasmvm.OP_MUL_I64 // MUL_I64
	vm.PC = 0
	err = vm.Step()
	assert.Error(t, err)
	assert.Equal(t, 0, vm.ValueStack.Size())
	assert.True(t, vm.Trap)
	assert.Equal(t, "MUL_I64: Stack Underflow", vm.TrapReason)
}

func TestMUL_I64_MultiplyByZero(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt64(12345)
	vm.ValueStack.PushInt64(0)
	vm.Memory[0] = wasmvm.OP_MUL_I64
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, ok := vm.ValueStack.Pop()
	assert.True(t, ok)
	assert.Equal(t, uint64(0), val.Value_I64)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, 0, vm.ValueStack.Size())
}

func TestMUL_I64_MultiplyByOne(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt64(1)
	vm.ValueStack.PushInt64(0x0123456789ABCDEF)
	vm.Memory[0] = wasmvm.OP_MUL_I64
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, ok := vm.ValueStack.Pop()
	assert.True(t, ok)
	assert.Equal(t, uint64(0x0123456789ABCDEF), val.Value_I64)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, 0, vm.ValueStack.Size())
}

func TestMUL_I64_PositiveTimesNegative(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt64(7)
	vm.ValueStack.PushInt64(0xFFFFFFFFFFFFFFFD) // QWORD decimal -3
	vm.Memory[0] = wasmvm.OP_MUL_I64
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, ok := vm.ValueStack.Pop()
	assert.True(t, ok)
	assert.Equal(t, uint64(0xFFFFFFFFFFFFFFEB), val.Value_I64) // QWORD decimal -21
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, 0, vm.ValueStack.Size())
}

func TestMUL_I64_NegativeTimesNegative(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt64(0xFFFFFFFFFFFFFFFE) // QWORD decimal -2
	vm.ValueStack.PushInt64(0xFFFFFFFFFFFFFFFC) // QWORD decimal -4
	vm.Memory[0] = wasmvm.OP_MUL_I64
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, ok := vm.ValueStack.Pop()
	assert.True(t, ok)
	assert.Equal(t, uint64(8), val.Value_I64)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, 0, vm.ValueStack.Size())
}

func TestMUL_I64_OverflowWrap(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt64(0xFFFFFFFFFFFFFFFF) // QWORD decimal -1
	vm.ValueStack.PushInt64(2)
	vm.Memory[0] = wasmvm.OP_MUL_I64
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	val, ok := vm.ValueStack.Pop()
	assert.True(t, ok)
	// 0xFFFFFFFFFFFFFFFF * 2 = 0xFFFFFFFFFFFFFFFE (wraps as unsigned), which as int64 is -2
	assert.Equal(t, ^uint64(0)-1, val.Value_I64)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, 0, vm.ValueStack.Size())
}

func TestDIVU_I64_DivideByZero(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt64(1)
	vm.ValueStack.PushInt64(0)
	vm.Memory[0] = wasmvm.OP_DIVU_I64
	vm.PC = 0
	err = vm.Step()
	assert.Error(t, err)
	assert.Equal(t, 0, vm.ValueStack.Size())
	assert.True(t, vm.Trap)
	assert.Equal(t, "DIVU_I64: Divide by Zero", vm.TrapReason)
}

func TestDIVU_I64_StackUnderflow(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt64(1)
	vm.Memory[0] = wasmvm.OP_DIVU_I64
	vm.PC = 0
	err = vm.Step()
	assert.Error(t, err)
	assert.Equal(t, 1, vm.ValueStack.Size())
	assert.True(t, vm.Trap)
	assert.Equal(t, "DIVU_I64: Stack Underflow", vm.TrapReason)
}

func TestDIVU_I64_SmallValues(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt64(10)
	vm.ValueStack.PushInt64(5)
	vm.Memory[0] = wasmvm.OP_DIVU_I64
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	assert.Equal(t, 1, vm.ValueStack.Size())
	val, ok := vm.ValueStack.Pop()
	assert.True(t, ok)
	assert.Equal(t, uint64(2), val.Value_I64)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, 0, vm.ValueStack.Size())
}

func TestDIVU_I64_DivideByOne(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt64(42)
	vm.ValueStack.PushInt64(1)
	vm.Memory[0] = wasmvm.OP_DIVU_I64
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	assert.Equal(t, 1, vm.ValueStack.Size())
	val, ok := vm.ValueStack.Pop()
	assert.True(t, ok)
	assert.Equal(t, uint64(42), val.Value_I64)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, 0, vm.ValueStack.Size())
}

func TestDIVU_I64_DivisorLargerThanDividend(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt64(42)
	vm.ValueStack.PushInt64(137)
	vm.Memory[0] = wasmvm.OP_DIVU_I64
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	assert.Equal(t, 1, vm.ValueStack.Size())
	val, ok := vm.ValueStack.Pop()
	assert.True(t, ok)
	assert.Equal(t, uint64(0), val.Value_I64)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, 0, vm.ValueStack.Size())
}

func TestDIVU_I64_ZeroDividend(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt64(0)
	vm.ValueStack.PushInt64(137)
	vm.Memory[0] = wasmvm.OP_DIVU_I64
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	assert.Equal(t, 1, vm.ValueStack.Size())
	val, ok := vm.ValueStack.Pop()
	assert.True(t, ok)
	assert.Equal(t, uint64(0), val.Value_I64)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, 0, vm.ValueStack.Size())
}

func TestDIVU_I64_MaxValueBySelf(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt64(^uint64(0))
	vm.ValueStack.PushInt64(^uint64(0))
	vm.Memory[0] = wasmvm.OP_DIVU_I64
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	assert.Equal(t, 1, vm.ValueStack.Size())
	val, ok := vm.ValueStack.Pop()
	assert.True(t, ok)
	assert.Equal(t, uint64(1), val.Value_I64)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, 0, vm.ValueStack.Size())
}

func TestDIVU_I64_MaxValueByOnef(t *testing.T) {
	cfg := &wasmvm.VMConfig{Size: 1}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt64(^uint64(0))
	vm.ValueStack.PushInt64(1)
	vm.Memory[0] = wasmvm.OP_DIVU_I64
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	assert.Equal(t, 1, vm.ValueStack.Size())
	val, ok := vm.ValueStack.Pop()
	assert.True(t, ok)
	assert.Equal(t, ^uint64(0), val.Value_I64)
	assert.Equal(t, uint64(1), vm.PC)
	assert.Equal(t, 0, vm.ValueStack.Size())
}
