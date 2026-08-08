package PVM

import "testing"

func TestBlockBasedInvokeDecodedBlocksChargesContainingBlock(t *testing.T) {
	prog := decodedInvocationTestProgram(
		t,
		ProgramCode{2, 1},
		Bitmask{0x03, 0x01},
	)
	var memory Memory
	interp := NewInterpreter(&prog, Registers{}, &memory, 0)

	reason, pc := interp.BlockBasedInvokeDecodedBlocks(1)
	if reason != ExitOOG {
		t.Fatalf("exit reason = %v, want %v", reason, ExitOOG)
	}
	if pc != 1 {
		t.Fatalf("pc = %d, want 1", pc)
	}
	if interp.Gas != 0 {
		t.Fatalf("gas = %d, want 0", interp.Gas)
	}
	if interp.GasCharged {
		t.Fatal("gas flag set after failed block charge")
	}
}

func TestBlockBasedInvokeDecodedBlocksResumesAfterHostCall(t *testing.T) {
	prog := ecalliFallthroughTrapProgram(t)
	block := prog.LookupBlock(0)
	if block == nil {
		t.Fatal("missing block at PC 0")
	}

	var memory Memory
	interp := NewInterpreter(&prog, Registers{}, &memory, block.GasCost)

	reason, pc := interp.BlockBasedInvokeDecodedBlocks(0)
	if reason.GetReasonType() != HOST_CALL {
		t.Fatalf("first exit reason = %v, want host call", reason)
	}
	if pc != 2 {
		t.Fatalf("first pc = %d, want 2", pc)
	}
	if interp.Gas != 0 {
		t.Fatalf("gas after block charge = %d, want 0", interp.Gas)
	}
	if !interp.GasCharged {
		t.Fatal("ecalli must preserve GasCharged")
	}

	// Resume with flag still ⊤ — no second charge (A.4).
	reason, pc = interp.BlockBasedInvokeDecodedBlocks(pc)
	if reason != ExitPanic {
		t.Fatalf("second exit reason = %v, want panic (trap)", reason)
	}
	if pc != 0 {
		t.Fatalf("second pc = %d, want 0 after trap", pc)
	}
	if interp.Gas != 0 {
		t.Fatalf("gas after resume = %d, want 0", interp.Gas)
	}
	if !interp.GasCharged {
		t.Fatal("trap/panic must preserve GasCharged")
	}
}

func TestBlockBasedInvokeChargesContainingBlock(t *testing.T) {
	prog := decodedInvocationTestProgram(
		t,
		ProgramCode{2, 1},
		Bitmask{0x03, 0x01},
	)
	var memory Memory
	interp := NewInterpreter(&prog, Registers{}, &memory, 0)

	reason, pc := interp.BlockBasedInvoke(1)
	if reason != ExitOOG {
		t.Fatalf("exit reason = %v, want %v", reason, ExitOOG)
	}
	if pc != 1 {
		t.Fatalf("pc = %d, want 1", pc)
	}
	if interp.Gas != 0 {
		t.Fatalf("gas = %d, want 0", interp.Gas)
	}
	if interp.GasCharged {
		t.Fatal("gas flag set after failed block charge")
	}
}

func TestBlockBasedInvokeResumesAfterHostCall(t *testing.T) {
	prog := ecalliFallthroughTrapProgram(t)
	block := prog.LookupBlock(0)
	if block == nil {
		t.Fatal("missing block at PC 0")
	}

	var memory Memory
	interp := NewInterpreter(&prog, Registers{}, &memory, block.GasCost)

	reason, pc := interp.BlockBasedInvoke(0)
	if reason.GetReasonType() != HOST_CALL {
		t.Fatalf("first exit reason = %v, want host call", reason)
	}
	if pc != 2 {
		t.Fatalf("first pc = %d, want 2", pc)
	}
	if interp.Gas != 0 {
		t.Fatalf("gas after block charge = %d, want 0", interp.Gas)
	}
	if !interp.GasCharged {
		t.Fatal("ecalli must preserve GasCharged")
	}

	reason, pc = interp.BlockBasedInvoke(pc)
	if reason != ExitPanic {
		t.Fatalf("second exit reason = %v, want panic (trap)", reason)
	}
	if pc != 0 {
		t.Fatalf("second pc = %d, want 0 after trap", pc)
	}
	if interp.Gas != 0 {
		t.Fatalf("gas after resume = %d, want 0", interp.Gas)
	}
}

