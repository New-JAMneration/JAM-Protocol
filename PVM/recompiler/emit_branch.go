//go:build linux && amd64

package recompiler

import (
	"fmt"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

// haltTarget is the PVM address that signifies HALT (0xFFFF0000).
const haltTarget = uint32(0xFFFF0000)

// emitCmpReg32Imm32Unsigned emits CMP r32, imm32 (32-bit compare)
// using the 32-bit opcode form that doesn't sign-extend to 64-bit.
func emitCmpReg32Imm32Unsigned(a *asm.Assembler, reg asm.Register, val uint32) {
	a.CmpReg32Imm32(reg, int32(val))
}

// emitHaltAtPC writes ExitPC=instrPC and ExitReason=HALT, then exits to Go.
func emitHaltAtPC(a *asm.Assembler, instrPC PVM.ProgramCounter) {
	a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(instrPC))
	a.MovImm64ToReg(RegScratch, uint64(PVM.ExitHalt))
	a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
	a.Jmp("exit_trampoline")
}

// emitExitToPC emits the exit sequence that sets ExitPC and ExitReason=CONTINUE,
// then jumps to exit_trampoline. Used for block-to-block transitions (Phase 4: always exit to Go dispatcher).
func emitExitToPC(a *asm.Assembler, targetPC PVM.ProgramCounter) {
	a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(targetPC))
	a.MovMemImm32(RegGuestBase, -int32(OffsetExitReason), 0) // ExitContinue = 0
	a.Jmp("exit_trampoline")
}

// ---- 4.9.1 Unconditional jump ----

// opcode 40: jump — unconditional jump to target PC
func (c *Compiler) emitJump(a *asm.Assembler, instr *PVM.InstrMeta) error {
	pc := instr.PC
	targetPC := PVM.ProgramCounter(instr.Imm[0])

	// Validate target is a basic block start
	if !c.program.Bitmasks.IsStartOfBasicBlock(targetPC) {
		// Target is not a valid basic block → emit panic
		a.MovImm64ToReg(RegScratch, uint64(PVM.ExitPanic))
		a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
		a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(pc))
		a.Jmp("exit_trampoline")
		return nil
	}

	emitLinkOrExit(a, c.linkTaken, targetPC)
	return nil
}

// ---- 4.9.2 Indirect jump ----

// opcode 50: jump_ind — indirect jump via Reg[rA] + vX
func (c *Compiler) emitJumpInd(a *asm.Assembler, instr *PVM.InstrMeta) error {
	pc := instr.PC
	xReg, vX := oneRegImmFromMeta(instr)

	// target = uint32(Reg[rA] + vX)
	a.MovRegToReg(RegScratch, xReg)
	emitAddUint64ToReg(a, RegScratch, vX)
	emitZeroExt32(a, RegScratch)

	// Check for HALT: target == 0xFFFF0000
	// RegScratch is already zero-extended to 32-bit. Use 32-bit CMP.
	haltLabel := fmt.Sprintf("halt_%d", pc)
	emitCmpReg32Imm32Unsigned(a, RegScratch, haltTarget)
	a.Jcc(asm.CondEQ, haltLabel)

	// 64-bit store of uint32 target (see emitDjumpExit).
	if err := c.emitDjumpNative(a, RegScratch, pc); err != nil {
		return err
	}

	// HALT path
	_ = a.BindLabel(haltLabel)
	emitHaltAtPC(a, pc)

	return nil
}

// ---- 4.9.3 Conditional branch (one reg + imm + offset) ----

// opcode 80: load_imm_jump — Reg[rA] = vX, then jump to target
func (c *Compiler) emitLoadImmJump(a *asm.Assembler, instr *PVM.InstrMeta) error {
	pc := instr.PC
	xReg, vX, targetPC := branchOneRegImmFromMeta(instr)
	if fitsInt32(vX) {
		a.MovImm32ToReg(xReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(xReg, vX)
	}

	if !c.program.Bitmasks.IsStartOfBasicBlock(targetPC) {
		a.MovImm64ToReg(RegScratch, uint64(PVM.ExitPanic))
		a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
		a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(pc))
		a.Jmp("exit_trampoline")
		return nil
	}

	emitLinkOrExit(a, c.linkTaken, targetPC)
	return nil
}

