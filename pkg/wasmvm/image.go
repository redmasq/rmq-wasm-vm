package wasmvm

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	//"golang.org/x/text/cases"
	//"golang.org/x/text/language"
)

//go:generate stringer -type=ImageType
type ImageType byte

const (
	Unknown ImageType = iota
	File
	Array
	Empty
	SparseArray
)

var imageTypeStringToEnum = map[string]ImageType{
	"unknown":     Unknown,
	"file":        File,
	"array":       Array,
	"empty":       Empty,
	"sparsearray": SparseArray,
}

func ParseImageType(s string) (ImageType, error) {
	key := strings.ToLower(strings.TrimSpace(s))
	if val, ok := imageTypeStringToEnum[key]; ok {
		return val, nil
	}
	return Unknown, NewImageInitializationError(UnknownImageType, fmt.Sprintf("unknown image type: %q", s))
}

func (t *ImageType) UnmarshalText(text []byte) error {
	val, err := ParseImageType(string(text))
	if err != nil {
		return err
	}
	*t = val
	return nil
}

func (t *ImageType) UnmarshalJSON(data []byte) error {
	// Try as string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		return t.UnmarshalText([]byte(s))
	}
	// Try as number
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*t = ImageType(i)
		return nil
	}
	return fmt.Errorf("ImageType: cannot unmarshal %s", string(data))
}

//go:generate stringer -type=ImageInitializationErrorType
type ImageInitializationErrorType byte

const (
	UndefinedImageError ImageInitializationErrorType = iota
	UnknownImageType
	FileImageOtherError
	ImageSizeRequired
	ImageSizeTooLargeForConfig
	ImageSizeTooLargeForMemory
	ImageInitArrayLargerThanConfig
	SparseEntryOutOfBounds
	SparseEntryMemoryOverwrite
	SparseEntryMultipleTypes
)

// Custom error struct
type ImageInitializationError struct {
	Msg   string                       `json:"message"`
	Type  ImageInitializationErrorType `json:"type"`
	Cause error                        `json:"cause,omitempty"`
	Meta  any                          `json:"metadata,omitempty"`
}

// Implement the `error` interface
func (e *ImageInitializationError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type.String(), e.Msg)
}

// Another from the `error` interface
func (e *ImageInitializationError) Unwrap() error {
	return e.Cause
}

/*
func (e *ImageInitializationError) ApplyCause(err error) *ImageInitializationError {
	e.Cause = err
	return e
}

*/

func (e *ImageInitializationError) ApplyMeta(meta any) *ImageInitializationError {
	e.Meta = meta
	return e
}

func (sa *SparseArrayEntry) CreateErrorMeta(errorType ImageInitializationErrorType) *SparseArrayErrorEntry {
	return &SparseArrayErrorEntry{
		Offset:    sa.Offset,
		Array:     sa.Array,
		ErrorType: errorType,
	}
}

/*
// This might get added later to support configuration export
func (e *ImageInitializationError) MarshalJSON() ([]byte, error) {
	type Dummy ImageInitializationError
	return json.Marshal(&struct {
		Type string `json:"type"` // override only this
		*Dummy
	}{
		Type:  e.Type.String(),
		Dummy: (*Dummy)(e), // ← actually using the alias
	})
}
*/

// Constructor helper
func NewImageInitializationError(t ImageInitializationErrorType, msg string) error {
	return &ImageInitializationError{
		Type: t,
		Msg:  msg,
	}
}

func NewImageInitializationErrorWithCause(t ImageInitializationErrorType, msg string, cause error) error {
	return &ImageInitializationError{
		Type:  t,
		Msg:   msg,
		Cause: cause,
	}
}

// Defined so as to allow for mocking
var ReadFile = os.ReadFile

type ImageConfig struct {
	Type     ImageType          `json:"type"`
	Filename string             `json:"filename,omitempty"`
	Array    []uint8            `json:"array,omitempty"`
	Size     uint64             `json:"size,omitempty"`
	Sparse   []SparseArrayEntry `json:"sparsearray,omitempty"`
}

func (ic *ImageConfig) SetType(it ImageType) *ImageConfig {
	ic.Type = it
	return ic
}

func (ic *ImageConfig) SetFilename(filename string) *ImageConfig {
	ic.Type = File
	ic.Filename = filename
	return ic
}

func (ic *ImageConfig) SetArray(arr []uint8) *ImageConfig {
	ic.Type = Array
	ic.Array = arr
	return ic
}

func (ic *ImageConfig) SetSize(size uint64) *ImageConfig {
	ic.Size = size
	return ic
}

func (ic *ImageConfig) SetSparseArray(sa []SparseArrayEntry) *ImageConfig {
	ic.Type = SparseArray
	ic.Sparse = sa
	return ic
}

type SparseArrayEntry struct {
	Offset uint64  `json:"offset"`
	Array  []uint8 `json:"array"`
}

type SparseArrayErrorEntry struct {
	Offset    uint64  `json:"offset"`
	Array     []uint8 `json:"array"`
	ErrorType ImageInitializationErrorType
}

