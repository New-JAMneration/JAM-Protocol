//go:build linux && amd64 && cgo

package recompiler

import (
	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

// emitBlockGasCheck charges once for the current basic block (A.9 gascostforblock
// baked in at compile time). Subtracts first; on OOG, emitBlockOutOfGasExit
// restores the cost so gas is left unchanged (matches interpreter A.4).
func (c *Compiler) emitBlockGasCheck(a *asm.Assembler, blockOOG asm.Label, blockGas int64) {
	charged := a.NewLabel()
	a.LoadByte(RegScratch, RegGuestBase, -int32(OffsetGasCharged))
	a.TestRegReg(RegScratch, RegScratch)
	a.Jcc(asm.CondNE, charged)

	a.SubMemImm32(RegGuestBase, -int32(OffsetGas), int32(blockGas))
	a.Jcc(asm.CondS, blockOOG)
	emitGasCharged(a, true)

	_ = a.BindLabel(charged)
}

func emitGasCharged(a *asm.Assembler, charged bool) {
	value := int32(0)
	if charged {
		value = 1
	}
	a.MovMemImm32_32(RegGuestBase, -int32(OffsetGasCharged), value)
}

// emitBlockOutOfGasExit emits the block-entry OOG landing pad.
// Restores the just-subtracted block cost, reports block start PC, then exits.
func emitBlockOutOfGasExit(a *asm.Assembler, blockOOG asm.Label, blockStartPC PVM.ProgramCounter, blockGas int64) {
	_ = a.BindLabel(blockOOG)
	// SUB of a negative imm adds the cost back (same pattern as legacy per-instr OOG).
	a.SubMemImm32(RegGuestBase, -int32(OffsetGas), -int32(blockGas))
	a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(blockStartPC))
	a.MovImm64ToReg(RegScratch, uint64(PVM.ExitOOG))
	a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
	a.JmpExit()
}
