package wasmvm_test

import (
	"math"
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
)

// For the i64 test cases
type i64TestCase struct {
	name          string   // Test case description
	memoryContent []byte   // Initial memory content
	expectTrap    bool     // Expect a trap error
	trapReason    string   // Expected reason for trap, if any
	expectValue   []uint64 // Expected value pushed on the stack
	expectPC      uint64   // Expected program counter after execution
	stackValues   []uint64
	expectedStack int // Stack size after execution, before popping result
}

// runTestBatchI64 runs a suite of i64TestCase VM table tests and asserts expected VM and stack outcomes.
func runTestBatchI64(t *testing.T, tests []i64TestCase) {
	for i := range tests {
		tc := tests[i]
		name := tc.name
		memorySize := uint64(len(tc.memoryContent))
		t.Run(name, func(t *testing.T) {
			// Initialize VM configuration
			cfg := &wasmvm.VMConfig{
				Size: memorySize,
				Image: &wasmvm.ImageConfig{
					Type:  "array",
					Array: tc.memoryContent,
					Size:  memorySize,
				},
			}
			vm, err := wasmvm.NewVM(cfg)
			assert.NoError(t, err)

			// Push test values onto stack
			for _, val := range tc.stackValues {
				vm.ValueStack.PushInt64(val)
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
						assert.Equal(t, v, val.Value_I64)
					}
				}
				assert.Equal(t, tc.expectPC, vm.PC)
			}
		})
	}
}

// Tests for const.i64
func TestCONST_I64(t *testing.T) {
	np := "CONST_I64: "
	tests := []i64TestCase{
		{
			// Should accept little endian immediate value and place i64 on the stack
			name:        np + "Happy Path",
			stackValues: []uint64{},
			memoryContent: []byte{
				wasmvm.OP_CONST_I64, 0xEF, 0xCD, 0xAB, 0x89, 0x67, 0x45, 0x23, 0x01,
			},
			expectTrap:    false,
			expectValue:   []uint64{uint64(0x123456789ABCDEF)},
			expectPC:      9,
			expectedStack: 1,
		},
		{
			// This should detect a trap since there is no enough memory to finish the QWord
			name:        np + "Out of Bounds",
			stackValues: []uint64{},
			memoryContent: []byte{
				wasmvm.OP_CONST_I64, 0x78, 0x56,
			},
			expectTrap:    true,
			trapReason:    np + "Out of bounds",
			expectPC:      0,
			expectedStack: 0,
		},
	}
	runTestBatchI64(t, tests)
}

