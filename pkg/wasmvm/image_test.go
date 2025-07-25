package wasmvm_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/redmasq/rmq-wasm-vm/pkg/wasmvm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type readFileMock func(string) ([]byte, error)
type testType int8

const (
	testFile testType = iota
	testArray
	testSparse
	testEmpty
	testOther
)

const _testType_name = "testFiletestArraytestSparsetestEmptytestOther"

var _testType_index = [...]uint8{0, 8, 17, 27, 36, 45}

func (i testType) String() string {
	if i < 0 || i >= testType(len(_testType_index)-1) {
		return "testType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _testType_name[_testType_index[i]:_testType_index[i+1]]
}

func GetImageErrorType(err error) wasmvm.ImageInitializationErrorType {
	var imgErr *wasmvm.ImageInitializationError
	if errors.As(err, &imgErr) {
		return imgErr.Type
	}
	return wasmvm.UndefinedImageError
}

type imageTestCase struct {
	name              string
	tType             testType
	mockReadFile      readFileMock
	mockArray         []byte
	expectError       bool
	expertWarns       bool
	checkErrorType    bool
	checkErrorCause   bool
	errorType         *wasmvm.ImageInitializationError
	errorContains     string
	warnsContains     string
	prepopulateMemory []byte
	memoryContains    []byte
	imageSize         int
	forceImageSize    bool
	memorySize        int
	useStrict         bool
}

func TestPopulateImage(t *testing.T) {
	tests := []imageTestCase{
		{
			name:  "success - same size",
			tType: testFile,
			mockReadFile: func(string) ([]byte, error) {
				return []byte{0xAB, 0xCD}, nil
			},
			expectError:    false,
			expertWarns:    false,
			memoryContains: []byte{0xAB, 0xCD},
			memorySize:     2,
			useStrict:      true,
		},
		{
			name:  "success - smaller size",
			tType: testFile,
			mockReadFile: func(string) ([]byte, error) {
				return []byte{0xAB, 0xCD}, nil
			},
			expectError:    false,
			expertWarns:    false,
			memoryContains: []byte{0xAB, 0xCD, 0, 0},
			memorySize:     4,
			useStrict:      true,
		},
		{
			name:  "success - empty file",
			tType: testFile,
			mockReadFile: func(string) ([]byte, error) {
				return []byte{}, nil
			},
			expectError:    false,
			expertWarns:    false,
			memoryContains: []byte{0x00, 0x00, 0x00, 0x00},
			memorySize:     4,
			useStrict:      true,
		},
		{
			name:  "failure - read error",
			tType: testFile,
			mockReadFile: func(string) ([]byte, error) {
				return nil, errors.New("I/O Error because \"reasons\"")
			},
			expectError:    true,
			expertWarns:    false,
			memoryContains: nil,
			checkErrorType: true,
			errorType: &wasmvm.ImageInitializationError{
				Type:  wasmvm.FileImageOtherError,
				Msg:   "I/O Error because \"reasons\"", // TODO: Fix the message check
				Cause: fmt.Errorf("Whatever"),          // Message isn't checked, just type
			},
			memorySize: 4,
			useStrict:  true,
		},
		{
			name:  "warn - oversized file",
			tType: testFile,
			mockReadFile: func(string) ([]byte, error) {
				return []byte{0xAB, 0xCD, 0x12, 0x34}, nil
			},
			expectError:    false,
			expertWarns:    true,
			memoryContains: []byte{0xAB, 0xCD, 0x12},
			warnsContains:  "file entry image is larger than memory file:4 vs mem:3",
			memorySize:     3,
			useStrict:      false,
		},
		{
			name:  "failure - oversized file",
			tType: testFile,
			mockReadFile: func(string) ([]byte, error) {
				return []byte{0xAB, 0xCD, 0x12, 0x34}, nil
			},
			expectError:    true,
			expertWarns:    false,
			memoryContains: nil,
			errorContains:  "file entry image is larger than memory file:4 vs mem:3",
			memorySize:     3,
			useStrict:      true,
		},
		{
			name:           "success - same size",
			tType:          testArray,
			mockArray:      []byte{0xAB, 0xCD},
			expectError:    false,
			expertWarns:    false,
			memoryContains: []byte{0xAB, 0xCD},
			memorySize:     2,
			useStrict:      true,
		},
		{
			name:           "success - smaller size",
			tType:          testArray,
			mockArray:      []byte{0xAB, 0xCD},
			expectError:    false,
			expertWarns:    false,
			memoryContains: []byte{0xAB, 0xCD, 0x00, 0x00},
			memorySize:     4,
			useStrict:      true,
		},
		{
			name:           "warn - out of bounds",
			tType:          testArray,
			mockArray:      []byte{0xAB, 0xCD, 0x12, 0x34},
			expectError:    false,
			expertWarns:    true,
			memoryContains: []byte{0xAB, 0xCD, 0x12},
			warnsContains:  "array entry larger than size",
			memorySize:     3,
			useStrict:      false,
		},
		{
			name:           "failure - out of bounds",
			tType:          testArray,
			mockArray:      []byte{0xAB, 0xCD, 0x12, 0x34},
			expectError:    true,
			expertWarns:    false,
			memoryContains: nil,
			warnsContains:  "array entry larger than size",
			memorySize:     3,
			useStrict:      true,
		},
		{
			name:           "warn - size mismatch",
			tType:          testArray,
			mockArray:      []byte{0xAB, 0xCD},
			expectError:    false,
			expertWarns:    true,
			memoryContains: []byte{0xAB, 0xCD, 0x00, 0x00},
			warnsContains:  "array size larger than memory",
			memorySize:     4,
			imageSize:      6,
			useStrict:      false,
		},
		{
			name:           "failure - size mismatch",
			tType:          testArray,
			mockArray:      []byte{0xAB, 0xCD},
			expectError:    true,
			expertWarns:    false,
			memoryContains: nil,
			errorContains:  "array size larger than memory",
			memorySize:     4,
			imageSize:      6,
			useStrict:      true,
		},
		{
			name:           "failure - zero size for non-strict",
			tType:          testArray,
			mockArray:      []byte{0xAB, 0xCD},
			expectError:    true, // This is a specific case where non-strict still fails
			expertWarns:    false,
			memoryContains: nil,
			errorContains:  "array type requires size",
			memorySize:     4,
			imageSize:      0,
			forceImageSize: true,
			useStrict:      false,
		},
		{
			name:           "failure - zero size for strict",
			tType:          testArray,
			mockArray:      []byte{0xAB, 0xCD},
			expectError:    true,
			expertWarns:    false,
			memoryContains: nil,
			errorContains:  "array type requires size",
			memorySize:     4,
			imageSize:      0,
			forceImageSize: true,
			useStrict:      true,
		},
		{
			name:              "success - normal size",
			tType:             testEmpty,
			prepopulateMemory: []byte{0xCA, 0xFE, 0xD0, 0x0D},
			expectError:       false,
			expertWarns:       false,
			memoryContains:    []byte{0x00, 0x00, 0x00, 0x00},
			imageSize:         4,
			useStrict:         true,
		},
		{
			name:              "success - smaller size",
			tType:             testEmpty,
			prepopulateMemory: []byte{0xCA, 0xFE, 0xD0, 0x0D},
			expectError:       false,
			expertWarns:       false,
			memoryContains:    []byte{0x00, 0x00, 0xD0, 0x0D},
			imageSize:         2,
			useStrict:         true,
		},
		{
			name:              "warn - larger size",
			tType:             testEmpty,
			prepopulateMemory: []byte{0xCA, 0xFE, 0xD0, 0x0D},
			expectError:       false,
			expertWarns:       true,
			memoryContains:    []byte{0x00, 0x00, 0x00, 0x00},
			warnsContains:     "memory is smaller than image size",
			imageSize:         6,
			useStrict:         false,
		},
		{
			name:              "failure - larger size",
			tType:             testEmpty,
			prepopulateMemory: []byte{0xCA, 0xFE, 0xD0, 0x0D},
			expectError:       true,
			expertWarns:       false,
			memoryContains:    nil,
			errorContains:     "memory is smaller than image size",
			imageSize:         6,
			useStrict:         true,
		},
	}

	for i := range tests {
		tc := tests[i]
		name := tc.tType.String() + ": " + tc.name
		var cfg *wasmvm.ImageConfig = nil
		var replaceReadFile readFileMock = nil
		size := uint64(tc.memorySize)
		/*
			The lack of a ternary operator actually annoys me about the language
			I almost made an Excel style if
			func If[T any](someCondition bool, ifTrue, ifFalse T) T {
				if someCondition {
					return ifTrue
				}
				return ifFalse
			}

			and then

			size := uint64(If(tc.imageSize > 0, tc.imageSize, tc.memorySize))

			But that's probably not proper Golang either... Well, assuming
			that generics can even be used like that
		*/

		if tc.forceImageSize || tc.imageSize != 0 {
			size = uint64(tc.imageSize)
		}
		t.Run(name, func(t *testing.T) {
			switch tc.tType {
			case testFile:
				cfg = &wasmvm.ImageConfig{
					Type:     "file",
					Filename: "fake.file",
				}
				replaceReadFile = tc.mockReadFile
			case testArray:
				cfg = &wasmvm.ImageConfig{
					Type:  "array",
					Size:  size,
					Array: tc.mockArray,
				}
			case testEmpty:
				cfg = &wasmvm.ImageConfig{
					Type: "empty",
					Size: size,
				}
			}

			// idea borrowed technically from JavaScript
			// Self-executing function/closure. In this case
			// defer triggers when a function ends, so I am
			// bounding the scope with the embedded function
			// in order to ensure that defer executes inside
			// the loop correctly for cleanup
			// Or...
			// Just to say the "fightin' words" ...
			// "glorified try/finally block" [or try-with-resources or using(){}]
			// Better than saying "RAII, can I haz it?"; Upyo~~! I just did!
			func() {

				if replaceReadFile != nil {
					original := wasmvm.ReadFile
					defer func() { wasmvm.ReadFile = original }()
					wasmvm.ReadFile = replaceReadFile
				}

				var mem []byte

				if tc.prepopulateMemory != nil {
					mem = tc.prepopulateMemory
				} else {
					mem = make([]byte, tc.memorySize)
				}

				warns, err := wasmvm.PopulateImage(mem, cfg, tc.useStrict)

				if tc.expectError {
					assert.Error(t, err)
					assert.Empty(t, warns)
					if tc.checkErrorType {
						require.NotEmpty(t, tc.errorType, "Test configuration error, errorType must not be nil")
						require.ErrorAs(t, err, &tc.errorType)
						assert.Equal(t, tc.errorType.Type, GetImageErrorType(err))
						if tc.checkErrorCause {
							assert.ErrorAs(t, tc.errorType.Cause, tc.errorType.Unwrap().Error())
						}

					} else if tc.errorContains != "" {
						assert.Contains(t, err.Error(), tc.errorContains)
					}

				} else {
					assert.NoError(t, err)
					if tc.expertWarns {
						assert.NotEmpty(t, warns)
						if tc.warnsContains != "" {
							assert.Contains(t, warns[0], tc.warnsContains)
						}
					}
					if tc.memoryContains != nil {
						assert.Equal(t, tc.memoryContains, mem)
					} else {
						assert.Empty(t, mem)
					}

				}
			}()
		})
	}
}

func TestPopulateImage_EmptyType(t *testing.T) {
	mem := []byte{99, 88, 77}
	cfg := &wasmvm.ImageConfig{
		Type: "empty",
		Size: 3,
	}
	warns, err := wasmvm.PopulateImage(mem, cfg, false)
	assert.NoError(t, err)
	assert.Empty(t, warns)
	assert.Equal(t, []byte{0, 0, 0}, mem)
}

func TestPopulateImage_EmptyType_SizeZero(t *testing.T) {
	mem := make([]byte, 4)
	cfg := &wasmvm.ImageConfig{
		Type: "empty",
		Size: 0,
	}

	_, err := wasmvm.PopulateImage(mem, cfg, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty type requires size")
}

func TestPopulateImage_SparseArrayType_Normal(t *testing.T) {
	mem := make([]byte, 10) // memory with 10 bytes, all zero by default
	cfg := &wasmvm.ImageConfig{
		Type: "sparsearray",
		Sparse: []wasmvm.SparseArrayEntry{
			{Offset: 0, Array: []uint8{1, 2, 3}}, // fills mem[0], mem[1], mem[2]
			{Offset: 7, Array: []uint8{8, 9}},    // fills mem[7], mem[8]
		},
	}
	warns, err := wasmvm.PopulateImage(mem, cfg, true)
	assert.NoError(t, err)
	assert.Empty(t, warns)
	assert.Equal(t, []byte{1, 2, 3, 0, 0, 0, 0, 8, 9, 0}, mem)
}

func TestPopulateImage_SparseArrayType_StrictAndLenient_OOB(t *testing.T) {
	// Test strict
	mem := make([]byte, 2)
	cfg := &wasmvm.ImageConfig{
		Type: "sparsearray",
		Sparse: []wasmvm.SparseArrayEntry{
			{Offset: 0, Array: []uint8{7}},
			{Offset: 2, Array: []uint8{8}}, // out of bounds
		},
	}
	warns, err := wasmvm.PopulateImage(mem, cfg, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sparsearray entry out of bounds")

	// Test lenient
	mem = make([]byte, 2)
	warns, err = wasmvm.PopulateImage(mem, cfg, false)
	assert.NoError(t, err)
	assert.Contains(t, warns[0], "sparsearray entry out of bounds")
	assert.Equal(t, uint8(7), mem[0])
}

func TestPopulateImage_OverwriteDetection(t *testing.T) {
	// Overwrite warning, lenient
	mem := []byte{5, 0}
	cfg := &wasmvm.ImageConfig{
		Type: "sparsearray",
		Sparse: []wasmvm.SparseArrayEntry{
			{Offset: 0, Array: []uint8{6}},
		},
	}
	warns, err := wasmvm.PopulateImage(mem, cfg, false)
	assert.NoError(t, err)
	assert.Contains(t, warns[0], "overwrite at offset 0")

	// Overwrite error, strict
	mem = []byte{5, 0}
	warns, err = wasmvm.PopulateImage(mem, cfg, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "overwrite at offset 0")
}

func TestPopulateImage_UnknownType(t *testing.T) {
	mem := make([]byte, 1)
	cfg := &wasmvm.ImageConfig{
		Type: "foobar",
	}
	warns, err := wasmvm.PopulateImage(mem, cfg, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown image type")
	assert.Empty(t, warns)
}

func TestParseImageConfig_JSON(t *testing.T) {
	raw := []byte(`{"type":"array", "array":[1,2,3], "size":4}`)
	cfg, err := wasmvm.ParseImageConfig(raw)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "array", cfg.Type)
	assert.Equal(t, []uint8{1, 2, 3}, cfg.Array)
	assert.Equal(t, uint64(4), cfg.Size)
}

func TestParseImageConfig_JSON_BogusInput(t *testing.T) {
	raw := []byte(`<image><type>array</type><array><item>1</item><item>2</item><item>3</item><size>4</size></image>`) // I know it could have been just jibberish, but why not?
	cfg, err := wasmvm.ParseImageConfig(raw)
	require.Error(t, err)
	require.Nil(t, cfg)
}
