package PVM

import "testing"

// TestGasSimConvergesBeyondFormerStepCap builds a long serial high-latency
// block that exceeded the old maxSteps=100000 guard (~1613 DIV/REM per review).
// A.9 requires convergence; an artificial cap must not panic or undercharge.
func TestGasSimConvergesBeyondFormerStepCap(t *testing.T) {
	const nDiv = 2000
	code := make(ProgramCode, 0, nDiv*3+1)
	bm := make(Bitmask, 0, nDiv*3+1)
	for range nDiv {
		// div_u_64: r0 = r0 / r1 — dependent chain, D-unit serial.
		// Bitmask is one octet per code byte: 0x03 = instruction start.
		code = append(code, 203, 0x10, 0)
		bm = append(bm, 0x03, 0x00, 0x00)
	}
	code = append(code, 0) // trap terminator
	bm = append(bm, 0x03)

	prog := decodedGasTestProgram(t, code, bm)
	block := prog.LookupBlock(0)
	if block == nil {
		t.Fatal("missing block at PC 0")
	}
	if block.GasCost < 50_000 {
		t.Fatalf("block gas = %d, want large serial-div cost (converged)", block.GasCost)
	}
	if got := GasCostForBlock(&prog, 0); got != block.GasCost {
		t.Fatalf("GasCostForBlock=%d preDecode=%d", got, block.GasCost)
	}
}

func TestBlockGasFromCycles(t *testing.T) {
	tests := []struct {
		cycles Gas
		want   Gas
	}{
		{0, 1},
		{1, 1},
		{3, 1},
		{4, 1},
		{5, 2},
		{10, 7},
	}
	for _, tt := range tests {
		if got := blockGasFromCycles(tt.cycles); got != tt.want {
			t.Fatalf("blockGasFromCycles(%d) = %d, want %d", tt.cycles, got, tt.want)
		}
	}
}

func TestGasCostSingleTrapBlock(t *testing.T) {
	prog := decodedGasTestProgram(t,
		ProgramCode{0},
		Bitmask{0x01},
	)
	block := prog.LookupBlock(0)
	if block == nil {
		t.Fatal("missing block at PC 0")
	}
	if block.GasCost != 2 {
		t.Fatalf("trap block gas = %d, want 2 (max(cycles-3,1))", block.GasCost)
	}
}

func TestGasCostMoveRegBlock(t *testing.T) {
	// move_reg r0<-r1; trap
	prog := decodedGasTestProgram(t,
		ProgramCode{100, 0x10, 0},
		Bitmask{0x03, 0x00, 0x03},
	)
	block := prog.LookupBlock(0)
	if block == nil {
		t.Fatal("missing block at PC 0")
	}
	if block.GasCost < 1 {
		t.Fatalf("move_reg+trap block gas = %d, want >= 1", block.GasCost)
	}
}

func TestGasCostBranchToTrap(t *testing.T) {
	// branch_eq r0,r1 -> trap (PC 7); fallthrough: unlikely (PC 6)
	prog := decodedGasTestProgram(t,
		ProgramCode{
			170, 0x01, 7, 0, 0, 0, // branch_eq at 0, target PC 7
			2, // unlikely (fallthrough at PC 6)
			0, // trap (branch target at PC 7)
		},
		Bitmask{0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x03},
	)
	block := prog.LookupBlock(0)
	if block == nil {
		t.Fatal("missing block at PC 0")
	}
	cost := InstructionCost(&prog, &prog.Instrs[0])
	if cost.Cycles != 1 {
		t.Fatalf("branch to trap/unlikely cycles = %d, want 1", cost.Cycles)
	}
	if block.GasCost < 1 {
		t.Fatalf("branch block gas = %d, want >= 1", block.GasCost)
	}
}

func TestBranchCyclesOutOfRangeTargetIsTrap(t *testing.T) {
	prog := decodedGasTestProgram(t,
		ProgramCode{
			170, 0x01, 100, 0, 0, 0, // branch_eq target PC 100 (past end)
			0, // trap fallthrough at PC 6
		},
		Bitmask{0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03},
	)
	cost := InstructionCost(&prog, &prog.Instrs[0])
	if cost.Cycles != 1 {
		t.Fatalf("out-of-range branch target cycles = %d, want 1", cost.Cycles)
	}
}

func TestInstructionCostEcalli(t *testing.T) {
	instr := InstrMeta{Opcode: 10}
	cost := InstructionCost(&Program{}, &instr)
	if cost.Cycles != 100 || cost.Decode != 4 || cost.Units.A != 1 {
		t.Fatalf("ecalli cost = %+v, want cycles=100 decode=4 A=1", cost)
	}
}

func TestInstRegsEcalli(t *testing.T) {
	instr := InstrMeta{Opcode: 10, Dst: 0, Src: [2]uint8{0, 1}}
	if got := instSrcRegs(&instr); len(got) != 0 {
		t.Fatalf("ecalli src regs = %v, want none", got)
	}
	if got := instDstRegs(&instr); len(got) != 0 {
		t.Fatalf("ecalli dst regs = %v, want none", got)
	}
}

func decodedGasTestProgram(t *testing.T, code ProgramCode, bitmask Bitmask) Program {
	t.Helper()
	prog := Program{
		InstructionData: code,
		Bitmasks:        bitmask,
	}
	if reason := prog.preDecodeBlocks(); reason != ExitContinue {
		t.Fatalf("preDecodeBlocks() = %v, want %v", reason, ExitContinue)
	}
	return prog
}
