package wasmvm_test

import (
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
)

// For the i32 test cases
type controlTestCase struct {
	name          string   // Test case description
	memoryContent []byte   // Initial memory content
	expectTrap    bool     // Expect a trap error
	trapReason    string   // Expected reason for trap, if any
	expectValue   []uint32 // Expected value pushed on the stack
	expectPC      uint64   // Expected program counter after execution
	stackValues   []uint32
	expectedStack int // Stack size after execution, before popping result
}

// runTestBatchControl runs a suite of i32TestCase VM table tests and asserts expected VM and stack outcomes.
func runTestBatchControl(t *testing.T, tests []controlTestCase) {
	for i := range tests {
		tc := tests[i]
		name := tc.name
		memorySize := uint64(len(tc.memoryContent))
		t.Run(name, func(t *testing.T) {
			// Initialize VM configuration
			cfg := &wasmvm.VMConfig{
				Size: memorySize,
				Image: &wasmvm.ImageConfig{
					Type:  wasmvm.Array,
					Array: tc.memoryContent,
					Size:  memorySize,
				},
			}
			vm, err := wasmvm.NewVM(cfg)
			assert.NoError(t, err)

			// Push test values onto stack
			for _, val := range tc.stackValues {
				vm.ValueStack.PushInt32(val)
			}

			vm.PC = 0

			err = vm.Step()

			if tc.expectTrap {
				assert.Error(t, err)
				assert.True(t, vm.Trap)
				assert.Equal(t, tc.trapReason, vm.TrapReason)
				assert.Equal(t, tc.expectedStack, vm.ValueStack.Size())
			} else {
				assert.NoError(t, err)
				assert.False(t, vm.Trap, "Trap was raised: "+vm.TrapReason)
				if tc.expectedStack > 0 {
					assert.Equal(t, tc.expectedStack, vm.ValueStack.Size())
					for i := range tc.expectValue {
						v := tc.expectValue[len(tc.expectValue)-i-1]
						val, success := vm.ValueStack.Pop()
						assert.True(t, success)
						assert.Equal(t, v, val.Value_I32)
					}
				}
				assert.Equal(t, tc.expectPC, vm.PC)
			}
		})
	}
}

// Tests NOP and END
// This might need to be split later when more instructions are added
func TestControl(t *testing.T) {
	tests := []controlTestCase{
		{
			// Program counter should increment, and no trap
			name:        "NOP",
			stackValues: []uint32{},
			memoryContent: []byte{
				wasmvm.OP_NOP,
			},
			expectTrap:    false,
			expectValue:   []uint32{},
			expectPC:      1,
			expectedStack: 0,
		},
		{
			// Program counter should increment, and there should be a trap
			name:        "END",
			stackValues: []uint32{},
			memoryContent: []byte{
				wasmvm.OP_END,
			},
			expectTrap:    true,
			trapReason:    "END: Call Stack Empty",
			expectPC:      1,
			expectedStack: 0,
		},
	}
	runTestBatchControl(t, tests)
}