type ImageErrorMetaData struct {
	Filename   string `json:"filename,omitempty"`
	DataSize   uint64 `json:"dataSize"`
	ConfigSize uint64 `json:"configSize"`
	MemSize    uint64 `json:"memSize"`
}

type ImageErrorSparseMetaData struct {
	ConfigSize     uint64                  `json:"configSize"`
	MemSize        uint64                  `json:"memSize"`
	ProblemEntries []SparseArrayErrorEntry `json:"problemEntries"`
}

const errmsg_ReadingFile = "Error while reading image file"
const errmsg_FileLargerMemory = "file entry image is larger than memory file:%d vs mem:%d"
const errmsg_ArrayRequiresSize = "array type requires size"
const errmsg_ArrayLargerMemory = "array configured size larger than memory"
const errmsg_ArrayLargerSize = "array entry larger than size"
const errmsg_EmptyRequireSize = "empty type requires size"
const errmsg_EmptyMemorySmallerThanSize = "memory is smaller than image size"
const errmsg_SparseArrayOOBNumbered = "sparsearray entry out of bounds at offset %d"
const errmsg_SparseArrayOOB = "sparsearray entry out of bounds detected"
const errmsg_SparseArrayOverwriteNumbered = "sparsearray: overwrite at offset %d"
const errmsg_SparseArrayOverwrite = "sparsearray: overwrite detected"
const errmsg_SparseArrayMultiple = "sparsearray: multiple errors"
const errmsg_SpareArrayUnknown = "sparearray: unknown error"

// PopulateImage fills mem according to config; returns warnings and error if any
func PopulateImage(mem []byte, cfg *ImageConfig, strict bool) ([]string, error) {
	warns := []string{}
	switch cfg.Type {
	case File:
		warns, err := handleFile(cfg, warns, mem, strict)
		return warns, err
	case Array:
		warns, err := handleArray(cfg, mem, warns, strict)
		return warns, err
	case Empty:
		warns, err := handleEmpty(cfg, mem, warns, strict)
		return warns, err
	case SparseArray:
		warns, err := handleSparse(cfg, mem, strict, warns)
		return warns, err
	default:
		return warns, NewImageInitializationError(UnknownImageType, fmt.Sprintf("unknown image type: %s", cfg.Type.String()))
	}
}

func createSparseProblemKey(et ImageInitializationErrorType, i int) string {
	return fmt.Sprintf("%d:%d", byte(et), i)
}

func handleSparse(cfg *ImageConfig, mem []byte, strict bool, warns []string) ([]string, error) {
	cache := make(map[string]struct{})
	problemEntries := []SparseArrayErrorEntry{}
	eType := UndefinedImageError
	for j, entry := range cfg.Sparse {
		for i, b := range entry.Array {
			addr := entry.Offset + uint64(i)
			if addr >= uint64(len(mem)) {
				if strict {
					if eType == UndefinedImageError {
						eType = SparseEntryOutOfBounds
					} else if eType == SparseEntryOutOfBounds {
						// Do Nothing
					} else if eType != SparseEntryMultipleTypes {
						eType = SparseEntryMultipleTypes
					}
					// Distinct check for entry
					key := createSparseProblemKey(SparseEntryOutOfBounds, j)
					if _, ok := cache[key]; !ok {
						cache[key] = struct{}{}
						em := *entry.CreateErrorMeta(SparseEntryOutOfBounds)
						problemEntries = append(problemEntries, em)
					}

					continue
				} else {
					warns = append(warns, fmt.Sprintf(errmsg_SparseArrayOOBNumbered, addr))
				}
			} else if mem[addr] != 0x00 && !strict {
				// Note that Overwrite means replacing non-zero data
				// rather than than a range check
				warns = append(warns, fmt.Sprintf(errmsg_SparseArrayOverwriteNumbered, addr))
			} else if mem[addr] != 0x00 && strict {
				if eType == UndefinedImageError {
					eType = SparseEntryMemoryOverwrite
				} else if eType == SparseEntryMemoryOverwrite {
					// Do Nothing
				} else if eType != SparseEntryMultipleTypes {
					eType = SparseEntryMultipleTypes
				}
				key := createSparseProblemKey(SparseEntryOutOfBounds, j)
				if _, ok := cache[key]; !ok {
					cache[key] = struct{}{}
					em := *entry.CreateErrorMeta(SparseEntryMemoryOverwrite)
					problemEntries = append(problemEntries, em)
				}
				continue
			}
			if addr < uint64(len(mem)) {
				mem[addr] = b
			}
		}
	}
	// We don't abort early
	if len(problemEntries) != 0 {
		var msg string
		switch eType {
		case SparseEntryOutOfBounds:
			msg = errmsg_SparseArrayOOB
		case SparseEntryMemoryOverwrite:
			msg = errmsg_SparseArrayOverwrite
		case SparseEntryMultipleTypes:
			msg = errmsg_SparseArrayMultiple
		default:
			// There shouldn't be any condition that actually triggers this
			// But here for future ease of identification of issues
			msg = errmsg_SpareArrayUnknown
		}
		ferr := NewImageInitializationError(eType, msg)
		if bldErr, ok := ferr.(*ImageInitializationError); ok {
			bldErr.ApplyMeta(ImageErrorSparseMetaData{
				ConfigSize:     uint64(cfg.Size),
				MemSize:        uint64(len(mem)),
				ProblemEntries: problemEntries,
			})
		}
		return nil, ferr
	}
	return warns, nil
}

