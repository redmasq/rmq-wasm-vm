package wasmvm

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type ImageConfig struct {
	Type     string             `json:"type"`
	Filename string             `json:"filename,omitempty"`
	Array    []uint8            `json:"array,omitempty"`
	Size     uint64             `json:"size,omitempty"`
	Sparse   []SparseArrayEntry `json:"sparsearray,omitempty"`
}

type SparseArrayEntry struct {
	Offset uint64  `json:"offset"`
	Array  []uint8 `json:"array"`
}

// PopulateImage fills mem according to config; returns warnings and error if any
func PopulateImage(mem []byte, cfg *ImageConfig, strict bool) ([]string, error) {
	warns := []string{}
	switch cfg.Type {
	case "file":
		data, err := os.ReadFile(cfg.Filename)
		if err != nil {
			return warns, err
		}
		copy(mem, data)
	case "array":
		if cfg.Size == 0 {
			return warns, errors.New("array type requires size")
		}
		copy(mem, cfg.Array)
		for i := uint64(len(cfg.Array)); i < cfg.Size && i < uint64(len(mem)); i++ {
			mem[i] = 0x00
		}
	case "empty":
		if cfg.Size == 0 {
			return warns, errors.New("empty type requires size")
		}
		for i := uint64(0); i < cfg.Size && i < uint64(len(mem)); i++ {
			mem[i] = 0x00
		}
	case "sparsearray":
		for _, entry := range cfg.Sparse {
			for i, b := range entry.Array {
				addr := entry.Offset + uint64(i)
				if addr >= uint64(len(mem)) {
					if strict {
						return warns, fmt.Errorf("sparsearray entry out of bounds at offset %d", addr)
					}
					warns = append(warns, fmt.Sprintf("sparsearray entry out of bounds at offset %d", addr))
					continue
				}
				if mem[addr] != 0x00 && !strict {
					warns = append(warns, fmt.Sprintf("sparsearray: overwrite at offset %d", addr))
				} else if mem[addr] != 0x00 && strict {
					return warns, fmt.Errorf("sparsearray: overwrite at offset %d", addr)
				}
				mem[addr] = b
			}
		}
	default:
		return warns, fmt.Errorf("unknown image type: %s", cfg.Type)
	}
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
