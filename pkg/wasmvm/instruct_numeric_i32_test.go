package wasmvm_test

import (
	"math"
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
)

// For the i32 test cases
type i32TestCase struct {
	name          string   // Test case description
	memoryContent []byte   // Initial memory content
	expectTrap    bool     // Expect a trap error
	trapReason    string   // Expected reason for trap, if any
	expectValue   []uint32 // Expected value pushed on the stack
	expectPC      uint64   // Expected program counter after execution
	stackValues   []uint32
	expectedStack int // Stack size after execution, before popping result
}

// runTestBatchI32 runs a suite of i32TestCase VM table tests and asserts expected VM and stack outcomes.
func runTestBatchI32(t *testing.T, tests []i32TestCase) {
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

// Tests for const.i32
func TestCONST_I32(t *testing.T) {
	np := "CONST_I32: "
	tests := []i32TestCase{
		{
			// Should accept little endian immediate value and place i32 on the stack
			name:        np + "Happy Path",
			stackValues: []uint32{},
			memoryContent: []byte{
				wasmvm.OP_CONST_I32, 0x78, 0x56, 0x34, 0x12,
			},
			expectTrap:    false,
			expectValue:   []uint32{uint32(0x12345678)},
			expectPC:      5,
			expectedStack: 1,
		},
		{
			// This should detect a trap since there is no enough memory to finish the DWORD
			name:        np + "Out of Bounds",
			stackValues: []uint32{},
			memoryContent: []byte{
				wasmvm.OP_CONST_I32, 0x78, 0x56,
			},
			expectTrap:    true,
			trapReason:    np + "Out of bounds",
			expectPC:      0,
			expectedStack: 0,
		},
	}
	runTestBatchI32(t, tests)
}

// Tests for add.i32
func TestADD_I32(t *testing.T) {
	np := "ADD_I32: "
	tests := []i32TestCase{
		{
			// This should detect a trap since there is a lack of i32 values on the stack
			name:        np + "Stack Underflow",
			stackValues: []uint32{},
			memoryContent: []byte{
				wasmvm.OP_ADD_I32,
			},
			expectTrap:    true,
			trapReason:    np + "Stack Underflow",
			expectPC:      0,
			expectedStack: 0,
		},
		{
			// happy path - Small Numbers
			name:        np + "Small Numbers",
			stackValues: []uint32{uint32(5), uint32(7)},
			memoryContent: []byte{
				wasmvm.OP_ADD_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{uint32(12)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			// happy path - Overflow Wrap
			// Going past 0xffffffff will wrap back to 0, or in this case for +2 to 1
			name:        np + "Overflow Wrap",
			stackValues: []uint32{^uint32(0), uint32(2)}, // ^uint32(0) or 0xffffffff is DWORD for -1
			memoryContent: []byte{
				wasmvm.OP_ADD_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{uint32(1)},
			expectPC:      1,
			expectedStack: 1,
		},
	}
	runTestBatchI32(t, tests)
}

// Tests for sub.i32
func TestSUB_I32(t *testing.T) {
	np := "SUB_I32: "
	tests := []i32TestCase{
		{
			// This should detect a trap since there is a lack of i32 values on the stack
			name:        np + "Stack Underflow",
			stackValues: []uint32{},
			memoryContent: []byte{
				wasmvm.OP_SUB_I32,
			},
			expectTrap:    true,
			trapReason:    np + "Stack Underflow",
			expectPC:      0,
			expectedStack: 0,
		},
		{
			// happy path - Small Numbers
			name:        np + "Small Numbers",
			stackValues: []uint32{uint32(7), uint32(5)},
			memoryContent: []byte{
				wasmvm.OP_SUB_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{uint32(2)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			// happy path - Overflow Wrap
			// Going past 0 will wrap back to 0xffffffff, or in this case for -2 from 1 to 0xffffffff
			name:        np + "Overflow Wrap",
			stackValues: []uint32{uint32(1), uint32(2)},
			memoryContent: []byte{
				wasmvm.OP_SUB_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{^uint32(0)}, // ^uint32(0) or 0xffffffff is DWORD for -1
			expectPC:      1,
			expectedStack: 1,
		},
	}
	runTestBatchI32(t, tests)
}

// Tests for mul.i32
func TestMUL_I32(t *testing.T) {
	np := "MUL_I32: "
	tests := []i32TestCase{
		{
			// This should detect a trap since there is a lack of i32 values on the stack
			name:        np + "Stack Underflow",
			stackValues: []uint32{},
			memoryContent: []byte{
				wasmvm.OP_MUL_I32,
			},
			expectTrap:    true,
			trapReason:    np + "Stack Underflow",
			expectPC:      0,
			expectedStack: 0,
		},
		{
			// happy path - Small Numbers
			name:        np + "Small Numbers",
			stackValues: []uint32{uint32(7), uint32(5)},
			memoryContent: []byte{
				wasmvm.OP_MUL_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{uint32(35)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			// happy path - Overflow Wrap
			// Going past 0xffffffff will wrap back to 0xfffffffe
			name:        np + "Overflow Wrap",
			stackValues: []uint32{^uint32(0), uint32(2)}, // ^uint32(0) or 0xffffffff is DWORD for -1
			memoryContent: []byte{
				wasmvm.OP_MUL_I32,
			},
			expectTrap: false,
			// 0xFFFFFFFF * 2 = 0xFFFFFFFE (wraps as unsigned), which as int32 is -2
			expectValue:   []uint32{^uint32(0) - 1},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			// happy path - Multiply by 0
			// The typical whatever * 0 = 0
			name:        np + "Multiply by 0",
			stackValues: []uint32{uint32(137), uint32(0)},
			memoryContent: []byte{
				wasmvm.OP_MUL_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{uint32(0)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			// happy path - Multiply by 0
			// The typical whatever * 1 = whatever
			name:        np + "Multiply by 1",
			stackValues: []uint32{uint32(0x12345678), uint32(1)},
			memoryContent: []byte{
				wasmvm.OP_MUL_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{uint32(0x12345678)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			// happy path - Positive times Negative
			// The typical 7 * -3 = -21
			name:        np + "Positive times Negative",
			stackValues: []uint32{uint32(7), (^uint32(0) - 2)}, // DWord -1 then -2 = -3 or 0xFFFFFFFD
			memoryContent: []byte{
				wasmvm.OP_MUL_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{(^uint32(0) - 20)}, // DWord -21 or 0xFFFFFFEB
			expectPC:      1,
			expectedStack: 1,
		},
		{
			// happy path - Negative times Negative
			// The typical -2 * -4 = 8
			name:        np + "Positive times Negative",
			stackValues: []uint32{(^uint32(0) - 1), (^uint32(0) - 3)}, // DWord -2 and -4 respectively
			memoryContent: []byte{
				wasmvm.OP_MUL_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{uint32(8)}, // Negative times negative equals positive
			expectPC:      1,
			expectedStack: 1,
		},
	}
	runTestBatchI32(t, tests)
}

// Table tests for div_u.i32
func TestDIVU_I32(t *testing.T) {
	np := "DIVU_I32: "
	tests := []i32TestCase{
		{
			name:        np + "Divide By Zero",
			stackValues: []uint32{1, 0},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I32,
			},
			expectTrap:    true,
			trapReason:    np + "Divide by Zero",
			expectPC:      0,
			expectedStack: 0,
		},
		{
			name:        np + "Small Values - Exact",
			stackValues: []uint32{10, 5},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{2},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Small Values - Non-Exact",
			stackValues: []uint32{11, 5},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{2},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Stack Underflow",
			stackValues: []uint32{uint32(1)},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I32,
			},
			expectTrap:    true,
			trapReason:    np + "Stack Underflow",
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "One by One",
			stackValues: []uint32{1, 1},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{1},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Max value by itself",
			stackValues: []uint32{^uint32(0), ^uint32(0)},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{1},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Large by One",
			stackValues: []uint32{uint32(math.MaxInt32), 1},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{uint32(math.MaxInt32)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Large by Large - Overflows to 1",
			stackValues: []uint32{uint32(math.MaxInt32), uint32(math.MaxInt32)},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{uint32(1)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Zero Dividend",
			stackValues: []uint32{uint32(0), uint32(42)},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{uint32(0)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Dividend Smaller than Divisor",
			stackValues: []uint32{uint32(41), uint32(42)},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{uint32(0)},
			expectPC:      1,
			expectedStack: 1,
		},
	}
	runTestBatchI32(t, tests)
}

// Table tests for div_s.i32
func TestDIVS_I32(t *testing.T) {
	np := "DIVS_I32: "
	tests := []i32TestCase{
		{
			name:        np + "Divide By Zero",
			stackValues: []uint32{1, 0},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I32,
			},
			expectTrap:    true,
			trapReason:    np + "Divide by Zero",
			expectPC:      0,
			expectedStack: 0,
		},
		{
			name:        np + "Small Values - Exact",
			stackValues: []uint32{10, 5},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{2},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Small Values - Non-Exact",
			stackValues: []uint32{11, 5},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{2},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Min Negative by Negative One Division Overflow Trap",
			stackValues: []uint32{uint32(0x80000000), ^uint32(0)},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I32,
			},
			expectTrap:    true,
			trapReason:    np + "Signed Division Overflow",
			expectPC:      0,
			expectedStack: 0,
		},
		{
			name:        np + "Stack Underflow",
			stackValues: []uint32{uint32(1)},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I32,
			},
			expectTrap:    true,
			trapReason:    np + "Stack Underflow",
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Positive One by Positive One",
			stackValues: []uint32{1, 1},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{1},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Positive One by Negative One",
			stackValues: []uint32{1, ^uint32(0)},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{^uint32(0)}, // DWORD for -1
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Negative One by Positive One",
			stackValues: []uint32{^uint32(0), 1},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{^uint32(0)}, // DWORD for -1
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Negative One by Negative One",
			stackValues: []uint32{^uint32(0), ^uint32(0)},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{1},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Max Positive by Positive One",
			stackValues: []uint32{uint32(math.MaxInt32), 1},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{uint32(math.MaxInt32)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Max Positive by Negative One",
			stackValues: []uint32{uint32(math.MaxInt32), ^uint32(0)},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{uint32(0x80000001)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Max Positive by Max Positive",
			stackValues: []uint32{uint32(math.MaxInt32), uint32(math.MaxInt32)},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{uint32(1)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Zero Dividend",
			stackValues: []uint32{uint32(0), uint32(42)},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{uint32(0)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Dividend Smaller than Divisor",
			stackValues: []uint32{uint32(41), uint32(42)},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I32,
			},
			expectTrap:    false,
			expectValue:   []uint32{uint32(0)},
			expectPC:      1,
			expectedStack: 1,
		},
	}
	runTestBatchI32(t, tests)
}
