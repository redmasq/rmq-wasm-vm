package wasmvm

import (
	"errors"
	"fmt"
	"io"
)

type RingConfig struct {
	Enabled bool
	// Add additional properties later
}

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

type ExposedFunc struct {
	Parameters map[string]interface{} // Metadata for the function
	Function   func(*VMState, ...interface{}) error
}

const WIDTH_I32 = 4
const WIDTH_F32 = 4

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

func (vm *VMState) ExecuteNext() error {
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

func (vm *VMState) MainLoop() {
	for !vm.Trap {
		err := vm.ExecuteNext()
		if err != nil && vm.Config != nil && vm.Config.Stderr != nil {
			fmt.Fprintf(vm.Config.Stderr, "Execution error: %v\n", err)
		}
	}
}
