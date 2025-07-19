package main_test

import (
	"testing"

	cmdwasmvm "github.com/redmasq/rmq-wasm-vm/cmd/wasmvm"
	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"

	"github.com/stretchr/testify/assert"
)

// TestSetupAndRunVM uses a test table to do various combinations of executions
// and then verifies the stack
func TestSetupAndRunVM(t *testing.T) {
	tests := []cmdwasmvm.ExecutionContext{
		{
			Id:    0,
			Title: "2 + 3 = 5",
			Want:  uint32(5),
			WantT: wasmvm.TYPE_I32,
			Program: []byte{
				wasmvm.OP_CONST_I32, 2, 0, 0, 0,
				wasmvm.OP_CONST_I32, 3, 0, 0, 0,
				wasmvm.OP_ADD_I32,
				wasmvm.OP_END,
			},
		},
		{
			Id:    1,
			Title: "NOP, 5 * 8 = 40",
			Want:  uint32(40),
			WantT: wasmvm.TYPE_I32,
			Program: []byte{
				wasmvm.OP_NOP,
				wasmvm.OP_CONST_I32, 5, 0, 0, 0,
				wasmvm.OP_CONST_I32, 8, 0, 0, 0,
				wasmvm.OP_MUL_I32,
				wasmvm.OP_END,
			},
		},
		{
			Id:    1,
			Title: "NOP, 8 - 5 = 3, NOP",
			Want:  uint64(3),
			WantT: wasmvm.TYPE_I64,
			Program: []byte{
				wasmvm.OP_NOP,
				wasmvm.OP_CONST_I64, 8, 0, 0, 0, 0, 0, 0, 0,
				wasmvm.OP_CONST_I64, 5, 0, 0, 0, 0, 0, 0, 0,
				wasmvm.OP_SUB_I64,
				wasmvm.OP_END,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Title, func(t *testing.T) {
			vm, err := cmdwasmvm.SetupAndRunVM(&tc)
			assert.NoError(t, err, "setupAndRunVM failed: %v", err)

			assert.Equal(t, 1, vm.ValueStack.Size(), "Incorrect number of entries in stack")
			success, collect := vm.ValueStack.HasAtLeastOfType(1, tc.WantT)
			assert.True(t, success, "There should be at least one entry of the required type %s", tc.WantT.String())
			assert.Equal(t, len(collect), 1, "Wrong size for the stack")
			switch tc.WantT {
			case wasmvm.TYPE_I32:
				assert.Equal(t, tc.Want, collect[0].Value_I32)
			case wasmvm.TYPE_I64:
				assert.Equal(t, tc.Want, collect[0].Value_I64)
			default:
				assert.Fail(t, "Unexpected test type %s", tc.WantT)
			}
		})
	}
}
