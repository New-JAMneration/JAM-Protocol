package PVM

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// GP A.9 gas-model vectors under new-gas-cost-model. program is a full PVM blob;
// block-gas-costs holds expected gascostforblock per basic-block entry PC.
const (
	gasModelProgramsDir      = "../pkg/test_data/new-gas-cost-model/tests/programs"
	gasModelIntegrationDir   = "../pkg/test_data/new-gas-cost-model/integration-tests"
)

type blockGasEntry struct {
	PC   ProgramCounter `json:"pc"`
	Cost Gas            `json:"cost"`
}

// blockGasCosts accepts vectors in either array form
// [{"pc":0,"cost":1},...] or map form {"0":1,"6":52,...} (integration tests).
type blockGasCosts []blockGasEntry

func (b *blockGasCosts) UnmarshalJSON(data []byte) error {
	var arr []blockGasEntry
	if err := json.Unmarshal(data, &arr); err == nil {
		*b = arr
		return nil
	}
	var m map[string]Gas
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	entries := make([]blockGasEntry, 0, len(m))
	for pcStr, cost := range m {
		pc, err := strconv.ParseUint(pcStr, 10, 64)
		if err != nil {
			return err
		}
		entries = append(entries, blockGasEntry{PC: ProgramCounter(pc), Cost: cost})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].PC < entries[j].PC })
	*b = entries
	return nil
}

type gasModelVector struct {
	Name         string         `json:"name"`
	InitialPC    ProgramCounter `json:"initial-pc"`
	InitialGas   uint64         `json:"initial-gas"`
	Program      []byte         `json:"program"`
	BlockGasCost blockGasCosts  `json:"block-gas-costs"`
	Steps        []struct {
		Run    *struct{} `json:"run"`
		Assert *struct {
			Gas    uint64 `json:"gas"`
			Status string `json:"status"`
		} `json:"assert"`
	} `json:"steps"`
}

func loadGasModelVector(t *testing.T, path string) gasModelVector {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var vec gasModelVector
	if err := json.Unmarshal(raw, &vec); err != nil {
		t.Fatalf("json %s: %v", path, err)
	}
	if vec.Name == "" {
		vec.Name = filepath.Base(path)
	}
	return vec
}

func programFromGasVectorBlob(t *testing.T, blob []byte) *Program {
	t.Helper()
	prog, reason := DeBlobProgramCode(blob, 0)
	if reason != ExitContinue {
		t.Fatalf("DeBlobProgramCode: %v", reason)
	}
	return &prog
}

func assertBlockGasCosts(t *testing.T, prog *Program, expected blockGasCosts) {
	t.Helper()
	for _, e := range expected {
		got := GasCostForBlock(prog, e.PC)
		if got != e.Cost {
			t.Errorf("block @pc=%d: got %d, want %d", e.PC, got, e.Cost)
		}
	}
}

func finalAssertGas(vec gasModelVector) (uint64, string, bool) {
	for _, s := range vec.Steps {
		if s.Assert != nil {
			return s.Assert.Gas, s.Assert.Status, true
		}
	}
	return 0, "", false
}

// TestGasVectorHarnessSanity cross-checks vector JSON: deblob, block PCs, and that
// block-gas-costs for single-block programs match integration gas consumption.
func TestGasVectorHarnessSanity(t *testing.T) {
	for _, path := range gasModelJSONFiles(t, gasModelProgramsDir, "gas_") {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			vec := loadGasModelVector(t, path)
			if len(vec.BlockGasCost) == 0 {
				t.Fatal("missing block-gas-costs")
			}

			prog, reason := DeBlobProgramCode(vec.Program, 0)
			if reason != ExitContinue {
				t.Fatalf("DeBlobProgramCode: %v", reason)
			}

			for _, e := range vec.BlockGasCost {
				if int(e.PC) >= len(prog.InstrIdxAt) || prog.InstrIdxAt[e.PC] < 0 {
					t.Fatalf("block-gas-costs pc=%d is not a valid instruction entry", e.PC)
				}
				block := prog.LookupBlock(e.PC)
				if block == nil || block.StartPC != e.PC {
					t.Fatalf("pc=%d is not a basic block entry", e.PC)
				}
				fromAPI := GasCostForBlock(&prog, e.PC)
				if fromAPI != block.GasCost {
					t.Errorf("pc=%d: GasCostForBlock=%d != preDecode=%d", e.PC, fromAPI, block.GasCost)
				}
			}

			// Single-block programs that halt with panic charge exactly one
			// block; out-of-gas / multi-iteration runs consume many charges.
			if len(vec.BlockGasCost) == 1 {
				finalGas, status, ok := finalAssertGas(vec)
				if !ok {
					t.Fatal("no assert.gas in steps")
				}
				if status == "" || status == "panic" {
					consumed := vec.InitialGas - finalGas
					if Gas(consumed) != vec.BlockGasCost[0].Cost {
						t.Errorf("integration consumed %d gas, block-gas-costs want %d", consumed, vec.BlockGasCost[0].Cost)
					}
				}
			}
		})
	}
}

func gasModelJSONFiles(t *testing.T, dir, prefix string) []string {
	t.Helper()
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("gas model vectors not found at %s (init submodule?): %v", dir, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		if prefix != "" && !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		files = append(files, filepath.Join(dir, e.Name()))
	}
	return files
}

func runGasModelProgramPrefix(t *testing.T, prefix string) {
	t.Helper()
	for _, path := range gasModelJSONFiles(t, gasModelProgramsDir, prefix) {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			vec := loadGasModelVector(t, path)
			if len(vec.BlockGasCost) == 0 {
				t.Fatal("missing block-gas-costs")
			}
			prog, reason := DeBlobProgramCode(vec.Program, 0)
			if reason != ExitContinue {
				t.Fatalf("DeBlobProgramCode: %v", reason)
			}
			assertBlockGasCosts(t, &prog, vec.BlockGasCost)
		})
	}
}

// TestGasModelProgramVectors checks gas_*.json against GasCostForBlock
// (A.9 eq:gascostforblock).
func TestGasModelProgramVectors(t *testing.T) {
	runGasModelProgramPrefix(t, "gas_")
}

// TestGasModelInstVectors checks per-opcode inst_*.json block-gas-costs.
func TestGasModelInstVectors(t *testing.T) {
	runGasModelProgramPrefix(t, "inst_")
}

// TestGasModelRiscvVectors checks compiled riscv_*.json block-gas-costs.
func TestGasModelRiscvVectors(t *testing.T) {
	runGasModelProgramPrefix(t, "riscv_")
}

// TestGasModelMultistepVectors checks multistep_*.json (ecalli/paging mid-block).
func TestGasModelMultistepVectors(t *testing.T) {
	runGasModelProgramPrefix(t, "multistep_")
}

// TestGasModelIntegrationVectors checks large program-only vectors (program + block-gas-costs).
func TestGasModelIntegrationVectors(t *testing.T) {
	for _, path := range gasModelJSONFiles(t, gasModelIntegrationDir, "") {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			vec := loadGasModelVector(t, path)
			if len(vec.BlockGasCost) == 0 {
				t.Fatal("missing block-gas-costs")
			}
			prog := programFromGasVectorBlob(t, vec.Program)
			assertBlockGasCosts(t, prog, vec.BlockGasCost)
		})
	}
}
