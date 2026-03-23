package wasmvm_test

import (
	"errors"
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
)

func TestTrapTypeString(t *testing.T) {
	assert.Equal(t, "TrapStackUnderflow", wasmvm.TrapStackUnderflow.String())
	assert.Equal(t, "TrapType(99)", wasmvm.TrapType(99).String())
}

func TestTrapAccessTypeString(t *testing.T) {
	assert.Equal(t, "TrapAccessExecute", wasmvm.TrapAccessExecute.String())
	assert.Equal(t, "TrapAccessType(99)", wasmvm.TrapAccessType(99).String())
}

func TestTrapErrStr(t *testing.T) {
	assert.Equal(t, "divide by zero", wasmvm.TrapErrStr(wasmvm.TrapDivideByZero))
	assert.Contains(t, wasmvm.TrapErrStr(wasmvm.TrapType(99)), "unknown trap error")
}

func TestNewTrapError(t *testing.T) {
	err := wasmvm.NewTrapError(wasmvm.TrapStackUnderflow, "stack underflow")

	assert.IsType(t, &wasmvm.TrapError{}, err)
	assert.Equal(t, &wasmvm.TrapError{
		Type:    wasmvm.TrapStackUnderflow,
		Message: "stack underflow",
	}, err)
}

func TestNewTrapErrorWithCauseOrMeta(t *testing.T) {
	cause := errors.New("root cause")
	meta := map[string]any{"foo": "bar"}

	err := wasmvm.NewTrapErrorWithCauseOrMeta(wasmvm.TrapInternalError, "internal trap error", cause, meta)

	assert.IsType(t, &wasmvm.TrapError{}, err)
	assert.Equal(t, &wasmvm.TrapError{
		Type:    wasmvm.TrapInternalError,
		Message: "internal trap error",
		Cause:   cause,
		Meta:    meta,
	}, err)
}

func TestTrapError_Error(t *testing.T) {
	err := &wasmvm.TrapError{
		Type:    wasmvm.TrapDivideByZero,
		Op:      "DIVS_I32",
		Message: "divide by zero",
	}

	assert.Equal(t, "[TrapDivideByZero] DIVS_I32: divide by zero", err.Error())
}

func TestTrapError_Unwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := &wasmvm.TrapError{
		Type:    wasmvm.TrapInternalError,
		Message: "internal trap error",
		Cause:   cause,
	}

	assert.Equal(t, cause, err.Unwrap())
}
