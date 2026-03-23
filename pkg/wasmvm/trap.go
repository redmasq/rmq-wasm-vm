package wasmvm

import "fmt"

type TrapType byte

const (
	UndefinedTrap TrapType = iota
	TrapUnknownInstruction
	TrapProgramCounterOutOfBounds
	TrapCallStackEmpty
	TrapStackUnderflow
	TrapStackCleanup
	TrapDivideByZero
	TrapSignedDivisionOverflow
	TrapMemoryAccess
	TrapHostFunction
	TrapInternalError
)

var trapTypeNames = map[TrapType]string{
	UndefinedTrap:              "UndefinedTrap",
	TrapUnknownInstruction:     "TrapUnknownInstruction",
	TrapProgramCounterOutOfBounds: "TrapProgramCounterOutOfBounds",
	TrapCallStackEmpty:         "TrapCallStackEmpty",
	TrapStackUnderflow:         "TrapStackUnderflow",
	TrapStackCleanup:           "TrapStackCleanup",
	TrapDivideByZero:           "TrapDivideByZero",
	TrapSignedDivisionOverflow: "TrapSignedDivisionOverflow",
	TrapMemoryAccess:           "TrapMemoryAccess",
	TrapHostFunction:           "TrapHostFunction",
	TrapInternalError:          "TrapInternalError",
}

func (t TrapType) String() string {
	if name, ok := trapTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("TrapType(%d)", t)
}

type TrapAccessType byte

const (
	TrapAccessUnknown TrapAccessType = iota
	TrapAccessRead
	TrapAccessWrite
	TrapAccessExecute
)

var trapAccessTypeNames = map[TrapAccessType]string{
	TrapAccessUnknown: "TrapAccessUnknown",
	TrapAccessRead:    "TrapAccessRead",
	TrapAccessWrite:   "TrapAccessWrite",
	TrapAccessExecute: "TrapAccessExecute",
}

func (t TrapAccessType) String() string {
	if name, ok := trapAccessTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("TrapAccessType(%d)", t)
}

type TrapError struct {
	Type        TrapType
	Op          string
	PC          uint64
	Message     string
	Cause       error
	AccessType  TrapAccessType
	Address     *uint64
	Ring        *uint8
	Instruction *uint8
	Meta        any
}

var trapDefaultMessageTemplates = map[TrapType]string{
	UndefinedTrap:              "undefined trap",
	TrapUnknownInstruction:     "unknown instruction trap",
	TrapProgramCounterOutOfBounds: "program counter out of bounds",
	TrapCallStackEmpty:         "call stack empty",
	TrapStackUnderflow:         "stack underflow",
	TrapStackCleanup:           "stack cleanup error",
	TrapDivideByZero:           "divide by zero",
	TrapSignedDivisionOverflow: "signed division overflow",
	TrapMemoryAccess:           "memory access trap",
	TrapHostFunction:           "host function trap",
	TrapInternalError:          "internal trap error",
}

func TrapErrStr(t TrapType, paras ...any) string {
	msg, ok := trapDefaultMessageTemplates[t]
	if !ok || msg == "" {
		msg = fmt.Sprintf("unknown trap error[%s,%d]", t.String(), t)
	}
	if len(paras) > 0 {
		return fmt.Sprintf(msg, paras...)
	}
	return msg
}

func NewTrapError(t TrapType, msg string) error {
	return &TrapError{
		Type:    t,
		Message: msg,
	}
}

func NewTrapErrorWithCauseOrMeta(t TrapType, msg string, cause error, meta any) error {
	return &TrapError{
		Type:    t,
		Message: msg,
		Cause:   cause,
		Meta:    meta,
	}
}

func (e *TrapError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Op != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Type.String(), e.Op, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Type.String(), e.Message)
}

func (e *TrapError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}
