package main

import (
	"fmt"
	"os"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
)

func main() {
	// Demo: minimal config, empty memory, runs NOP then ADD8
	cfg := &wasmvm.VMConfig{
		Size:   13,
		Rings:  map[uint8]wasmvm.RingConfig{},
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	vm, err := wasmvm.NewVM(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize VM: %v\n", err)
		os.Exit(1)
	}
	// Preload some instructions for testing
	vm.Memory[0] = wasmvm.OP_NOP       // NOP
	vm.Memory[1] = wasmvm.OP_CONST_I32 // const.i32
	vm.Memory[2] = 2                   // Little Endian a
	vm.Memory[3] = 0                   //
	vm.Memory[4] = 0                   //
	vm.Memory[5] = 0
	vm.Memory[6] = wasmvm.OP_CONST_I32 // const.i32
	vm.Memory[7] = 3                   // Little Endian b
	vm.Memory[8] = 0
	vm.Memory[9] = 0
	vm.Memory[10] = 0
	vm.Memory[11] = wasmvm.OP_ADD_I32 // add.i32
	vm.Memory[12] = wasmvm.OP_END     // END: Let's blow this popcicle stand
	// Set PC to 0
	vm.PC = 0
	vm.MainLoop()
	fmt.Printf("Memory after execution: %+v\n", vm.Memory)
	fmt.Printf("Stack after execution: %#v\n", vm.ValueStack)
}
