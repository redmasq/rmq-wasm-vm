package vm_test

import (
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
)

func TestNewVMBasic(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	assert.Equal(t, len(vm.Memory), 1)
}

func TestNewVMFileImage(t *testing.T) {
	original := wasmvm.ReadFile
	defer func() { wasmvm.ReadFile = original }()

	wasmvm.ReadFile = func(path string) ([]byte, error) {
		return []byte{0x42}, nil
	}

	image, err := wasmvm.ParseImageConfig(
		[]byte(`{
	"Type":"file",
	"Filename":"testImage.json"
}`))

	assert.NoError(t, err)

	cfg := &wasmvm.VMConfig{
		Size:  2,
		Image: image,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(vm.Memory))
	assert.Equal(t, uint8(0x42), vm.Memory[0])
}
