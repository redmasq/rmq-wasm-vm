package wasmvm_test

import (
	"bytes"
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type vmTestCase struct {
	name                        string
	config                      *wasmvm.VMConfig
	expect                      *wasmvm.VMState
	checkError                  bool
	checkExpect                 bool
	checkMemorySize             bool
	checkMemoryContent          bool
	expectError                 *wasmvm.VMInitializationError
	expectSize                  uint64
	replaceReadFile             func(string) ([]byte, error)
	prepareImage                func(t *testing.T) *wasmvm.ImageConfig
	expectMemoryContent         []byte
	clearInstructionMapOnActual bool
}

func executeVMTests(t *testing.T, tests []vmTestCase) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			func() {
				if test.replaceReadFile != nil {
					original := wasmvm.ReadFile
					defer func() { wasmvm.ReadFile = original }()
					wasmvm.ReadFile = test.replaceReadFile
				}
				if test.prepareImage != nil {
					test.config.Image = test.prepareImage(t)
				}
				vm, err := wasmvm.NewVM(test.config)
				if test.checkError {
					assert.Error(t, err)
					assert.IsType(t, test.expectError, err)
					assert.Equal(t, test.expectError, err)
				} else {
					assert.NoError(t, err)
				}
				if test.clearInstructionMapOnActual {
					vm.InstructionMap = nil
				}
				if test.checkMemorySize {
					assert.Equal(t, test.expectSize, uint64(len(vm.Memory)))
				}
				if test.checkExpect {
					assert.Equal(t, test.expect, vm)
				}
			}()

		})
	}

}