// Tests for add.i64
func TestADD_I64(t *testing.T) {
	np := "ADD_I64: "
	tests := []i64TestCase{
		{
			// This should detect a trap since there is a lack of i64 values on the stack
			name:        np + "Stack Underflow",
			stackValues: []uint64{},
			memoryContent: []byte{
				wasmvm.OP_ADD_I64,
			},
			expectTrap:    true,
			trapReason:    np + "Stack Underflow",
			expectPC:      0,
			expectedStack: 0,
		},
		{
			// happy path - Small Numbers
			name:        np + "Small Numbers",
			stackValues: []uint64{uint64(5), uint64(7)},
			memoryContent: []byte{
				wasmvm.OP_ADD_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{uint64(12)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			// happy path - Overflow Wrap
			// Going past 0xffffffffffffffff will wrap back to 0, or in this case for +2 to 1
			name:        np + "Overflow Wrap",
			stackValues: []uint64{^uint64(0), uint64(2)}, // ^uint64(0) or 0xffffffffffffffff is QWORD for -1
			memoryContent: []byte{
				wasmvm.OP_ADD_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{uint64(1)},
			expectPC:      1,
			expectedStack: 1,
		},
	}
	runTestBatchI64(t, tests)
}

// Tests for sub.i64
func TestSUB_I64(t *testing.T) {
	np := "SUB_I64: "
	tests := []i64TestCase{
		{
			// This should detect a trap since there is a lack of i64 values on the stack
			name:        np + "Stack Underflow",
			stackValues: []uint64{},
			memoryContent: []byte{
				wasmvm.OP_SUB_I64,
			},
			expectTrap:    true,
			trapReason:    np + "Stack Underflow",
			expectPC:      0,
			expectedStack: 0,
		},
		{
			// happy path - Small Numbers
			name:        np + "Small Numbers",
			stackValues: []uint64{uint64(7), uint64(5)},
			memoryContent: []byte{
				wasmvm.OP_SUB_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{uint64(2)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			// happy path - Overflow Wrap
			// Going past 0 will wrap back to 0xffffffffffffffff, or in this case for -2 from 1 to 0xffffffffffffffff
			name:        np + "Overflow Wrap",
			stackValues: []uint64{uint64(1), uint64(2)},
			memoryContent: []byte{
				wasmvm.OP_SUB_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{^uint64(0)}, // ^uint64(0) or 0xffffffffffffffff is QWORD for -1
			expectPC:      1,
			expectedStack: 1,
		},
	}
	runTestBatchI64(t, tests)
}

// Tests for mul.i64
func TestMUL_I64(t *testing.T) {
	np := "MUL_I64: "
	tests := []i64TestCase{
		{
			// This should detect a trap since there is a lack of i64 values on the stack
			name:        np + "Stack Underflow",
			stackValues: []uint64{},
			memoryContent: []byte{
				wasmvm.OP_MUL_I64,
			},
			expectTrap:    true,
			trapReason:    np + "Stack Underflow",
			expectPC:      0,
			expectedStack: 0,
		},
		{
			// happy path - Small Numbers
			name:        np + "Small Numbers",
			stackValues: []uint64{uint64(7), uint64(5)},
			memoryContent: []byte{
				wasmvm.OP_MUL_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{uint64(35)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			// happy path - Overflow Wrap
			// Going past 0xffffffffffffffff will wrap back to 0xfffffffffffffffe
			name:        np + "Overflow Wrap",
			stackValues: []uint64{^uint64(0), uint64(2)}, // ^uint64(0) or 0xffffffffffffffff is QWORD for -1
			memoryContent: []byte{
				wasmvm.OP_MUL_I64,
			},
			expectTrap: false,
			// 0xFFFFFFFFFFFFFFFF * 2 = 0xFFFFFFFFFFFFFFFE (wraps as unsigned), which as int64 is -2
			expectValue:   []uint64{^uint64(0) - 1},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			// happy path - Multiply by 0
			// The typical whatever * 0 = 0
			name:        np + "Multiply by 0",
			stackValues: []uint64{uint64(137), uint64(0)},
			memoryContent: []byte{
				wasmvm.OP_MUL_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{uint64(0)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			// happy path - Multiply by 0
			// The typical whatever * 1 = whatever
			name:        np + "Multiply by 1",
			stackValues: []uint64{uint64(0x123456789ABCDEF), uint64(1)},
			memoryContent: []byte{
				wasmvm.OP_MUL_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{uint64(0x123456789ABCDEF)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			// happy path - Positive times Negative
			// The typical 7 * -3 = -21
			name:        np + "Positive times Negative",
			stackValues: []uint64{uint64(7), (^uint64(0) - 2)}, // QWord -1 then -2 = -3 or 0xFFFFFFFFFFFFFFFD
			memoryContent: []byte{
				wasmvm.OP_MUL_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{(^uint64(0) - 20)}, // QWord -21 or 0xFFFFFFFFFFFFFFEB
			expectPC:      1,
			expectedStack: 1,
		},
		{
			// happy path - Negative times Negative
			// The typical -2 * -4 = 8
			name:        np + "Positive times Negative",
			stackValues: []uint64{(^uint64(0) - 1), (^uint64(0) - 3)}, // QWord -2 and -4 respectively
			memoryContent: []byte{
				wasmvm.OP_MUL_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{uint64(8)}, // Negative times negative equals positive
			expectPC:      1,
			expectedStack: 1,
		},
	}
	runTestBatchI64(t, tests)
}

// Table tests for div_u.i64
func TestDIVU_I64(t *testing.T) {
	np := "DIVU_I64: "
	tests := []i64TestCase{
		{
			name:        np + "Divide By Zero",
			stackValues: []uint64{1, 0},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I64,
			},
			expectTrap:    true,
			trapReason:    np + "Divide by Zero",
			expectPC:      0,
			expectedStack: 0,
		},
		{
			name:        np + "Small Values - Exact",
			stackValues: []uint64{10, 5},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{2},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Small Values - Non-Exact",
			stackValues: []uint64{11, 5},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{2},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Stack Underflow",
			stackValues: []uint64{uint64(1)},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I64,
			},
			expectTrap:    true,
			trapReason:    np + "Stack Underflow",
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "One by One",
			stackValues: []uint64{1, 1},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{1},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Max value by itself",
			stackValues: []uint64{^uint64(0), ^uint64(0)},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{1},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Large by One",
			stackValues: []uint64{uint64(math.MaxInt64), 1},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{uint64(math.MaxInt64)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Large by Large - Overflows to 1",
			stackValues: []uint64{uint64(math.MaxInt64), uint64(math.MaxInt64)},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{uint64(1)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Zero Dividend",
			stackValues: []uint64{uint64(0), uint64(42)},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{uint64(0)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Dividend Smaller than Divisor",
			stackValues: []uint64{uint64(41), uint64(42)},
			memoryContent: []byte{
				wasmvm.OP_DIVU_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{uint64(0)},
			expectPC:      1,
			expectedStack: 1,
		},
	}
	runTestBatchI64(t, tests)
}

/*
// This was copied over and modified from i32 tests.
// Once verified, it will be useable for testing div_s.i64
// Table tests for div_s.i64
func TestDIVS_I64(t *testing.T) {
	np := "DIVS_I64: "
	tests := []i64TestCase{
		{
			name:        np + "Divide By Zero",
			stackValues: []uint64{1, 0},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I64,
			},
			expectTrap:    true,
			trapReason:    np + "Divide by Zero",
			expectPC:      0,
			expectedStack: 0,
		},
		{
			name:        np + "Small Values - Exact",
			stackValues: []uint64{10, 5},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{2},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Small Values - Non-Exact",
			stackValues: []uint64{11, 5},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{2},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Min Negative by Negative One Division Overflow Trap",
			stackValues: []uint64{uint64(0x8000000000000000), ^uint64(0)},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I64,
			},
			expectTrap:    true,
			trapReason:    np + "Signed Division Overflow",
			expectPC:      0,
			expectedStack: 0,
		},
		{
			name:        np + "Stack Underflow",
			stackValues: []uint64{uint64(1)},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I64,
			},
			expectTrap:    true,
			trapReason:    np + "Stack Underflow",
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Positive One by Positive One",
			stackValues: []uint64{1, 1},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{1},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Positive One by Negative One",
			stackValues: []uint64{1, ^uint64(0)},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{^uint64(0)}, // QWord for -1
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Negative One by Positive One",
			stackValues: []uint64{^uint64(0), 1},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{^uint64(0)}, // QWord for -1
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Negative One by Negative One",
			stackValues: []uint64{^uint64(0), ^uint64(0)},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{1},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Max Positive by Positive One",
			stackValues: []uint64{uint64(math.MaxInt64), 1},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{uint64(math.MaxInt64)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Max Positive by Negative One",
			stackValues: []uint64{uint64(math.MaxInt64), ^uint64(0)},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{uint64(0x8000000000000001)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Max Positive by Max Positive",
			stackValues: []uint64{uint64(math.MaxInt64), uint64(math.MaxInt64)},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{uint64(1)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Zero Dividend",
			stackValues: []uint64{uint64(0), uint64(42)},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{uint64(0)},
			expectPC:      1,
			expectedStack: 1,
		},
		{
			name:        np + "Dividend Smaller than Divisor",
			stackValues: []uint64{uint64(41), uint64(42)},
			memoryContent: []byte{
				wasmvm.OP_DIVS_I64,
			},
			expectTrap:    false,
			expectValue:   []uint64{uint64(0)},
			expectPC:      1,
			expectedStack: 1,
		},
	}
	runTestBatchI64(t, tests)
}
*/
