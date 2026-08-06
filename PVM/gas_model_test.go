package PVM

import "testing"

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
