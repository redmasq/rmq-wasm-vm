package wasmvm

import (
	"encoding/json"
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
	Rings         map[uint8]RingConfig    // 0-255
	Stdin         io.Reader               `json:"-"`
	Stdout        io.Writer               `json:"-"`
	Stderr        io.Writer               `json:"-"`
	ExposedFuncs  map[string]*ExposedFunc `json:"-"`
	StartOverride uint64                  // Optional entry point override
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
const (
	WidthI32 = 4
	WidthI64 = 8
	WidthF32 = 4
)

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

//go:generate stringer -type=VMInitializationErrorType
type VMInitializationErrorType byte

const (
	UndefinedVMInitError VMInitializationErrorType = iota
	VMConfigInternalError
	VMConfigRequired
	VMImageError
	MissingSizeOrFlatMemory
	StrictModeAttemptRing0Reconfigure
)

type VMInitializationError struct {
	Type  VMInitializationErrorType
	Msg   string
	Cause error
	Meta  any
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

var vmInitDefaultMessageTemplates = map[VMInitializationErrorType]string{
	UndefinedVMInitError:              "unknown VMInitializationErrorType: %s",
	VMConfigRequired:                  "config is required",
	VMConfigInternalError:             "an internal error occurred: %s",
	VMImageError:                      "an error occurred during Image initialization: %s",
	MissingSizeOrFlatMemory:           "either Size or FlatMemory must be specified",
	StrictModeAttemptRing0Reconfigure: "ring 0 cannot be reconfigured (strict mode)",
}

func VmInitErrStr(eType VMInitializationErrorType, paras ...any) string {
	ermsg, ok := vmInitDefaultMessageTemplates[eType]
	if !ok || ermsg == "" {
		ermsg = fmt.Sprintf("unknown vm initialization error[%s,%d]", eType.String(), eType)
	}
	if len(paras) > 0 {
		return fmt.Sprintf(ermsg, paras...)
	}
	return ermsg
}

// Implement the `error` interface
func (e *VMInitializationError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type.String(), e.Msg)
}

// Another from the `error` interface
func (e *VMInitializationError) Unwrap() error {
	return e.Cause
}

// This won't include the stdin, etc or the exposed functions
func (vmc *VMConfig) QuickClone() (*VMConfig, error) {
	if vmc == nil {
		return nil, nil
	}
	origJson, err := json.Marshal(vmc)
	if err != nil {
		return nil, err
	}
	clone := &VMConfig{}
	err = json.Unmarshal(origJson, clone)
	if err != nil {
		return nil, err
	}
	return clone, nil
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
