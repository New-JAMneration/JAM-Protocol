// Package deblob provides utilities for extracting and writing PVM program
// static metadata (InstrMeta, BlockMeta) from program blobs.
package deblob

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// InstrMetaJSON is the JSON representation of a single instruction's metadata.
type InstrMetaJSON struct {
	Index       int      `json:"index"`
	PC          uint32   `json:"pc"`
	Opcode      byte     `json:"opcode"`
	OpcodeName  string   `json:"opcode_name"`
	SkipLen     uint8    `json:"skip_len"`
	Dst         uint8    `json:"dst"`
	Src         [2]uint8 `json:"src"`
	Imm         [2]uint64 `json:"imm"`
	BlockStartPC uint32  `json:"block_start_pc"`
}

// BlockMetaJSON is the JSON representation of a basic block.
type BlockMetaJSON struct {
	StartPC    uint32 `json:"start_pc"`
	EndPC      uint32 `json:"end_pc"`
	InstrStart int    `json:"instr_start"`
	InstrEnd   int    `json:"instr_end"`
	GasCost    int64  `json:"gas_cost"`
}

// ProgramMetadata holds the complete deblob output.
type ProgramMetadata struct {
	FormatVersion    int             `json:"format_version"`
	CodeHash         string          `json:"codehash"`
	InstructionCount int             `json:"instruction_count"`
	BlockCount       int             `json:"block_count"`
	Instructions     []InstrMetaJSON `json:"instructions"`
	Blocks           []BlockMetaJSON `json:"blocks"`
}

// WriteMetadata writes program metadata to the output directory.
func WriteMetadata(outDir string, meta *ProgramMetadata) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	instrPath := filepath.Join(outDir, "instr_meta.json.gz")
	if err := writeGzipJSON(instrPath, struct {
		FormatVersion    int             `json:"format_version"`
		CodeHash         string          `json:"codehash"`
		InstructionCount int             `json:"instruction_count"`
		Instructions     []InstrMetaJSON `json:"instructions"`
	}{
		FormatVersion:    meta.FormatVersion,
		CodeHash:         meta.CodeHash,
		InstructionCount: meta.InstructionCount,
		Instructions:     meta.Instructions,
	}); err != nil {
		return fmt.Errorf("write instr_meta.json.gz: %w", err)
	}

	blocksPath := filepath.Join(outDir, "blocks.json.gz")
	if err := writeGzipJSON(blocksPath, struct {
		FormatVersion int             `json:"format_version"`
		CodeHash      string          `json:"codehash"`
		BlockCount    int             `json:"block_count"`
		Blocks        []BlockMetaJSON `json:"blocks"`
	}{
		FormatVersion: meta.FormatVersion,
		CodeHash:      meta.CodeHash,
		BlockCount:    meta.BlockCount,
		Blocks:        meta.Blocks,
	}); err != nil {
		return fmt.Errorf("write blocks.json.gz: %w", err)
	}

	return nil
}

func writeGzipJSON(path string, v interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	enc := json.NewEncoder(gz)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		gz.Close()
		return err
	}
	return gz.Close()
}