func TestNewVM(t *testing.T) {
	tests := []vmTestCase{
		{
			name: "success - NewVM Basic",
			config: &wasmvm.VMConfig{
				Size: 1,
			},
			checkMemorySize: true,
			expectSize:      1,
		},
		{
			name: "success - NewVM File",
			config: &wasmvm.VMConfig{
				Size:  2,
				Image: (&wasmvm.ImageConfig{}).SetFilename("testImage.file"),
			},
			checkMemorySize:     true,
			expectSize:          2,
			checkMemoryContent:  true,
			expectMemoryContent: []byte{0x42},
			replaceReadFile: func(path string) ([]byte, error) {
				return []byte{0x42}, nil
			},
		},
		{
			name:       "failure - NewVM zero nil config",
			checkError: true,
			expectError: &wasmvm.VMInitializationError{
				Type: wasmvm.VMConfigRequired,
				Msg:  "config is required",
			},
		},
		{
			name:       "failure - NewVM zero size and no image",
			config:     &wasmvm.VMConfig{},
			checkError: true,
			expectError: &wasmvm.VMInitializationError{
				Type: wasmvm.MissingSizeOrFlatMemory,
				Msg:  "either Size or FlatMemory must be specified",
			},
		},
		{
			name: "success - NewVM FlatMemory",
			config: &wasmvm.VMConfig{
				FlatMemory: []byte{1, 2, 3},
			},
			checkMemorySize:     true,
			expectSize:          3,
			checkMemoryContent:  true,
			expectMemoryContent: []byte{1, 2, 3},
		},
		{
			name: "success - ring 0 strict",
			config: &wasmvm.VMConfig{
				Size:   1,
				Strict: true,
				Rings: map[uint8]wasmvm.RingConfig{
					1: {Enabled: true},
				},
			},
			expect: &wasmvm.VMState{
				Memory: []byte{0x00},
				Config: &wasmvm.VMConfig{
					Size:   1,
					Strict: true,
					Rings: map[uint8]wasmvm.RingConfig{
						0: {Enabled: true},
						1: {Enabled: true},
					},
				},
			},
		},
		{
			name: "warn - ring 0 non-strict",
			config: &wasmvm.VMConfig{
				Size:   42,
				Strict: false,
				Rings: map[uint8]wasmvm.RingConfig{
					0: {Enabled: true},
				},
			},
			checkExpect:                 true,
			clearInstructionMapOnActual: true,
			expect: &wasmvm.VMState{
				Memory: make([]byte, 42),
				Config: &wasmvm.VMConfig{
					Size:   42,
					Strict: false,
					Rings: map[uint8]wasmvm.RingConfig{
						0: {Enabled: true},
					},
				},
				ImageInitWarn: []string{
					"Ring 0 redefinition ignored",
				},
			},
		},
		{
			name: "failure - ring 0 strict",
			config: &wasmvm.VMConfig{
				Size:   1,
				Strict: true,
				Rings: map[uint8]wasmvm.RingConfig{
					0: {Enabled: true},
				},
			},
			checkError: true,
			expectError: &wasmvm.VMInitializationError{
				Type: wasmvm.StrictModeAttemptRing0Reconfigure,
				Msg:  "ring 0 cannot be reconfigured (strict mode)",
			},
		},
		{
			name: "success - Start Override",
			config: &wasmvm.VMConfig{
				FlatMemory:    make([]byte, 10),
				StartOverride: uint64(5),
			},
			checkExpect:                 true,
			clearInstructionMapOnActual: true,
			expect: &wasmvm.VMState{
				Memory: make([]byte, 10),
				Config: &wasmvm.VMConfig{
					FlatMemory: make([]byte, 10),
					Rings: map[uint8]wasmvm.RingConfig{
						0: {Enabled: true},
					},
					StartOverride: uint64(5),
				},
				PC: uint64(5),
			},
		},
		{
			name: "success - NewVM Type Empty",
			config: &wasmvm.VMConfig{
				Size:  3,
				Image: (&wasmvm.ImageConfig{}).SetType(wasmvm.Empty).SetSize(3),
			},
			checkMemorySize:     true,
			expectSize:          3,
			checkMemoryContent:  true,
			expectMemoryContent: []byte{0, 0, 0},
		},
		{
			name: "failure - NewVM Type SparseArray",
			config: &wasmvm.VMConfig{
				Size:   3,
				Strict: true,
				Image: (&wasmvm.ImageConfig{}).SetType(wasmvm.SparseArray).SetSparseArray(
					[]wasmvm.SparseArrayEntry{
						{Offset: 0, Array: []byte{1, 2}},
						{Offset: 3, Array: []byte{9, 9}}, // 4th byte out of bounds
					},
				),
			},
			checkError: true,
			expectError: &wasmvm.VMInitializationError{
				Type: wasmvm.VMImageError,
				Msg:  "an error occurred during Image initialization: [SparseEntryOutOfBounds] sparsearray entry out of bounds detected",
				Cause: &wasmvm.ImageInitializationError{
					Type: wasmvm.SparseEntryOutOfBounds,
					Msg:  "sparsearray entry out of bounds detected",
					Meta: wasmvm.ImageErrorSparseMetaData{
						ConfigSize: uint64(0),
						MemSize:    uint64(3),
						ProblemEntries: []wasmvm.SparseArrayErrorEntry{
							{Offset: 3, Array: []uint8{9, 9}, ErrorType: wasmvm.SparseEntryOutOfBounds}, // out of bounds
						},
					},
				},
				Meta: (&wasmvm.ImageConfig{}).SetType(wasmvm.SparseArray).SetSparseArray(
					[]wasmvm.SparseArrayEntry{
						{Offset: 0, Array: []byte{1, 2}},
						{Offset: 3, Array: []byte{9, 9}}, // 4th byte out of bounds
					},
				),
			},
		},
		{
			name: "warn - NewVM Type SparseArray Lenient",
			config: &wasmvm.VMConfig{
				Size:   3,
				Strict: false,
				Image: (&wasmvm.ImageConfig{}).SetType(wasmvm.SparseArray).SetSparseArray(
					[]wasmvm.SparseArrayEntry{
						{Offset: 0, Array: []byte{1, 2}},
						{Offset: 3, Array: []byte{9, 9}}, // 4th byte out of bounds
					},
				),
			},
			checkExpect:                 true,
			clearInstructionMapOnActual: true,
			expect: &wasmvm.VMState{
				Memory: []byte{0x01, 0x02, 0x00},
				Config: &wasmvm.VMConfig{
					Size:   3,
					Strict: false,
					Image: (&wasmvm.ImageConfig{}).SetType(wasmvm.SparseArray).SetSparseArray(
						[]wasmvm.SparseArrayEntry{
							{Offset: 0, Array: []byte{1, 2}},
							{Offset: 3, Array: []byte{9, 9}}, // 4th byte out of bounds
						},
					),
					Rings: map[uint8]wasmvm.RingConfig{
						0: {Enabled: true},
					},
				},
				ImageInitWarn: []string{
					"sparsearray entry out of bounds at offset 3",
					"sparsearray entry out of bounds at offset 4",
				},
			},
		},
		{
			name: "failure - NewVM Type Unknown",
			config: &wasmvm.VMConfig{
				Size:   3,
				Strict: true,
				Image:  (&wasmvm.ImageConfig{}).SetType(wasmvm.Unknown),
			},
			checkError: true,
			expectError: &wasmvm.VMInitializationError{
				Type: wasmvm.VMImageError,
				Msg:  "an error occurred during Image initialization: [UnknownImageType] unknown image type: Unknown",
				Cause: &wasmvm.ImageInitializationError{
					Type: wasmvm.UnknownImageType,
					Msg:  "unknown image type: Unknown",
				},
				Meta: (&wasmvm.ImageConfig{}).SetType(wasmvm.Unknown),
			},
		},
	}
	executeVMTests(t, tests)
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
	vm, err := cfg.BuildVMState()
	require.NoError(t, err)
	vm.Memory[0] = 0xFF // Will cause unknown instruction
	vm.MainLoop()
	assert.Contains(t, buf.String(), "Execution error:")
}
