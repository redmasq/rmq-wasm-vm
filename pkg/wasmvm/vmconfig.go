package wasmvm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
)

//go:generate stringer -type=VMInitializationErrorType
type VMInitializationErrorType byte

const (
	UndefinedVMInitError VMInitializationErrorType = iota
	VMConfigInternalError
	VMConfigRequired
	VMImageError
	MissingSizeOrFlatMemory
	StrictModeAttemptRing0Reconfigure
	VMRingAlreadyExists
)

type VMInitializationError struct {
	Type  VMInitializationErrorType
	Msg   string
	Cause error
	Meta  any
}

type VMErrorMeta[K comparable, O any] struct {
	Key    K `json:"num"`
	Source O `json:"config"`
}

var vmInitDefaultMessageTemplates = map[VMInitializationErrorType]string{
	UndefinedVMInitError:              "unknown VMInitializationErrorType: %s",
	VMConfigRequired:                  "config is required",
	VMConfigInternalError:             "an internal error occurred: %s",
	VMImageError:                      "an error occurred during Image initialization: %s",
	MissingSizeOrFlatMemory:           "either Size or FlatMemory must be specified",
	StrictModeAttemptRing0Reconfigure: "ring 0 cannot be reconfigured (strict mode)",
	VMRingAlreadyExists:               "the ring %d is already present",
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
	Size uint64 // Memory size in bytes
	// TODO [RWV-18]: After some consideration, this needs to be replaced
	// There seems to be support for multiple memory ranges
	// So I'm going to design a memory compositor
	// This will come in handy for the Ring model anyways
	// Since I could add the rwx asserts there
	// Also, it technically needs to be a pointer to []byte
	// Anyways for future host functions
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

// Helper function since AppendRings and AppendExposedFuncs do almost the same thing
func mergeMaps[K comparable, V any, O any](a, b map[K]V, source O) (map[K]V, error) {
	merged := make(map[K]V)
	mergeErr := []error{}
	for k, v := range a {
		merged[k] = v
	}

	for k, v := range b {
		_, exists := merged[k]
		if !exists {
			// Happy Path
			merged[k] = v
			continue
		}
		mergeErr = append(mergeErr, NewVMFluentError(VMErrorMeta[K, O]{Key: k, Source: source}))
	}
	// Need to lookup the convention for returning the call bound object after
	// An error in a fluent interface
	if len(mergeErr) != 0 {
		return nil, errors.Join(mergeErr...)
	}
	return merged, nil
}

func (vmc *VMConfig) SetSize(size uint64) *VMConfig {
	vmc.Size = size
	return vmc
}

func (vmc *VMConfig) SetFlatMemory(fm []byte) *VMConfig {
	vmc.FlatMemory = fm
	return vmc
}

func (vmc *VMConfig) AppendFlatMemory(fm []byte) *VMConfig {
	// I didn't see a concat
	vmc.FlatMemory = slices.Concat(vmc.FlatMemory, fm)
	return vmc
}

func (vmc *VMConfig) SetRingConfig(rc map[uint8]RingConfig) *VMConfig {
	vmc.Rings = rc
	return vmc
}

func (vmc *VMConfig) AppendRingConfig(rc map[uint8]RingConfig) (*VMConfig, error) {
	if vmc.Rings == nil {
		vmc.Rings = rc
		return vmc, nil
	}

	merged, mergeErr := mergeMaps(vmc.Rings, rc, vmc)

	// Need to lookup the convention for returning the call bound object after
	// An error in a fluent interface
	if mergeErr != nil {
		return vmc, mergeErr
	}
	vmc.Rings = merged
	return vmc, nil
}

func (vmc *VMConfig) SetStdin(si io.Reader) *VMConfig {
	vmc.Stdin = si
	return vmc
}

func (vmc *VMConfig) SetStdout(so io.Writer) *VMConfig {
	vmc.Stdout = so
	return vmc
}

func (vmc *VMConfig) SetStderr(se io.Writer) *VMConfig {
	vmc.Stderr = se
	return vmc
}

func (vmc *VMConfig) SetExposedFunc(sef map[string]*ExposedFunc) *VMConfig {
	vmc.ExposedFuncs = sef
	return vmc
}

func (vmc *VMConfig) AppendExposedFunc(ef map[string]*ExposedFunc) (*VMConfig, error) {
	if vmc.ExposedFuncs == nil {
		vmc.ExposedFuncs = ef
		return vmc, nil
	}
	merged, mergeErr := mergeMaps(vmc.ExposedFuncs, ef, vmc)

	// Need to lookup the convention for returning the call bound object after
	// An error in a fluent interface
	if mergeErr != nil {
		return vmc, mergeErr
	}
	vmc.ExposedFuncs = merged
	return vmc, nil
}

func (vmc *VMConfig) SetStartOverride(start uint64) *VMConfig {
	vmc.StartOverride = start
	return vmc
}

// BuildVMState constructs a new VMState from this config.
// Returns (*VMState, error). The config is cloned during build.
func (vmc *VMConfig) BuildVMState() (*VMState, error) {
	return NewVM(vmc)
}

// This will eventually provide a function pointer for
// functions exposed to anything running in the VM
type ExposedFunc struct {
	Parameters map[string]interface{} // Metadata for the function
	Function   func(*VMState, ...interface{}) error
}

func NewVMFluentError[K comparable, O any](meta VMErrorMeta[K, O]) error {
	return &VMInitializationError{
		Type:  VMRingAlreadyExists,
		Msg:   VmInitErrStr(VMRingAlreadyExists, meta.Key),
		Cause: nil,
		Meta:  meta,
	}
}

// Number of octets in types so far
// As a side note, I'm working with uint values
// unless otherwise specified
const (
	WidthI32 = 4
	WidthI64 = 8
	WidthF32 = 4
)
