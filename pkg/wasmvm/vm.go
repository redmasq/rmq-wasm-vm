package wasmvm

import "fmt"

// The actual VM state itself. Right now, we are only assuming a
// single execution context. I'll need to refactor this when
// rings get added, which is the prereq to my threading model
// since I'm ignoring the "mark memory as shared" proposed
// standard for which I stumbled upon for something I think
// will allow for easier porting
type VMState struct {
	Memory         []byte
	PC             uint64 // Program Counter
	Trap           bool
	TrapErr        *TrapError
	ImageInitWarn  []string
	Config         *VMConfig
	InstructionMap map[uint8]Instruction
	StateStack     []VMState
	ValueStack     ValueStack

	// Add more state as needed
}

func NewVMInitializationError(eType VMInitializationErrorType, msg string) error {
	return &VMInitializationError{
		Type: eType,
		Msg:  msg,
	}
}

func NewVMInitializationErrorWithCauseOrMeta(eType VMInitializationErrorType, msg string, cause error, meta any) error {
	return &VMInitializationError{
		Type:  eType,
		Msg:   msg,
		Cause: cause,
		Meta:  meta,
	}
}

func (vm *VMState) SetTrapError(trap *TrapError) error {
	if trap == nil {
		trap = &TrapError{
			Type:    TrapInternalError,
			Message: "unknown trap",
		}
	}
	vm.Trap = true
	vm.TrapErr = trap
	return trap
}

// NewVM - Accepts VMConfig and returns a constructed VMState or error
// Note that the VMState.config is a modified clone of the original
func NewVM(config *VMConfig) (*VMState, error) {
	if config == nil {
		return nil, NewVMInitializationError(VMConfigRequired, VmInitErrStr(VMConfigRequired))
	}

	// We clone the config so that we own this copy
	// It allows for the config to be recycled without
	// worry
	vc, err := config.QuickClone()
	if err != nil {
		return nil, NewVMInitializationErrorWithCauseOrMeta(VMConfigInternalError, VmInitErrStr(VMConfigInternalError, err.Error()), err, nil)
	}

	if vc.Size == 0 && vc.FlatMemory == nil {
		return nil, NewVMInitializationError(MissingSizeOrFlatMemory, VmInitErrStr(MissingSizeOrFlatMemory))
	}

	vc.Stdin = config.Stdin
	vc.Stdout = config.Stdout
	vc.Stderr = config.Stderr
	vc.ExposedFuncs = config.ExposedFuncs

	mem := vc.FlatMemory
	if mem == nil {
		mem = make([]byte, config.Size)
	}
	state := &VMState{
		Memory:         mem,
		PC:             0,
		Trap:           false,
		Config:         vc,
		InstructionMap: defaultInstructionMap(),
	}
	// Populate memory/image via config.Image (see image.go)
	if vc.Image != nil {
		warns, err := PopulateImage(mem, vc.Image, vc.Strict)
		state.ImageInitWarn = warns
		if err != nil {
			if vc.Strict {
				return nil, NewVMInitializationErrorWithCauseOrMeta(VMImageError, VmInitErrStr(VMImageError, err), err, vc.Image)
			}
			// It's strange, I walk this line in debug,
			// but it doesn't show under coverage in VS Code
			// And stranger yet, warn - NewVM Type SparseArray Lenient
			// tests for the result
			state.ImageInitWarn = append(state.ImageInitWarn, err.Error())
		}
	}
	// Initialize rings
	if vc.Rings == nil {
		vc.Rings = make(map[uint8]RingConfig)
	}
	rc, ok := vc.Rings[0]

	// Ring 0 is always full access; ignore/override if defined
	if ok && (rc.Enabled || vc.Strict) {
		if vc.Strict {
			return nil, NewVMInitializationError(StrictModeAttemptRing0Reconfigure, VmInitErrStr(StrictModeAttemptRing0Reconfigure))
		}
		state.ImageInitWarn = append(state.ImageInitWarn, "Ring 0 redefinition ignored")
	}
	vc.Rings[0] = RingConfig{Enabled: true}

	// Set start point
	if vc.StartOverride != 0 {
		state.PC = vc.StartOverride
	}

	return state, nil
}

// Operates on a VMState - This fetches the next instruction and acts upon
// it using the configured InstructionMap. May return an error.
func (vm *VMState) Step() error {
	if vm.Trap {
		if vm.TrapErr != nil {
			return vm.TrapErr
		}
		return &TrapError{
			Type:    TrapInternalError,
			Op:      "STEP",
			PC:      vm.PC,
			Message: "execution trapped with no TrapErr",
		}
	}
	if vm.PC >= uint64(len(vm.Memory)) {
		return vm.SetTrapError(&TrapError{
			Type:    TrapProgramCounterOutOfBounds,
			Op:      "STEP",
			PC:      vm.PC,
			Message: "Program counter out of bounds",
		})
	}
	opcode := vm.Memory[vm.PC]
	handler, ok := vm.InstructionMap[opcode]
	if !ok {
		return vm.SetTrapError(&TrapError{
			Type:        TrapUnknownInstruction,
			Op:          "STEP",
			PC:          vm.PC,
			Message:     fmt.Sprintf("Unknown instruction: 0x%02X", opcode),
			Instruction: &opcode,
		})
	}
	return handler(vm)
}

// Operates on VMState - Calls vm.Step() until trap is reached
func (vm *VMState) MainLoop() {
	for !vm.Trap {
		err := vm.Step()
		if err != nil && vm.Config != nil && vm.Config.Stderr != nil {
			if vm.TrapErr == nil || vm.TrapErr.Type != TrapCallStackEmpty {
				fmt.Fprintf(vm.Config.Stderr, "Execution error: %v\n", err)
			}
		}
	}
}
