package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/PVMtrace/deblob"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
)

func runDeblob(args []string) error {
	if len(args) < 2 || args[0] != "program" {
		return fmt.Errorf("usage: pvmtrace deblob program <target-file> <codehash>")
	}

	targetFile := args[1]
	codeHashArg := ""
	if len(args) >= 3 {
		codeHashArg = args[2]
	}

	outRoot := os.Getenv("PVM_DEBLOB_DIR")
	if outRoot == "" {
		outRoot = "./pvm-deblob"
	}

	ext := strings.ToLower(filepath.Ext(targetFile))
	switch ext {
	case ".json":
		return deblobFromJSON(targetFile, codeHashArg, outRoot)
	case ".bin":
		return deblobFromBin(targetFile, codeHashArg, outRoot)
	default:
		return fmt.Errorf("unsupported file extension %q (use .json or .bin)", ext)
	}
}

func deblobFromJSON(targetFile, codeHashArg, outRoot string) error {
	data, err := os.ReadFile(targetFile)
	if err != nil {
		return err
	}

	var stateFile struct {
		PreState struct {
			KeyVals types.StateKeyVals `json:"keyvals"`
		} `json:"pre_state"`
		PostState struct {
			KeyVals types.StateKeyVals `json:"keyvals"`
		} `json:"post_state"`
	}
	if err := json.Unmarshal(data, &stateFile); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}

	allKeyVals := append(stateFile.PreState.KeyVals, stateFile.PostState.KeyVals...)

	for _, kv := range allKeyVals {
		valBytes := []byte(kv.Value)
		if len(valBytes) == 0 {
			continue
		}

		isPre, err := merklization.IsPreimage(kv.Key, valBytes)
		if err != nil || !isPre {
			continue
		}

		blobHash := sha256.Sum256(valBytes)
		blobHashHex := hex.EncodeToString(blobHash[:])

		if codeHashArg != "" {
			normalized := strings.TrimPrefix(strings.ToLower(codeHashArg), "0x")
			if !strings.HasPrefix(blobHashHex, normalized) {
				continue
			}
		}

		if err := tryWriteDeBlobOutput(targetFile, valBytes, blobHashHex, outRoot); err == nil {
			return nil
		}
	}

	return fmt.Errorf("no valid PVM program preimage found for codehash %q", codeHashArg)
}

func tryWriteDeBlobOutput(targetFile string, programBlob []byte, codeHashHex, outRoot string) error {
	programCode, _, _, exitReason := PVM.SingleInitializer(PVM.StandardCodeFormat(programBlob), nil)
	if exitReason != PVM.ExitContinue {
		return fmt.Errorf("SingleInitializer failed: exit=%v", exitReason)
	}
	program, exitReason := PVM.DeBlobProgramCode(programCode)
	if exitReason != PVM.ExitContinue {
		return fmt.Errorf("DeBlobProgramCode failed: exit=%v", exitReason)
	}
	return writeDeBlobOutput(targetFile, programBlob, programCode, &program, codeHashHex, outRoot)
}

func deblobFromBin(targetFile, codeHashArg, outRoot string) error {
	programBlob, err := os.ReadFile(targetFile)
	if err != nil {
		return err
	}

	blobHash := sha256.Sum256(programBlob)
	blobHashHex := hex.EncodeToString(blobHash[:])

	if codeHashArg != "" {
		blobHashHex = strings.TrimPrefix(strings.ToLower(codeHashArg), "0x")
	}

	return tryWriteDeBlobOutput(targetFile, programBlob, blobHashHex, outRoot)
}

func writeDeBlobOutput(targetFile string, programBlob, programCode []byte, program *PVM.Program, codeHashHex, outRoot string) error {
	parentFolder := filepath.Base(filepath.Dir(targetFile))
	prefix := codeHashHex
	if len(prefix) > 16 {
		prefix = prefix[:16]
	}

	outDir := filepath.Join(outRoot, parentFolder, prefix)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	progPath := filepath.Join(outDir, "program.bin")
	if err := os.WriteFile(progPath, programBlob, 0644); err != nil {
		return err
	}

	kvPath := filepath.Join(outDir, "key_value.json")
	kvData, _ := json.MarshalIndent(map[string]string{
		"codehash": "0x" + codeHashHex,
	}, "", "  ")
	os.WriteFile(kvPath, kvData, 0644)

	meta := buildProgramMetadata(program, codeHashHex)
	if err := deblob.WriteMetadata(outDir, meta); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "wrote deblob output to %s\n", outDir)
	fmt.Fprintf(os.Stderr, "  instructions: %d\n", meta.InstructionCount)
	fmt.Fprintf(os.Stderr, "  blocks: %d\n", meta.BlockCount)
	return nil
}

func buildProgramMetadata(program *PVM.Program, codeHashHex string) *deblob.ProgramMetadata {
	blockCount := 0
	blockStartPCs := make(map[uint32]uint32)
	for pc, block := range program.BlockAt {
		if block == nil {
			continue
		}
		blockCount++
		for i := block.InstrStart; i < block.InstrEnd; i++ {
			if i < len(program.Instrs) {
				blockStartPCs[uint32(program.Instrs[i].PC)] = uint32(pc)
			}
		}
	}

	meta := &deblob.ProgramMetadata{
		FormatVersion:    1,
		CodeHash:         "0x" + codeHashHex,
		InstructionCount: len(program.Instrs),
		BlockCount:       blockCount,
	}

	meta.Instructions = make([]deblob.InstrMetaJSON, len(program.Instrs))
	for i, instr := range program.Instrs {
		bpc, _ := blockStartPCs[uint32(instr.PC)]
		meta.Instructions[i] = deblob.InstrMetaJSON{
			Index:       i,
			PC:          uint32(instr.PC),
			Opcode:      instr.Opcode,
			OpcodeName:  PVM.OpcodeName(instr.Opcode),
			SkipLen:     instr.SkipLen,
			Dst:         instr.Dst,
			Src:         instr.Src,
			Imm:         instr.Imm,
			BlockStartPC: bpc,
		}
	}

	meta.Blocks = make([]deblob.BlockMetaJSON, 0, blockCount)
	for pc, block := range program.BlockAt {
		if block == nil {
			continue
		}
		var endPC uint32
		if block.InstrEnd > 0 && block.InstrEnd <= len(program.Instrs) {
			last := program.Instrs[block.InstrEnd-1]
			endPC = uint32(last.PC) + uint32(last.SkipLen) + 1
		}
		meta.Blocks = append(meta.Blocks, deblob.BlockMetaJSON{
			StartPC:    uint32(pc),
			EndPC:      endPC,
			InstrStart: block.InstrStart,
			InstrEnd:   block.InstrEnd,
			GasCost:    int64(block.GasCost),
		})
	}

	return meta
}
