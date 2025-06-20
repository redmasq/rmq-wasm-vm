package instructions

import (
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
)

func TestNOP(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.Memory[0] = 0x01
	vm.PC = 0
	err = vm.ExecuteNext()
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), vm.PC)
	assert.False(t, vm.Trap)
}

func TestEND(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.Memory[0] = 0x0B
	vm.PC = 0
	err = vm.ExecuteNext()
	assert.NoError(t, err)
	assert.True(t, vm.Trap)
}
