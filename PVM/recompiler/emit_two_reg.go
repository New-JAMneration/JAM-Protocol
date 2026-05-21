//go:build linux && amd64

package recompiler

import (
	"fmt"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

// opcode 100: move_reg — Reg[rA] = Reg[rB]
func (c *Compiler) emitMoveReg(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	if dReg != aReg {
		a.MovRegToReg(dReg, aReg)
	}
	return nil
}

// opcode 101: sbrk — heap expansion.
//
// The inline paths are an optimization that short-circuits the runtime exit only
// when no mprotect is needed:
//   - amount == 0:                   rD = heapPointer (query only, no growth)
//   - amount != 0, no page crossing: pages already mapped — update heapPointer + rD inline
//
// The page-crossing case MUST take the runtime exit to Go (HandleSbrk → mprotect);
// skipping it would leave the new pages PROT_NONE and the next guest write would
// raise a hardware SIGSEGV.
func (c *Compiler) emitSbrk(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	pc := instr.PC
	queryLabel := fmt.Sprintf("sbrk_query_%08x", uint32(pc))
	doneLabel := fmt.Sprintf("sbrk_done_%08x", uint32(pc))

	a.TestRegReg(aReg, aReg)
	a.Jcc(asm.CondEQ, queryLabel)

	c.emitSbrkExpand(a, dReg, aReg, instr, doneLabel)

	_ = a.BindLabel(queryLabel)
	a.MovMemToReg(dReg, RegGuestBase, -int32(OffsetHeapPointer))
	_ = a.BindLabel(doneLabel)
	return nil
}

// emitSbrkExpand handles amount != 0. Three outcomes:
//   - overflow or limit exceeded → rD = 0, continue in JIT
//   - no page crossing → update heapPointer and rD inline, continue in JIT
//   - page crossing → exit to Go for mprotect (HandleSbrk)
func (c *Compiler) emitSbrkExpand(a *asm.Assembler, dReg, aReg asm.Register, instr *PVM.InstrMeta, doneLabel string) {
	pc := instr.PC
	heapLimit := c.ctx.heapLimit
	t0 := PVMToX86[2]
	fail := fmt.Sprintf("sbrk_fail_%08x", uint32(pc))
	mprotect := fmt.Sprintf("sbrk_mprotect_%08x", uint32(pc))

	// Save t0 (PVM T0 / RBX) — we borrow it as scratch.
	a.MovRegToMem(RegGuestBase, regOffset(2), t0)

	// t0 = oldHP
	a.MovMemToReg(t0, RegGuestBase, -int32(OffsetHeapPointer))

	// RegScratch = amount (handle aReg == t0 clobber)
	if aReg == t0 {
		a.MovMemToReg(RegScratch, RegGuestBase, regOffset(2))
	} else {
		a.MovRegToReg(RegScratch, aReg)
	}

	// RegScratch = newHP = oldHP + amount
	a.AddRegReg(RegScratch, t0)

	// Overflow: newHP < oldHP → fail
	a.CmpRegReg(RegScratch, t0)
	a.Jcc(asm.CondB, fail)

	// Page boundary: pageCeil(oldHP) = (oldHP + 0xFFF) & ~0xFFF
	a.AddRegImm32(t0, 0xFFF)
	a.AndRegImm32(t0, ^int32(0xFFF))
	// t0 = pageCeil(oldHP), RegScratch = newHP

	// If newHP > pageCeil(oldHP) → need mprotect via Go
	a.CmpRegReg(RegScratch, t0)
	a.Jcc(asm.CondA, mprotect)

	// Limit check (only for inline path; mprotect path lets HandleSbrk check)
	a.MovImm64ToReg(t0, heapLimit)
	a.CmpRegReg(RegScratch, t0)
	a.Jcc(asm.CondA, fail)

	// Inline success: no page crossing, within limits.
	// Update heapPointer = newHP, rD = newHP.
	a.MovRegToMem(RegGuestBase, -int32(OffsetHeapPointer), RegScratch)
	a.MovMemToReg(t0, RegGuestBase, regOffset(2)) // restore t0
	a.MovRegToReg(dReg, RegScratch)
	a.Jmp(doneLabel)

	// Page crossing → exit to Go for mprotect.
	_ = a.BindLabel(mprotect)
	a.MovMemToReg(t0, RegGuestBase, regOffset(2)) // restore t0
	emitRuntimeExit(a, uint64(PVM.ExitHostCall)|uint64(SbrkCallID), fallthroughPC(instr))

	// Fail: overflow or heapLimit exceeded → rD = 0.
	_ = a.BindLabel(fail)
	a.MovMemToReg(t0, RegGuestBase, regOffset(2)) // restore t0
	a.MovImm64ToReg(dReg, 0)
	a.Jmp(doneLabel)
}

// opcode 102: count_set_bits_64 — Reg[rA] = popcnt(Reg[rB])
func (c *Compiler) emitCountSetBits64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.Popcnt(dReg, aReg)
	return nil
}

// opcode 103: count_set_bits_32 — Reg[rA] = popcnt(uint32(Reg[rB]))
func (c *Compiler) emitCountSetBits32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.Popcnt32(dReg, aReg)
	return nil
}

// opcode 104: leading_zero_bits_64 — Reg[rA] = lzcnt(Reg[rB])
func (c *Compiler) emitLeadingZeroBits64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.Lzcnt(dReg, aReg)
	return nil
}

// opcode 105: leading_zero_bits_32 — Reg[rA] = lzcnt(uint32(Reg[rB]))
func (c *Compiler) emitLeadingZeroBits32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.Lzcnt32(dReg, aReg)
	return nil
}

// opcode 106: trailing_zero_bits_64 — Reg[rA] = tzcnt(Reg[rB])
func (c *Compiler) emitTrailingZeroBits64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.Tzcnt(dReg, aReg)
	return nil
}

// opcode 107: trailing_zero_bits_32 — Reg[rA] = tzcnt(uint32(Reg[rB]))
func (c *Compiler) emitTrailingZeroBits32(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.Tzcnt32(dReg, aReg)
	return nil
}

// opcode 108: sign_extend_8 — Reg[rA] = sign_extend_8_to_64(Reg[rB])
func (c *Compiler) emitSignExtend8(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.MovFromByteToRegSx(dReg, aReg)
	return nil
}

// opcode 109: sign_extend_16 — Reg[rA] = sign_extend_16_to_64(Reg[rB])
func (c *Compiler) emitSignExtend16(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.MovFromWordToRegSx(dReg, aReg)
	return nil
}

// opcode 110: zero_extend_16 — Reg[rA] = zero_extend_16_to_64(Reg[rB])
func (c *Compiler) emitZeroExtend16(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	a.MovFromWordToRegZx(dReg, aReg)
	return nil
}

// opcode 111: reverse_bytes — Reg[rA] = bswap64(Reg[rB])
func (c *Compiler) emitReverseBytes(a *asm.Assembler, instr *PVM.InstrMeta) error {
	dReg, aReg := twoRegFromMeta(instr)
	if dReg != aReg {
		a.MovRegToReg(dReg, aReg)
	}
	a.Bswap(dReg)
	return nil
}
