package wasmvm

// 0x01 NOP: No Operation
func NOP(vm *VMState) error {
	vm.PC++
	return nil
}

// 0x0B
func END(vm *VMState) error {
	// Stub
	// TODO: Add handling when there are entries on the call stack for the execution frame
	vm.Trap = true
	vm.TrapReason = "End of final function"
	return nil
}
