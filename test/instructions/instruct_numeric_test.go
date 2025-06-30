package instructions

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
	vm.Memory[0] = 0x43
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

func TestADD_I32_SmallNumbers(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	// opcode
	vm.ValueStack.PushInt32(5)
	vm.ValueStack.PushInt32(7)
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