func TestBlockBasedInvokeReturnsPageFault(t *testing.T) {
	// store_u8 r0, 0x10000; trap — address is above ZZ but unmapped.
	prog := decodedInvocationTestProgram(
		t,
		ProgramCode{59, 0, 0x00, 0x00, 0x01, 0},
		Bitmask{0x03, 0x00, 0x00, 0x00, 0x00, 0x03},
	)
	var memory Memory
	interp := NewInterpreter(&prog, Registers{}, &memory, 30)

	reason, pc := interp.BlockBasedInvoke(0)
	if reason.GetReasonType() != PAGE_FAULT {
		t.Fatalf("exit reason = %v, want page fault", reason)
	}
	if pc != 0 {
		t.Fatalf("pc = %d, want 0 (faulting instruction)", pc)
	}
	if got := reason.GetPageFaultAddress(); got != 0x10000 {
		t.Fatalf("fault address = 0x%x, want 0x10000", got)
	}
}

func TestDebugSingleStepInvokeChargesContainingBlock(t *testing.T) {
	prog := decodedInvocationTestProgram(
		t,
		ProgramCode{2, 1},
		Bitmask{0x03, 0x01},
	)
	var memory Memory
	interp := NewInterpreter(&prog, Registers{}, &memory, 0)

	reason, pc := interp.DebugSingleStepInvoke(1)
	if reason != ExitOOG {
		t.Fatalf("exit reason = %v, want %v", reason, ExitOOG)
	}
	if pc != 1 {
		t.Fatalf("pc = %d, want 1", pc)
	}
	if interp.Gas != 0 {
		t.Fatalf("gas = %d, want 0", interp.Gas)
	}
	if interp.GasCharged {
		t.Fatal("gas flag set after failed block charge")
	}
}

func TestDebugSingleStepInvokeResumesAfterHostCall(t *testing.T) {
	prog := ecalliFallthroughTrapProgram(t)
	block := prog.LookupBlock(0)
	if block == nil {
		t.Fatal("missing block at PC 0")
	}

	var memory Memory
	interp := NewInterpreter(&prog, Registers{}, &memory, block.GasCost)

	reason, pc := interp.DebugSingleStepInvoke(0)
	if reason.GetReasonType() != HOST_CALL {
		t.Fatalf("first exit reason = %v, want host call", reason)
	}
	if pc != 2 {
		t.Fatalf("first pc = %d, want 2", pc)
	}
	if interp.Gas != 0 {
		t.Fatalf("gas after block charge = %d, want 0", interp.Gas)
	}
	if !interp.GasCharged {
		t.Fatal("ecalli must preserve GasCharged")
	}

	reason, pc = interp.DebugSingleStepInvoke(pc)
	if reason != ExitPanic {
		t.Fatalf("second exit reason = %v, want panic (trap)", reason)
	}
	if pc != 0 {
		t.Fatalf("second pc = %d, want 0 after trap", pc)
	}
}

func TestBlockGasAtPCChargesFullContainingBlock(t *testing.T) {
	prog := ecalliFallthroughTrapProgram(t)
	block := prog.LookupBlock(0)
	if block == nil {
		t.Fatal("missing block at PC 0")
	}
	if got := blockGasAtPC(&prog, 0, block); got != block.GasCost {
		t.Fatalf("block entry: got %d, want cached %d", got, block.GasCost)
	}
	mid := prog.BlockContaining(2)
	if mid == nil {
		t.Fatal("missing containing block at pc=2")
	}
	if got := blockGasAtPC(&prog, 2, mid); got != block.GasCost {
		t.Fatalf("mid-block entry: got %d, want full containing block %d", got, block.GasCost)
	}
}

func TestSelfLoopJumpRetakesTerminator(t *testing.T) {
	// jump to PC 0 (self); fallthrough unreachable.
	prog := decodedGasTestProgram(t,
		ProgramCode{40, 0, 0, 0, 0}, // jump imm=0
		Bitmask{0x03, 0x00, 0x00, 0x00, 0x00},
	)
	gas := Gas(10_000)
	interp := &Interpreter{Program: &prog, Gas: gas, GasCharged: false}
	exit, pc := interp.BlockBasedInvokeDecodedBlocks(0)
	// Should OOG eventually while looping, with PC at the jump.
	if exit.GetReasonType() != OUT_OF_GAS {
		t.Fatalf("exit = %v, want OOG (self-loop must re-enter block)", exit)
	}
	if pc != 0 {
		t.Fatalf("OOG pc = %d, want 0 (self-loop)", pc)
	}
}

func ecalliFallthroughTrapProgram(t *testing.T) Program {
	t.Helper()
	// ecalli 0; load_imm_64 r0, 42; trap — matches recompiler host-call gas tests.
	inst := []byte{10, 0, 51, 0, 42, 0}
	blob := buildTestBlob(t, inst, []int{0, 2, 5})
	prog, reason := DeBlobProgramCode(blob, 0)
	if reason != ExitContinue {
		t.Fatalf("DeBlobProgramCode: %v", reason)
	}
	return prog
}

func decodedInvocationTestProgram(t *testing.T, code ProgramCode, bitmask Bitmask) Program {
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