func handleEmpty(cfg *ImageConfig, mem []byte, warns []string, strict bool) ([]string, error) {
	if cfg.Size == 0 {
		ferr := NewImageInitializationError(ImageSizeRequired, errmsg_EmptyRequireSize)
		if bldErr, ok := ferr.(*ImageInitializationError); ok {
			bldErr.ApplyMeta(ImageErrorMetaData{
				Filename:   cfg.Filename,
				DataSize:   uint64(len(cfg.Array)),
				ConfigSize: uint64(cfg.Size),
				MemSize:    uint64(len(mem)),
			})
		}
		return warns, ferr
	}
	if cfg.Size > uint64(len(mem)) {
		if strict {
			ferr := NewImageInitializationError(ImageSizeTooLargeForMemory, errmsg_EmptyMemorySmallerThanSize)
			if bldErr, ok := ferr.(*ImageInitializationError); ok {
				bldErr.ApplyMeta(ImageErrorMetaData{
					Filename:   cfg.Filename,
					DataSize:   uint64(len(cfg.Array)),
					ConfigSize: uint64(cfg.Size),
					MemSize:    uint64(len(mem)),
				})
			}
			return warns, ferr
		}
		warns = append(warns, errmsg_EmptyMemorySmallerThanSize)
	}
	for i := uint64(0); i < cfg.Size && i < uint64(len(mem)); i++ {
		mem[i] = 0x00
	}
	return warns, nil
}

func handleArray(cfg *ImageConfig, mem []byte, warns []string, strict bool) ([]string, error) {
	if cfg.Size == 0 {
		ferr := NewImageInitializationError(ImageSizeRequired, errmsg_ArrayRequiresSize)
		if bldErr, ok := ferr.(*ImageInitializationError); ok {
			bldErr.ApplyMeta(ImageErrorMetaData{
				Filename:   cfg.Filename,
				DataSize:   uint64(len(cfg.Array)),
				ConfigSize: uint64(cfg.Size),
				MemSize:    uint64(len(mem)),
			})
		}
		return warns, ferr
	}
	if cfg.Size > uint64(len(mem)) {
		if strict {
			ferr := NewImageInitializationError(ImageSizeTooLargeForMemory, errmsg_ArrayLargerMemory)
			if bldErr, ok := ferr.(*ImageInitializationError); ok {
				bldErr.ApplyMeta(ImageErrorMetaData{
					Filename:   cfg.Filename,
					DataSize:   uint64(len(cfg.Array)),
					ConfigSize: uint64(cfg.Size),
					MemSize:    uint64(len(mem)),
				})
			}
			return warns, ferr
		}
		warns = append(warns, errmsg_ArrayLargerMemory)
	}
	if cfg.Size < uint64(len(cfg.Array)) {
		if strict {
			ferr := NewImageInitializationError(ImageInitArrayLargerThanConfig, errmsg_ArrayLargerSize)
			if bldErr, ok := ferr.(*ImageInitializationError); ok {
				bldErr.ApplyMeta(ImageErrorMetaData{
					Filename:   cfg.Filename,
					DataSize:   uint64(len(cfg.Array)),
					ConfigSize: uint64(cfg.Size),
					MemSize:    uint64(len(mem)),
				})
			}
			return warns, ferr
		}
		warns = append(warns, errmsg_ArrayLargerSize)
	}
	copy(mem, cfg.Array)
	for i := uint64(len(cfg.Array)); i < cfg.Size && i < uint64(len(mem)); i++ {
		mem[i] = 0x00
	}
	return warns, nil
}

func handleFile(cfg *ImageConfig, warns []string, mem []byte, strict bool) ([]string, error) {
	data, err := ReadFile(cfg.Filename)
	if err != nil {
		return warns, NewImageInitializationErrorWithCause(FileImageOtherError, errmsg_ReadingFile, err)
	}
	if len(data) > len(mem) {
		if strict {
			ferr := NewImageInitializationError(ImageSizeTooLargeForMemory,
				fmt.Sprintf(errmsg_FileLargerMemory, len(data), len(mem)))
			if bldErr, ok := ferr.(*ImageInitializationError); ok {
				bldErr.ApplyMeta(ImageErrorMetaData{
					Filename:   cfg.Filename,
					DataSize:   uint64(len(data)),
					ConfigSize: uint64(cfg.Size),
					MemSize:    uint64(len(mem)),
				})
			}

			return warns, ferr
		}
		// For now, we are keeping warns as just strings
		warns = append(warns, fmt.Sprintf(errmsg_FileLargerMemory, len(data), len(mem)))
	}
	copy(mem, data)
	return warns, nil
}

// ParseImageConfig parses JSON and returns *ImageConfig
func ParseImageConfig(jsonBytes []byte) (*ImageConfig, error) {
	var cfg ImageConfig
	if err := json.Unmarshal(jsonBytes, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
