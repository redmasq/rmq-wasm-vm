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

// For the struct to define the function signature
type readFileMock func(string) ([]byte, error)

// enum for the test type
type testType int8

// And we autonumber it
const (
	testFile testType = iota
	testArray
	testSparse
	testEmpty
	testOther
)

// This will for any of the defined const above
// be the source for a splice
const _testType_name = "testFiletestArraytestSparsetestEmptytestOther"

// map of boundary locations
var _testType_index = [...]uint8{0, 8, 17, 27, 36, 45}

// Basically, what's the highest index on the map
const _testType_indexlimit = testType(len(_testType_index) - 1)

// prints for defined values just the name from
// the string above, spliced out, otherwise, the
// type name with the integer value
// mimicked from the go generate since I didn't want
// to generate for a private type
func (i testType) String() string {
	if i < 0 || i >= _testType_indexlimit {
		return "testType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _testType_name[_testType_index[i]:_testType_index[i+1]] // Splice
}

// Helper functions for getting values
func GetImageErrorType(err error) wasmvm.ImageInitializationErrorType {
	var imgErr *wasmvm.ImageInitializationError
	if errors.As(err, &imgErr) {
		return imgErr.Type
	}
	return wasmvm.UndefinedImageError
}

func GetImageErrorMessage(err error) string {
	var imgErr *wasmvm.ImageInitializationError
	if errors.As(err, &imgErr) {
		return imgErr.Msg
	}
	return err.Error()
}

func GetImageErrorMeta(err error) any {
	var imgErr *wasmvm.ImageInitializationError
	if errors.As(err, &imgErr) {
		return imgErr.Meta
	}
	return nil
}

// Configuration struct itself
type imageTestCase struct {
	name              string
	tType             testType
	mockReadFile      readFileMock
	mockArray         []byte
	expectError       bool
	expertWarns       bool
	checkErrorType    bool
	checkErrorMessage bool
	checkErrorCause   bool
	checkCauseString  bool
	checkMeta         bool
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

func executeTests(t *testing.T, tests []imageTestCase) {
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
			cfg, replaceReadFile = prepareConfigForTest(tc, cfg, replaceReadFile, size)

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
					processTestExpectError(t, err, warns, tc)

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

// Extracted function for handling the error condition
func processTestExpectError(t *testing.T, err error, warns []string, tc imageTestCase) {
	assert.Error(t, err)
	assert.Empty(t, warns)
	if tc.checkErrorType {
		require.NotEmpty(t, tc.errorType, "Test configuration error, errorType must not be nil")
		require.IsType(t, err, tc.errorType) // Not checking the chain since we are breaking into steps

		assert.Equal(t, tc.errorType.Type, GetImageErrorType(err))
		if tc.checkErrorMessage {
			assert.Contains(t, GetImageErrorMessage(err), GetImageErrorMessage(tc.errorType))
		}
		if tc.checkErrorCause {
			err2 := errors.Unwrap(err)
			errCmp := errors.Unwrap(tc.errorType)
			assert.IsType(t, errCmp, err2)
			if tc.checkCauseString {
				assert.Contains(t, err2.Error(), errCmp.Error())
			}
		}
		if tc.checkMeta {
			errMeta := GetImageErrorMeta(err)
			errCmpMeta := GetImageErrorMeta(tc.errorType)
			assert.True(t, (errMeta != nil && errCmpMeta != nil) || errMeta == nil && errCmpMeta == nil)
			if errMeta != nil {
				assert.IsType(t, errCmpMeta, errMeta)
				assert.Equal(t, errCmpMeta, errMeta)
			}

		}

	} else if tc.errorContains != "" {
		assert.Contains(t, err.Error(), tc.errorContains)
	}
}

func prepareConfigForTest(tc imageTestCase, cfg *wasmvm.ImageConfig, replaceReadFile readFileMock, size uint64) (*wasmvm.ImageConfig, readFileMock) {
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
	return cfg, replaceReadFile
}

func TestPopulateImage_File(t *testing.T) {
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
			expectError:       true,
			expertWarns:       false,
			memoryContains:    nil,
			checkErrorType:    true,
			checkErrorMessage: true,
			checkErrorCause:   true,
			checkCauseString:  true,
			errorType: &wasmvm.ImageInitializationError{
				Type:  wasmvm.FileImageOtherError,
				Msg:   "Error while reading image file",
				Cause: fmt.Errorf("I/O Error because \"reasons\""),
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
			expectError:       true,
			expertWarns:       false,
			memoryContains:    nil,
			checkErrorType:    true,
			checkErrorMessage: true,
			checkMeta:         true,
			errorType: &wasmvm.ImageInitializationError{
				Type: wasmvm.ImageSizeTooLargeForMemory,
				Msg:  "file entry image is larger than memory file:4 vs mem:3",
				Meta: wasmvm.ImageErrorMetaData{
					Filename: "fake.file",
					DataSize: uint64(4),
					MemSize:  uint64(3),
				},
			},
			memorySize: 3,
			useStrict:  true,
		},
	}
	executeTests(t, tests)
}

func TestPopulateImage_Array(t *testing.T) {
	tests := []imageTestCase{
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
			name:              "failure - out of bounds",
			tType:             testArray,
			mockArray:         []byte{0xAB, 0xCD, 0x12, 0x34},
			expectError:       true,
			expertWarns:       false,
			memoryContains:    nil,
			checkErrorType:    true,
			checkErrorMessage: true,
			checkMeta:         true,
			errorType: &wasmvm.ImageInitializationError{
				Type: wasmvm.ImageInitArrayLargerThanConfig,
				Msg:  "array entry larger than size",
				Meta: wasmvm.ImageErrorMetaData{
					DataSize:   uint64(4),
					ConfigSize: uint64(3),
					MemSize:    uint64(3),
				},
			},
			memorySize: 3,
			useStrict:  true,
		},
		{
			name:           "warn - size mismatch",
			tType:          testArray,
			mockArray:      []byte{0xAB, 0xCD},
			expectError:    false,
			expertWarns:    true,
			memoryContains: []byte{0xAB, 0xCD, 0x00, 0x00},
			warnsContains:  "array configured size larger than memory",
			memorySize:     4,
			imageSize:      6,
			useStrict:      false,
		},
		{
			name:              "failure - size mismatch",
			tType:             testArray,
			mockArray:         []byte{0xAB, 0xCD},
			expectError:       true,
			expertWarns:       false,
			memoryContains:    nil,
			checkErrorType:    true,
			checkErrorMessage: true,
			checkMeta:         true,
			errorType: &wasmvm.ImageInitializationError{
				Type: wasmvm.ImageSizeTooLargeForMemory,
				Msg:  "array configured size larger than memory",
				Meta: wasmvm.ImageErrorMetaData{
					DataSize:   uint64(2),
					ConfigSize: uint64(6),
					MemSize:    uint64(4),
				},
			},
			memorySize: 4,
			imageSize:  6,
			useStrict:  true,
		},
		{
			name:              "failure - zero size for non-strict",
			tType:             testArray,
			mockArray:         []byte{0xAB, 0xCD},
			expectError:       true,
			expertWarns:       false,
			memoryContains:    nil,
			checkErrorType:    true,
			checkErrorMessage: true,
			checkMeta:         true,
			errorType: &wasmvm.ImageInitializationError{
				Type: wasmvm.ImageSizeRequired,
				Msg:  "array type requires size",
				Meta: wasmvm.ImageErrorMetaData{
					DataSize:   uint64(2),
					ConfigSize: uint64(0),
					MemSize:    uint64(4),
				},
			},
			memorySize:     4,
			imageSize:      0,
			forceImageSize: true,
			useStrict:      false,
		},
		{
			name:              "failure - zero size for strict",
			tType:             testArray,
			mockArray:         []byte{0xAB, 0xCD},
			expectError:       true,
			expertWarns:       false,
			memoryContains:    nil,
			checkErrorType:    true,
			checkErrorMessage: true,
			checkMeta:         true,
			errorType: &wasmvm.ImageInitializationError{
				Type: wasmvm.ImageSizeRequired,
				Msg:  "array type requires size",
				Meta: wasmvm.ImageErrorMetaData{
					DataSize:   uint64(2),
					ConfigSize: uint64(0),
					MemSize:    uint64(4),
				},
			},
			memorySize:     4,
			imageSize:      0,
			forceImageSize: true,
			useStrict:      true,
		},
	}
	executeTests(t, tests)
}

func TestPopulateImage_Empty(t *testing.T) {
	tests := []imageTestCase{
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
			checkErrorType:    true,
			checkErrorMessage: true,
			checkMeta:         true,
			errorType: &wasmvm.ImageInitializationError{
				Type: wasmvm.ImageSizeTooLargeForMemory,
				Msg:  "memory is smaller than image size",
				Meta: wasmvm.ImageErrorMetaData{
					DataSize:   uint64(0),
					ConfigSize: uint64(6),
					MemSize:    uint64(4),
				},
			},
			imageSize: 6,
			useStrict: true,
		},
		{
			name:              "failure - zero size",
			tType:             testEmpty,
			prepopulateMemory: []byte{0xCA, 0xFE, 0xD0, 0x0D},
			expectError:       true,
			expertWarns:       false,
			memoryContains:    nil,
			checkErrorType:    true,
			checkErrorMessage: true,
			checkMeta:         true,
			errorType: &wasmvm.ImageInitializationError{
				Type: wasmvm.ImageSizeRequired,
				Msg:  "empty type requires size",
				Meta: wasmvm.ImageErrorMetaData{
					DataSize:   uint64(0),
					ConfigSize: uint64(0),
					MemSize:    uint64(4),
				},
			},
			imageSize: 0,
			useStrict: true,
		},
	}

	executeTests(t, tests)
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
