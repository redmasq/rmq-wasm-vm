package wasmvm

import (
	"errors"
	"fmt"
	"io"
)

// TODO: Stub for when ring support comes much later
type RingConfig struct {
	Enabled bool
	// Add additional properties later
}

// This contains the basic structure of the initialization
// with the ability to define various parameters. Some of
// them are meant to be unchanged once execution starts,
// thus we have a different struct
type VMConfig struct {
	Size          uint64 // Memory size in bytes
	FlatMemory    []byte // Optional: existing memory
	Strict        bool
	Image         *ImageConfig
	Rings         map[uint8]RingConfig // 0-255
	Stdin         io.Reader
	Stdout        io.Writer
	Stderr        io.Writer
	ExposedFuncs  map[string]*ExposedFunc
	StartOverride *uint64 // Optional entry point override
}

// This will eventually provide a function pointer for
// functions exposed to anything running in the VM
type ExposedFunc struct {
	Parameters map[string]interface{} // Metadata for the function
	Function   func(*VMState, ...interface{}) error
}

// Number of octets in types so far
// As a side note, I'm working with uint values
// unless otherwise specified
const WIDTH_I32 = 4
const WIDTH_I64 = 8
const WIDTH_F32 = 4

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
	TrapReason     string
	ImageInitWarn  []string
	Config         *VMConfig
	InstructionMap map[uint8]Instruction
	StateStack     []VMState
	ValueStack     ValueStack

	// Add more state as needed
}

// NewVM - Accepts VMConfig and returns a constructed VMConfig or error
func NewVM(config *VMConfig) (*VMState, error) {
	if config.Size == 0 && config.FlatMemory == nil {
		return nil, errors.New("either Size or FlatMemory must be specified")
	}
	mem := config.FlatMemory
	if mem == nil {
		mem = make([]byte, config.Size)
	}
	state := &VMState{
		Memory:         mem,
		PC:             0,
		Trap:           false,
		Config:         config,
		InstructionMap: defaultInstructionMap(),
	}
	// Populate memory/image via config.Image (see image.go)
	if config.Image != nil {
		warns, err := PopulateImage(mem, config.Image, config.Strict)
		state.ImageInitWarn = warns
		if err != nil {
			if config.Strict {
				return nil, err
			}
			state.ImageInitWarn = append(state.ImageInitWarn, err.Error())
		}
	}
	// Initialize rings
	if config.Rings == nil {
		config.Rings = make(map[uint8]RingConfig)
	}
	// Ring 0 is always full access; ignore/override if defined
	if rc, ok := config.Rings[0]; ok && (rc.Enabled || config.Strict) {
		if config.Strict {
			return nil, errors.New("ring 0 cannot be reconfigured (strict mode)")
		}
		state.ImageInitWarn = append(state.ImageInitWarn, "Ring 0 redefinition ignored")
	}
	config.Rings[0] = RingConfig{Enabled: true}

	// Set start point
	if config.StartOverride != nil {
		state.PC = *config.StartOverride
	}

	return state, nil
}

// Operates on a VMState - This fetches the next instruction and acts upon
// it using the configured InstructionMap. May return an error.
func (vm *VMState) Step() error {
	if vm.Trap {
		return fmt.Errorf("execution trapped: %s", vm.TrapReason)
	}
	if vm.PC >= uint64(len(vm.Memory)) {
		vm.Trap = true
		vm.TrapReason = "Program counter out of bounds"
		return errors.New(vm.TrapReason)
	}
	opcode := vm.Memory[vm.PC]
	handler, ok := vm.InstructionMap[opcode]
	if !ok {
		vm.Trap = true
		vm.TrapReason = fmt.Sprintf("Unknown instruction: 0x%02X", opcode)
		return errors.New(vm.TrapReason)
	}
	return handler(vm)
}

// Operates on VMState - Calls vm.Step() until trap is reached
func (vm *VMState) MainLoop() {
	for !vm.Trap {
		err := vm.Step()
		if err != nil && vm.Config != nil && vm.Config.Stderr != nil {
			if vm.TrapReason != "END: Call Stack Empty" {
				fmt.Fprintf(vm.Config.Stderr, "Execution error: %v\n", err)
			}
		}
	}
}
