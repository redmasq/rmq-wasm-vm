package wasmvm_test

import (
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
)

func TestCloneEmpty(t *testing.T) {
	var vmc *wasmvm.VMConfig
	testvmc, err := vmc.QuickClone()
	assert.Nil(t, testvmc)
	assert.NoError(t, err)
}

func TestErrStr(t *testing.T) {
	errStr := wasmvm.VmInitErrStr(wasmvm.VMInitializationErrorType(byte(wasmvm.VMRingAlreadyExists) + 1))
	assert.Contains(t, errStr, "unknown vm initialization error")
}