// opcode 81-90: branch_xx_imm — conditional branch with immediate comparison
func (c *Compiler) emitBranchImm(a *asm.Assembler, instr *PVM.InstrMeta, cc asm.ConditionCode) error {
	pc := instr.PC
	xReg, vX, targetPC := branchOneRegImmFromMeta(instr)
	takenLabel := fmt.Sprintf("taken_%d", pc)

	// Compare Reg[rA] with vX
	if fitsInt32(vX) {
		a.CmpRegImm32(xReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(RegScratch, vX)
		a.CmpRegReg(xReg, RegScratch)
	}

	a.Jcc(cc, takenLabel)

	// Not taken: fall through to next instruction (block exit handled by caller)
	nextPC := fallthroughPC(instr)
	emitLinkOrExit(a, c.linkFallthrough, nextPC)

	// Taken: exit to target PC
	_ = a.BindLabel(takenLabel)
	if !c.program.Bitmasks.IsStartOfBasicBlock(targetPC) {
		a.MovImm64ToReg(RegScratch, uint64(PVM.ExitPanic))
		a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
		a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(pc))
		a.Jmp("exit_trampoline")
	} else {
		emitLinkOrExit(a, c.linkTaken, targetPC)
	}

	return nil
}

// ---- 4.9.4 Conditional branch (two registers + offset) ----

// opcode 170-175: branch_xx — two-register comparison
func (c *Compiler) emitBranch(a *asm.Assembler, instr *PVM.InstrMeta, cc asm.ConditionCode) error {
	pc := instr.PC
	aReg, bReg, targetPC := branchTwoRegFromMeta(instr)
	takenLabel := fmt.Sprintf("taken_%d", pc)

	a.CmpRegReg(aReg, bReg)
	a.Jcc(cc, takenLabel)

	// Not taken
	nextPC := fallthroughPC(instr)
	emitLinkOrExit(a, c.linkFallthrough, nextPC)

	// Taken
	_ = a.BindLabel(takenLabel)
	if !c.program.Bitmasks.IsStartOfBasicBlock(targetPC) {
		a.MovImm64ToReg(RegScratch, uint64(PVM.ExitPanic))
		a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
		a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(pc))
		a.Jmp("exit_trampoline")
	} else {
		emitLinkOrExit(a, c.linkTaken, targetPC)
	}

	return nil
}

// ---- 4.9.5 load_imm_jump_ind ----

// emitDjumpExit stores raw jump-table address in ExitPC and exits with DjumpCallID for
// PVM.DjumpResolve in BlockBasedInvoke (includes basic-block-start validation).
// Uses 64-bit MOV into [R15-32]: low 4 bytes are ExitPC (uint32); upper 4 are struct padding.
func emitDjumpExit(a *asm.Assembler, targetReg asm.Register) {
	a.MovRegToMem(RegGuestBase, -int32(OffsetExitPC), targetReg)
	a.MovImm64ToReg(RegScratch, uint64(PVM.ExitHostCall)|uint64(DjumpCallID))
	a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
	a.Jmp("exit_trampoline")
}

// opcode 180: load_imm_jump_ind — Reg[rA] = vX, then djump(uint32(Reg[rB] + vY))
func (c *Compiler) emitLoadImmJumpInd(a *asm.Assembler, instr *PVM.InstrMeta) error {
	pc := instr.PC
	aReg, bReg, vX, vY := twoRegTwoImmFromMeta(instr)

	// First compute target from rB before overwriting rA (in case rA == rB)
	a.MovRegToReg(RegScratch, bReg)
	emitAddUint64ToReg(a, RegScratch, vY)
	emitZeroExt32(a, RegScratch)

	// Load immediate into rA
	if fitsInt32(vX) {
		a.MovImm32ToReg(aReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(aReg, vX)
	}

	// Check for HALT
	haltLabel := fmt.Sprintf("halt_%d", pc)
	emitCmpReg32Imm32Unsigned(a, RegScratch, haltTarget)
	a.Jcc(asm.CondEQ, haltLabel)

	if err := c.emitDjumpNative(a, RegScratch, pc); err != nil {
		return err
	}

	_ = a.BindLabel(haltLabel)
	emitHaltAtPC(a, pc)

	return nil
}
