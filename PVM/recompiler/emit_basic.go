//go:build linux && amd64

package recompiler

import (
	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

// ---- 4.3 No-argument instructions ----

// opcode 0: trap — set ExitReason = PANIC, ExitPC = pc, exit
func (c *Compiler) emitTrap(a *asm.Assembler, instr *PVM.InstrMeta) error {
	pc := instr.PC
	a.MovImm64ToReg(RegScratch, uint64(PVM.ExitPanic))
	a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
	a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(pc))
	a.Jmp("exit_trampoline")
	return nil
}

// opcode 1: fallthrough — no-op
func (c *Compiler) emitFallthrough(a *asm.Assembler, instr *PVM.InstrMeta) error {
	_ = instr
	a.Nop()
	return nil
}

// ---- 4.4 Immediate instructions ----

// opcode 10: ecalli — host call exit
func (c *Compiler) emitEcalli(a *asm.Assembler, instr *PVM.InstrMeta) error {
	callID := int(instr.Imm[0])
	nextPC := fallthroughPC(instr)
	exitReason := PVM.ExitHostCall | PVM.ExitReason(uint8(callID))

	a.MovImm64ToReg(RegScratch, uint64(exitReason))
	a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
	a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(nextPC))
	a.Jmp("exit_trampoline")
	return nil
}

// opcode 20: load_imm_64 — Reg[rA] = 64-bit immediate
func (c *Compiler) emitLoadImm64(a *asm.Assembler, instr *PVM.InstrMeta) error {
	xReg := PVMReg(instr.Dst)
	a.MovImm64ToReg(xReg, instr.Imm[0])
	return nil
}

// opcode 51: load_imm — Reg[rA] = sign_extended(vX)
func (c *Compiler) emitLoadImm(a *asm.Assembler, instr *PVM.InstrMeta) error {
	xReg, vX := oneRegImmFromMeta(instr)
	if fitsInt32(vX) {
		a.MovImm32ToReg(xReg, int32(int64(vX)))
	} else {
		a.MovImm64ToReg(xReg, vX)
	}
	return nil
}

func fitsInt32(v uint64) bool {
	sv := int64(v)
	return sv >= -2147483648 && sv <= 2147483647
}

// emitAddUint64ToReg adds a uint64 immediate to dst (64-bit add).
func emitAddUint64ToReg(a *asm.Assembler, dst asm.Register, addend uint64) {
	if addend == 0 {
		return
	}
	if fitsInt32(addend) {
		a.AddRegImm32(dst, int32(int64(addend)))
		return
	}
	spillReg := PVMToX86[2]
	a.MovRegToMem(RegGuestBase, regOffset(2), spillReg)
	a.MovImm64ToReg(spillReg, addend)
	a.AddRegReg(dst, spillReg)
	a.MovMemToReg(spillReg, RegGuestBase, regOffset(2))
}

// emitZeroExt32 clears the upper 32 bits (e.g. before 32-bit DIV operands).
func emitZeroExt32(a *asm.Assembler, reg asm.Register) {
	a.ZeroExtend32(reg)
}

// emitSignExt32 sign-extends the low 32 bits of reg to 64 bits.
func emitSignExt32(a *asm.Assembler, reg asm.Register) {
	a.MovFromDwordToRegSx(reg, reg)
}
