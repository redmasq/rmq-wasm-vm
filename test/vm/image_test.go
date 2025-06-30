package vm_test

import (
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPopulateImage_FileType(t *testing.T) {
	// Mock ReadFile
	original := wasmvm.ReadFile
	defer func() { wasmvm.ReadFile = original }()
	wasmvm.ReadFile = func(string) ([]byte, error) {
		return []byte{0xAB, 0xCD}, nil
	}

	mem := make([]byte, 4)
	cfg := &wasmvm.ImageConfig{
		Type:     "file",
		Filename: "fake.file",
	}
	warns, err := wasmvm.PopulateImage(mem, cfg, false)
	assert.NoError(t, err)
	assert.Empty(t, warns)
	assert.Equal(t, []byte{0xAB, 0xCD, 0, 0}, mem)
}

func TestPopulateImage_ArrayType(t *testing.T) {
	mem := make([]byte, 4)
	cfg := &wasmvm.ImageConfig{
		Type:  "array",
		Size:  4,
		Array: []uint8{1, 2},
	}
	warns, err := wasmvm.PopulateImage(mem, cfg, false)
	assert.NoError(t, err)
	assert.Empty(t, warns)
	assert.Equal(t, []byte{1, 2, 0, 0}, mem)
}

func TestPopulateImage_ArrayType_OOB(t *testing.T) {
	mem := make([]byte, 4)
	cfg := &wasmvm.ImageConfig{
		Type:  "array",
		Size:  1,
		Array: []uint8{1, 2},
	}
	_, err := wasmvm.PopulateImage(mem, cfg, true)
	assert.Error(t, err)

	warns, err := wasmvm.PopulateImage(mem, cfg, false)
	assert.NoError(t, err)
	assert.Contains(t, warns[0], "array entry larger than size")
}

func TestPopulateImage_EmptyType(t *testing.T) {
	mem := []byte{99, 88, 77}
	cfg := &wasmvm.ImageConfig{
		Type: "empty",
		Size: 3,
	}
	warns, err := wasmvm.PopulateImage(mem, cfg, false)
	assert.NoError(t, err)
	assert.Empty(t, warns)
	assert.Equal(t, []byte{0, 0, 0}, mem)
}

func TestPopulateImage_SparseArrayType_Normal(t *testing.T) {
	mem := make([]byte, 10) // memory with 10 bytes, all zero by default
	cfg := &wasmvm.ImageConfig{
		Type: "sparsearray",
		Sparse: []wasmvm.SparseArrayEntry{
			{Offset: 0, Array: []uint8{1, 2, 3}}, // fills mem[0], mem[1], mem[2]
			{Offset: 7, Array: []uint8{8, 9}},    // fills mem[7], mem[8]
		},
	}
	warns, err := wasmvm.PopulateImage(mem, cfg, true)
	assert.NoError(t, err)
	assert.Empty(t, warns)
	assert.Equal(t, []byte{1, 2, 3, 0, 0, 0, 0, 8, 9, 0}, mem)
}

func TestPopulateImage_SparseArrayType_StrictAndLenient_OOB(t *testing.T) {
	// Test strict
	mem := make([]byte, 2)
	cfg := &wasmvm.ImageConfig{
		Type: "sparsearray",
		Sparse: []wasmvm.SparseArrayEntry{
			{Offset: 0, Array: []uint8{7}},
			{Offset: 2, Array: []uint8{8}}, // out of bounds
		},
	}
	warns, err := wasmvm.PopulateImage(mem, cfg, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sparsearray entry out of bounds")

	// Test lenient
	mem = make([]byte, 2)
	warns, err = wasmvm.PopulateImage(mem, cfg, false)
	assert.NoError(t, err)
	assert.Contains(t, warns[0], "sparsearray entry out of bounds")
	assert.Equal(t, uint8(7), mem[0])
}

func TestPopulateImage_OverwriteDetection(t *testing.T) {
	// Overwrite warning, lenient
	mem := []byte{5, 0}
	cfg := &wasmvm.ImageConfig{
		Type: "sparsearray",
		Sparse: []wasmvm.SparseArrayEntry{
			{Offset: 0, Array: []uint8{6}},
		},
	}
	warns, err := wasmvm.PopulateImage(mem, cfg, false)
	assert.NoError(t, err)
	assert.Contains(t, warns[0], "overwrite at offset 0")

	// Overwrite error, strict
	mem = []byte{5, 0}
	warns, err = wasmvm.PopulateImage(mem, cfg, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "overwrite at offset 0")
}

func TestPopulateImage_UnknownType(t *testing.T) {
	mem := make([]byte, 1)
	cfg := &wasmvm.ImageConfig{
		Type: "foobar",
	}
	warns, err := wasmvm.PopulateImage(mem, cfg, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown image type")
	assert.Empty(t, warns)
}

func TestParseImageConfig_JSON(t *testing.T) {
	raw := []byte(`{"type":"array", "array":[1,2,3], "size":4}`)
	cfg, err := wasmvm.ParseImageConfig(raw)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "array", cfg.Type)
	assert.Equal(t, []uint8{1, 2, 3}, cfg.Array)
	assert.Equal(t, uint64(4), cfg.Size)
}
