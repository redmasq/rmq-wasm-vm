package instructions

import (
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
)

func TestADD8(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 8,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	// opcode, a, b, dest
	vm.Memory[0] = 0x01 // ADD8
	vm.Memory[1] = 5    // a
	vm.Memory[2] = 7    // b
	vm.Memory[3] = 0    // initial dest
	vm.PC = 0
	err = vm.ExecuteNext()
	assert.NoError(t, err)
	assert.Equal(t, uint8(12), vm.Memory[3])
	assert.Equal(t, uint64(4), vm.PC)
	assert.False(t, vm.Trap)
}
