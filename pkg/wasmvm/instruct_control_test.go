package wasmvm_test

import (
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
)

// Tests the NOP instruction.
// Program counter should increment, and no trap
func TestNOP(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.Memory[0] = 0x01
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), vm.PC)
	assert.False(t, vm.Trap)
}

// Tests the END instruction
// Program counter should increment, and there should be a trap
func TestEND(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.Memory[0] = 0x0B
	vm.PC = 0
	err = vm.Step()
	assert.NoError(t, err)
	assert.True(t, vm.Trap)
}
