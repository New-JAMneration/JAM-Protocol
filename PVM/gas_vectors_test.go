package PVM

import (
	"bytes"
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
	gasModelProgramsDir    = "../pkg/test_data/new-gas-cost-model/tests/programs"
	gasModelIntegrationDir = "../pkg/test_data/new-gas-cost-model/integration-tests"
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

type gasModelMemSlice struct {
	Address  uint64 `json:"address"`
	Contents []byte `json:"contents"`
}

type gasModelAssert struct {
	Status           string             `json:"status"`
	Hostcall         *uint64            `json:"hostcall"`
	PageFaultAddress *uint64            `json:"page-fault-address"`
	Gas              uint64             `json:"gas"`
	PC               ProgramCounter     `json:"pc"`
	Regs             []uint64           `json:"regs"`
	Memory           []gasModelMemSlice `json:"memory"`
}

type gasModelStep struct {
	Run    *struct{}       `json:"run"`
	Assert *gasModelAssert `json:"assert"`
	SetReg *struct {
		Reg   uint8  `json:"reg"`
		Value uint64 `json:"value"`
	} `json:"set-reg"`
	Map *struct {
		Address    uint64 `json:"address"`
		Length     uint64 `json:"length"`
		IsWritable bool   `json:"is-writable"`
	} `json:"map"`
	Write *gasModelMemSlice `json:"write"`
}

type gasModelVector struct {
	Name         string         `json:"name"`
	InitialPC    ProgramCounter `json:"initial-pc"`
	InitialGas   uint64         `json:"initial-gas"`
	Program      []byte         `json:"program"`
	BlockGasCost blockGasCosts  `json:"block-gas-costs"`
	Steps        []gasModelStep `json:"steps"`
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
	// Gas fixtures may omit a final terminator; use the gas-model decode path.
	prog, reason := deblobProgramForGasModel(blob)
	if reason != ExitContinue {
		t.Fatalf("deblobProgramForGasModel: %v", reason)
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

			prog, reason := deblobProgramForGasModel(vec.Program)
			if reason != ExitContinue {
				t.Fatalf("deblobProgramForGasModel: %v", reason)
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
			prog, reason := deblobProgramForGasModel(vec.Program)
			if reason != ExitContinue {
				t.Fatalf("deblobProgramForGasModel: %v", reason)
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

// TestGasModelMultistepVectors executes multistep_*.json step sequences
// (run / assert / set-reg / map / write) and still checks block-gas-costs.
func TestGasModelMultistepVectors(t *testing.T) {
	for _, path := range gasModelJSONFiles(t, gasModelProgramsDir, "multistep_") {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			vec := loadGasModelVector(t, path)
			if len(vec.BlockGasCost) == 0 {
				t.Fatal("missing block-gas-costs")
			}
			prog := programFromGasVectorBlob(t, vec.Program)
			assertBlockGasCosts(t, prog, vec.BlockGasCost)
			runGasModelMultistep(t, prog, vec)
		})
	}
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

type gasModelHarness struct {
	interp   *Interpreter
	pc       ProgramCounter
	lastExit ExitReason
	assertPC ProgramCounter
}

func runGasModelMultistep(t *testing.T, prog *Program, vec gasModelVector) {
	t.Helper()
	mem := &Memory{Pages: map[uint32]*Page{}}
	gas := Gas(vec.InitialGas)
	h := &gasModelHarness{
		interp: &Interpreter{
			Program:    prog,
			Registers:  Registers{},
			Memory:     mem,
			Gas:        gas,
			GasCharged: false,
		},
		pc: vec.InitialPC,
	}

	for i, step := range vec.Steps {
		switch {
		case step.Run != nil:
			h.lastExit, h.assertPC, h.pc = gasModelInvokeUntilExit(h.interp, h.pc)

		case step.Assert != nil:
			assertGasModelState(t, i, h.interp, h.lastExit, h.assertPC, step.Assert)

		case step.SetReg != nil:
			if int(step.SetReg.Reg) >= len(h.interp.Registers) {
				t.Fatalf("step %d: set-reg r%d out of range", i, step.SetReg.Reg)
			}
			h.interp.Registers[step.SetReg.Reg] = step.SetReg.Value

		case step.Map != nil:
			gasModelMapPages(mem, step.Map.Address, step.Map.Length, step.Map.IsWritable)

		case step.Write != nil:
			if !isWriteable(step.Write.Address, uint64(len(step.Write.Contents)), *mem) {
				t.Fatalf("step %d: write at %#x not writable", i, step.Write.Address)
			}
			mem.Write(step.Write.Address, step.Write.Contents)

		default:
			t.Fatalf("step %d: unrecognized action", i)
		}
	}
}

// gasModelInvokeUntilExit runs like BlockBasedInvokeDecodedBlocks but reports
// assertPC as the interrupting instruction (ecalli / fault / halt), while
// resumePC is where the next run should continue.
func gasModelInvokeUntilExit(interp *Interpreter, pc ProgramCounter) (exit ExitReason, assertPC, resumePC ProgramCounter) {
	prog := interp.Program
	for {
		if int(pc) >= len(prog.InstrIdxAt) {
			return ExitPanic, pc, pc
		}
		instrIdx := prog.InstrIdxAt[pc]
		block := prog.BlockContaining(pc)
		if instrIdx < 0 || block == nil {
			return ExitPanic, pc, pc
		}
		startIdx := int(instrIdx)
		if startIdx < block.InstrStart || startIdx >= block.InstrEnd {
			return ExitPanic, pc, pc
		}

		if !interp.GasCharged {
			blockGas := blockGasAtPC(prog, pc, block)
			if interp.Gas < blockGas {
				return ExitOOG, pc, pc
			}
			interp.Gas -= blockGas
			interp.GasCharged = true
		}

		instrs := prog.Instrs[startIdx:block.InstrEnd]
		branchTaken := false
		for i := range instrs {
			instr := &instrs[i]
			exitReason, newPC := instr.Exec(interp, instr)
			reason := exitReason.GetReasonType()
			if IsBlockTerminator(instr.Opcode) && (reason == CONTINUE || reason == HOST_CALL) {
				interp.GasCharged = false
			}
			switch reason {
			case PANIC, HALT:
				return exitReason, instr.PC, instr.PC
			case PAGE_FAULT, OUT_OF_GAS:
				return exitReason, instr.PC, instr.PC
			case HOST_CALL:
				next := instr.PC + ProgramCounter(instr.SkipLen) + 1
				return exitReason, instr.PC, next
			}
			if IsBlockTerminator(instr.Opcode) {
				pc = newPC
				branchTaken = true
				break
			}
		}
		if !branchTaken {
			last := &instrs[len(instrs)-1]
			pc = last.PC + ProgramCounter(last.SkipLen) + 1
		}
	}
}

func gasModelMapPages(mem *Memory, addr, length uint64, writable bool) {
	access := MemoryReadOnly
	if writable {
		access = MemoryReadWrite
	}
	end := addr + length
	allocateMemorySegment(mem, uint32(addr), uint32(end), nil, access)
}

func assertGasModelState(t *testing.T, step int, interp *Interpreter, exit ExitReason, assertPC ProgramCounter, want *gasModelAssert) {
	t.Helper()
	gotStatus := gasModelStatus(exit)
	if gotStatus != want.Status {
		t.Fatalf("step %d: status = %q, want %q (exit=%v)", step, gotStatus, want.Status, exit)
	}
	if uint64(interp.Gas) != want.Gas {
		t.Fatalf("step %d: gas = %d, want %d", step, interp.Gas, want.Gas)
	}
	if assertPC != want.PC {
		t.Fatalf("step %d: pc = %d, want %d", step, assertPC, want.PC)
	}
	if want.Hostcall != nil {
		if exit.GetReasonType() != HOST_CALL || uint64(exit.GetHostCallID()) != *want.Hostcall {
			t.Fatalf("step %d: hostcall = %d, want %d", step, exit.GetHostCallID(), *want.Hostcall)
		}
	}
	if want.PageFaultAddress != nil {
		// Draft vectors report the faulting page base; interpreter payload is the
		// access address (may be mid-page).
		gotFault := uint64(exit.GetPageFaultAddress()) &^ (uint64(ZP) - 1)
		if exit.GetReasonType() != PAGE_FAULT || gotFault != *want.PageFaultAddress {
			t.Fatalf("step %d: page-fault page = %#x (raw %#x), want %#x",
				step, gotFault, exit.GetPageFaultAddress(), *want.PageFaultAddress)
		}
	}
	if len(want.Regs) > 0 {
		if len(want.Regs) != len(interp.Registers) {
			t.Fatalf("step %d: regs len = %d, want %d", step, len(interp.Registers), len(want.Regs))
		}
		for i, w := range want.Regs {
			if interp.Registers[i] != w {
				t.Fatalf("step %d: r%d = %#x, want %#x", step, i, interp.Registers[i], w)
			}
		}
	}
	for _, m := range want.Memory {
		got := interp.Memory.Read(m.Address, uint64(len(m.Contents)))
		if !bytes.Equal(got, m.Contents) {
			t.Fatalf("step %d: memory[%#x] = %x, want %x", step, m.Address, got, m.Contents)
		}
	}
}

func gasModelStatus(exit ExitReason) string {
	switch exit.GetReasonType() {
	case HALT:
		return "halt"
	case PANIC:
		return "panic"
	case OUT_OF_GAS:
		return "out-of-gas"
	case PAGE_FAULT:
		return "page-fault"
	case HOST_CALL:
		return "ecalli"
	default:
		return exit.String()
	}
}
