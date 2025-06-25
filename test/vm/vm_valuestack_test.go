package vm_test

import (
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
)

func TestPushPopInt32(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	vm.ValueStack.PushInt32(uint32(0xDEADBEEF))
	val, success := vm.ValueStack.Pop()
	assert.True(t, success)
	assert.Equal(t, wasmvm.TYPE_I32, val.EntryType)
	assert.Equal(t, uint32(0xDEADBEEF), val.Value_I32)
}
