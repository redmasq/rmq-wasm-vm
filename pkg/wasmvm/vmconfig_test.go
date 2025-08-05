package wasmvm_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
)

type fluentTestCase struct {
	name                  string
	testCase              func() (*wasmvm.VMConfig, error)
	expectSize            uint64
	expectMemory          []byte
	expectRings           map[uint8]wasmvm.RingConfig
	expectEFunc           map[string]*wasmvm.ExposedFunc
	expectError           bool
	expectStdin           io.Reader
	expectStdout          io.Writer
	expectStderr          io.Writer
	expectStartupOverride uint64
}

func TestVMConfig_FluentAPI(t *testing.T) {
	dummyExposedFunc := func(*wasmvm.VMState, ...interface{}) error {
		return nil
	}
	in := new(io.Reader)
	out := new(io.Writer)

	tests := []fluentTestCase{
		{
			name: "success - SetSize only",
			testCase: func() (*wasmvm.VMConfig, error) {
				return new(wasmvm.VMConfig).SetSize(123), nil
			},
			expectSize: 123,
		},
		{
			name: "success - SetFlatMemory then AppendFlatMemory",
			testCase: func() (*wasmvm.VMConfig, error) {
				return new(wasmvm.VMConfig).
					SetFlatMemory([]byte{1, 2}).
					AppendFlatMemory([]byte{3, 4}), nil
			},
			expectMemory: []byte{1, 2, 3, 4},
		},
		{
			name: "success - SetRingConfig",
			testCase: func() (*wasmvm.VMConfig, error) {
				return new(wasmvm.VMConfig).
					SetRingConfig(map[uint8]wasmvm.RingConfig{1: {Enabled: true}}), nil
			},
			expectRings: map[uint8]wasmvm.RingConfig{1: {Enabled: true}},
		},
		{
			name: "success - AppendRingConfig",
			testCase: func() (*wasmvm.VMConfig, error) {
				conf := new(wasmvm.VMConfig).
					SetRingConfig(map[uint8]wasmvm.RingConfig{2: {Enabled: true}, 3: {Enabled: true}})
				conf, err := conf.AppendRingConfig(map[uint8]wasmvm.RingConfig{4: {Enabled: false}})
				// Should error, but config stays as before
				return conf, err
			},
			expectRings: map[uint8]wasmvm.RingConfig{2: {Enabled: true}, 3: {Enabled: true}, 4: {Enabled: false}},
		},
		{
			name: "failure - AppendRingConfig with conflict",
			testCase: func() (*wasmvm.VMConfig, error) {
				conf := new(wasmvm.VMConfig).
					SetRingConfig(map[uint8]wasmvm.RingConfig{2: {Enabled: true}, 3: {Enabled: true}})
				conf, err := conf.AppendRingConfig(map[uint8]wasmvm.RingConfig{2: {Enabled: false}, 4: {Enabled: false}})
				// Should error, but config stays as before
				return conf, err
			},
			expectRings: map[uint8]wasmvm.RingConfig{2: {Enabled: true}, 3: {Enabled: true}},
			expectError: true,
		},
		{
			name: "success - AppendRingConfig with empty",
			testCase: func() (*wasmvm.VMConfig, error) {
				conf := new(wasmvm.VMConfig)
				conf, err := conf.AppendRingConfig(map[uint8]wasmvm.RingConfig{2: {Enabled: false}, 4: {Enabled: false}})
				// Should error, but config stays as before
				return conf, err
			},
			expectRings: map[uint8]wasmvm.RingConfig{2: {Enabled: false}, 4: {Enabled: false}},
		},
		{
			name: "success - AppendExposeFunc",
			testCase: func() (*wasmvm.VMConfig, error) {
				conf := new(wasmvm.VMConfig).SetExposedFunc(map[string]*wasmvm.ExposedFunc{
					"temp": {
						Parameters: nil,
						Function:   &dummyExposedFunc,
					},
				})
				conf, err := conf.AppendExposedFunc(map[string]*wasmvm.ExposedFunc{
					"temp2": {
						Parameters: nil,
						Function:   &dummyExposedFunc,
					},
				})

				return conf, err
			},
			expectEFunc: map[string]*wasmvm.ExposedFunc{
				"temp": {
					Parameters: nil,
					Function:   &dummyExposedFunc,
				},
				"temp2": {
					Parameters: nil,
					Function:   &dummyExposedFunc,
				},
			},
		},
		{
			name: "success - AppendExposeFunc with empty",
			testCase: func() (*wasmvm.VMConfig, error) {
				conf := new(wasmvm.VMConfig)
				conf, err := conf.AppendExposedFunc(map[string]*wasmvm.ExposedFunc{
					"temp2": {
						Parameters: nil,
						Function:   &dummyExposedFunc,
					},
				})

				return conf, err
			},
			expectEFunc: map[string]*wasmvm.ExposedFunc{
				"temp2": {
					Parameters: nil,
					Function:   &dummyExposedFunc,
				},
			},
		},
		{
			name: "failure - AppendExposeFunc",
			testCase: func() (*wasmvm.VMConfig, error) {
				conf := new(wasmvm.VMConfig).SetExposedFunc(map[string]*wasmvm.ExposedFunc{
					"temp": {
						Parameters: nil,
						Function:   &dummyExposedFunc,
					},
				})
				conf, err := conf.AppendExposedFunc(map[string]*wasmvm.ExposedFunc{
					"temp": {
						Parameters: nil,
						Function:   &dummyExposedFunc,
					},
				})

				return conf, err
			},
			expectEFunc: map[string]*wasmvm.ExposedFunc{
				"temp": {
					Parameters: nil,
					Function:   &dummyExposedFunc,
				},
			},
			expectError: true,
		},
		{
			name: "success - SetStdin Only",
			testCase: func() (*wasmvm.VMConfig, error) {
				return new(wasmvm.VMConfig).SetStdin(*in), nil
			},
			expectStdin: *in,
		},
		{
			name: "success - SetStdout Only",
			testCase: func() (*wasmvm.VMConfig, error) {
				return new(wasmvm.VMConfig).SetStdout(*out), nil
			},
			expectStdout: *out,
		},
		{
			name: "success - SetStderr Only",
			testCase: func() (*wasmvm.VMConfig, error) {
				return new(wasmvm.VMConfig).SetStderr(*out), nil
			},
			expectStderr: *out,
		},
		{
			name: "success - SetStartOverride Only",
			testCase: func() (*wasmvm.VMConfig, error) {
				return new(wasmvm.VMConfig).SetStartOverride(137), nil
			},
			expectStartupOverride: 137,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conf, err := test.testCase()
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if test.expectSize != 0 {
				assert.Equal(t, test.expectSize, conf.Size)
			}
			if test.expectMemory != nil {
				assert.Equal(t, test.expectMemory, conf.FlatMemory)
			}
			if test.expectRings != nil {
				assert.Equal(t, test.expectRings, conf.Rings)
			}
			if test.expectRings != nil {
				assert.Equal(t, test.expectRings, conf.Rings)
			}
			if test.expectEFunc != nil {
				assert.Equal(t, test.expectEFunc, conf.ExposedFuncs)
			}
			if test.expectStdin != nil {
				assert.Equal(t, in, conf.Stdin)
			}
			if test.expectStdout != nil {
				assert.Equal(t, out, conf.Stdout)
			}
			if test.expectStderr != nil {
				assert.Equal(t, out, conf.Stderr)
			}
			if test.expectStartupOverride > 0 {
				assert.Equal(t, test.expectStartupOverride, conf.StartOverride)
			}
		})
	}
}

func TestCloneEmpty(t *testing.T) {
	var vmc *wasmvm.VMConfig
	testvmc, err := vmc.QuickClone()
	assert.Nil(t, testvmc)
	assert.NoError(t, err)
}

func TestErrStr(t *testing.T) {
	typ := wasmvm.VMInitializationErrorType(byte(wasmvm.VMRingAlreadyExists) + 1)
	errStr := wasmvm.VmInitErrStr(typ)
	assert.Contains(t, errStr, "unknown vm initialization error")
	err := &wasmvm.VMInitializationError{
		Type:  typ,
		Msg:   errStr,
		Cause: fmt.Errorf("Test Error"),
	}
	assert.Contains(t, err.Error(), "unknown vm initialization error")
	assert.Contains(t, err.Unwrap().Error(), "Test Error")
}
