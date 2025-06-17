package main

import (
	"fmt"
	"os"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
)

func main() {
	// Demo: minimal config, empty memory, runs NOP then ADD8
	cfg := &wasmvm.VMConfig{
		Size:   8,
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
	vm.Memory[0] = 0x00 // NOP
	vm.Memory[1] = 0x01 // ADD8
	vm.Memory[2] = 2    // a = 2
	vm.Memory[3] = 3    // b = 3
	vm.Memory[4] = 0    // destination
	vm.Memory[5] = 0x0B // END: Let's blow this popcicle stand
	// Set PC to 0
	vm.PC = 0
	vm.MainLoop()
	fmt.Printf("Memory after execution: %+v\n", vm.Memory)
}
