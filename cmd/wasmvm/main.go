package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
)

type ExecutionContext struct {
	Id      int
	Title   string
	Want    any // Test support intentionally leaked
	WantT   wasmvm.ValueStackEntryType
	Program []byte
}

var executions = []ExecutionContext{
	{
		Id:    0,
		Title: "2 + 3 = 5",
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
		Program: []byte{
			wasmvm.OP_NOP,
			wasmvm.OP_CONST_I64, 5, 0, 0, 0, 0, 0, 0, 0,
			wasmvm.OP_CONST_I64, 8, 0, 0, 0, 0, 0, 0, 0,
			wasmvm.OP_SUB_I64,
			wasmvm.OP_END,
		},
	},
}

// SetupAndRunVM initializes a WASM VM with the given ExecutionContext, loads
// its program into memory, and executes it. Returns the final VM state
// and any error encountered during setup or execution.
func SetupAndRunVM(context *ExecutionContext) (*wasmvm.VMState, error) {
	fmt.Printf("Starting %d with title of %s", context.Id, context.Title)
	size := uint64(len(context.Program))
	cfg := &wasmvm.VMConfig{
		Image: &wasmvm.ImageConfig{
			Type:  wasmvm.Array,
			Size:  size,
			Array: context.Program,
		},
		Size:   size,
		Rings:  map[uint8]wasmvm.RingConfig{},
		Stdin:  nil,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	vm, err := wasmvm.NewVM(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize vm: %v", err)
	}

	// Set PC to 0
	vm.PC = 0
	vm.MainLoop()
	fmt.Printf("(%d) Memory after execution: %+v\n", context.Id, vm.Memory)
	fmt.Printf("(%d) Stack after execution: %#v\n", context.Id, vm.ValueStack)

	return vm, nil
}

// wrapGoRoutine runs SetupAndRunVM in a goroutine for the provided context,
// and calls wg.Done() when finished. Intended for use with sync.WaitGroup
// to coordinate concurrent VM execution.
// This method is not part of the test suite
func wrapGoRoutine(context *ExecutionContext, wg *sync.WaitGroup) {
	defer wg.Done()
	SetupAndRunVM(context)
}

// main is the entry point for this program. It defines a set of VM execution
// contexts, then runs each in parallel goroutines, waiting for all to finish.
// This method is not part of the test suite
func main() {
	var wg sync.WaitGroup
	for i := range executions {
		context := &executions[i]
		wg.Add(1)
		go wrapGoRoutine(context, &wg)
	}
	wg.Wait()
}
