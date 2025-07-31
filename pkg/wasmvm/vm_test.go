package wasmvm_test

import (
	"bytes"
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestNewVM_MissingSizeAndMemory(t *testing.T) {
	cfg := &wasmvm.VMConfig{}
	vm, err := wasmvm.NewVM(cfg)
	assert.Nil(t, vm)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "either Size or FlatMemory must be specified")
}

func TestNewVM_FlatMemory(t *testing.T) {
	mem := []byte{1, 2, 3}
	cfg := &wasmvm.VMConfig{
		FlatMemory: mem,
	}
	vm, err := wasmvm.NewVM(cfg)
	require.NoError(t, err)
	require.NotNil(t, vm)
	assert.Equal(t, mem, vm.Memory)
}

func TestNewVM_RingZeroStrict(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size:   1,
		Strict: true,
		Rings: map[uint8]wasmvm.RingConfig{
			0: {Enabled: true},
		},
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.Nil(t, vm)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ring 0 cannot be reconfigured")
}

func TestNewVM_RingZeroNonStrictWarns(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 1,
		Rings: map[uint8]wasmvm.RingConfig{
			0: {Enabled: true},
		},
	}
	vm, err := wasmvm.NewVM(cfg)
	require.NoError(t, err)
	assert.NotNil(t, vm)
	assert.Contains(t, vm.Config.Rings, uint8(0))
}

func TestNewVM_StartOverride(t *testing.T) {
	start := uint64(5)
	cfg := &wasmvm.VMConfig{
		Size:          10,
		StartOverride: &start,
	}
	vm, err := wasmvm.NewVM(cfg)
	require.NoError(t, err)
	assert.Equal(t, start, vm.PC)
}

func TestNewVM_ImageEmptyType(t *testing.T) {
	img := &wasmvm.ImageConfig{
		Type: wasmvm.Empty,
		Size: 3,
	}
	cfg := &wasmvm.VMConfig{
		Size:  3,
		Image: img,
	}
	vm, err := wasmvm.NewVM(cfg)
	require.NoError(t, err)
	assert.NotNil(t, vm)
	assert.Equal(t, []byte{0, 0, 0}, vm.Memory)
}

func TestNewVM_ImageSparseArrayStrict(t *testing.T) {
	img := &wasmvm.ImageConfig{
		Type: wasmvm.SparseArray,
		Size: 4,
		Sparse: []wasmvm.SparseArrayEntry{
			{Offset: 0, Array: []byte{1, 2}},
			{Offset: 3, Array: []byte{9, 9}}, // 4th byte out of bounds
		},
	}
	cfg := &wasmvm.VMConfig{
		Size:   4,
		Image:  img,
		Strict: true,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.Nil(t, vm)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sparsearray entry out of bounds")
}

func TestNewVM_ImageSparseArrayLenient(t *testing.T) {
	img := &wasmvm.ImageConfig{
		Type: wasmvm.SparseArray,
		Size: 4,
		Sparse: []wasmvm.SparseArrayEntry{
			{Offset: 0, Array: []byte{1, 2}},
			{Offset: 3, Array: []byte{9, 9}}, // 4th byte out of bounds
		},
	}
	cfg := &wasmvm.VMConfig{
		Size:  4,
		Image: img,
		// Strict is false
	}
	vm, err := wasmvm.NewVM(cfg)
	require.NoError(t, err)
	assert.NotNil(t, vm)
	assert.Contains(t, vm.ImageInitWarn, "sparsearray entry out of bounds at offset 4")
}

func TestNewVM_ImageUnknownType(t *testing.T) {
	img := &wasmvm.ImageConfig{
		Type: wasmvm.ImageType(int(wasmvm.SparseArray) + 1),
		Size: 4,
	}
	cfg := &wasmvm.VMConfig{
		Size:   4,
		Image:  img,
		Strict: true,
	}
	vm, err := wasmvm.NewVM(cfg)
	assert.Nil(t, vm)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown image type")
}

func TestVMState_ExecuteNext_Trap(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 2,
	}
	vm, err := wasmvm.NewVM(cfg)
	require.NoError(t, err)
	vm.Trap = true
	vm.TrapReason = "Simulated trap"
	err = vm.Step()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Simulated trap")
}

func TestVMState_ExecuteNext_UnknownOpcode(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 2,
	}
	vm, err := wasmvm.NewVM(cfg)
	require.NoError(t, err)
	vm.Memory[0] = 0xFF // No handler for 0xFF in default map
	vm.PC = 0
	err = vm.Step()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unknown instruction")
}

func TestVMState_ExecuteNext_ProgramCounterOutOfBounds(t *testing.T) {
	cfg := &wasmvm.VMConfig{
		Size: 2,
	}
	vm, err := wasmvm.NewVM(cfg)
	require.NoError(t, err)
	vm.PC = 5
	err = vm.Step()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Program counter out of bounds")
}

func TestVMState_MainLoop_ErrorOutput(t *testing.T) {
	var buf bytes.Buffer
	cfg := &wasmvm.VMConfig{
		Size:   2,
		Stdout: &buf,
		Stderr: &buf,
	}
	vm, err := wasmvm.NewVM(cfg)
	require.NoError(t, err)
	vm.Memory[0] = 0xFF // Will cause unknown instruction
	vm.MainLoop()
	assert.Contains(t, buf.String(), "Execution error:")
}
