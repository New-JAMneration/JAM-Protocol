//go:build linux && amd64 && cgo

package recompiler

import (
	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

// Two inline instructions per PVM instruction (GP v0.7.2), fused charge+check:
//   - SubMemImm32: charge 1 gas
//   - Jcc(S): branch to the OOG landing pad when the result went negative
//
// (GP A.6) OOG when pre-charge Gas < 1. Gas is never negative at entry, so
// post-charge < 0 (SF=1) ⟺ pre-charge <= 0 ⟺ pre-charge < 1 — the same
// condition the interpreter checks, so a gas-exhausted program stops at the
// same instruction on both backends. The landing pad un-charges the 1 so the
// reported remaining gas also matches (the interpreter never charges on OOG).
//
// oog is the instruction's landing-pad label, allocated by the compile loop and
// bound later by emitOutOfGasExit (see compileBasicBlockAtDepth's oogLabels).
//
// If/when we switch to blockBased gas charging in GP v0.8.0, this helper
// should stop being called from the per-instruction compile loop.
func (c *Compiler) emitGasCheck(a *asm.Assembler, oog asm.Label) {
	a.SubMemImm32(RegGuestBase, -int32(OffsetGas), 1)
	a.Jcc(asm.CondS, oog)
}

// emitOutOfGasExit emits the temporary GP v0.7.2 per-instruction OOG landing pad.
// Each instruction gets its own exit label so ExitPC matches interpreter semantics.
//
// TODO: when switch to blockBased gas charging in GP v0.8.0
// the per-instruction callers should be commented out and replaced by a single block-entry OOG exit.
func emitOutOfGasExit(a *asm.Assembler, oog asm.Label, instrPC PVM.ProgramCounter) {
	_ = a.BindLabel(oog)
	// Undo the fused charge: on OOG the interpreter leaves gas unchanged.
	a.SubMemImm32(RegGuestBase, -int32(OffsetGas), -1)
	// per-instruction (GP v0.7.2): report the exact instruction PC that failed.
	a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(instrPC))
	a.MovImm64ToReg(RegScratch, uint64(PVM.ExitOOG))
	a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
	a.Jmp(a.ExitTrampoline())
}

// emitBlockGasCheck is the prepared block-based gas charging path for GP v0.8.0.
// It is intentionally not called yet; `compiler.go` keeps the call sites commented
// out until the JIT path switches from current per-instruction semantics.
func (c *Compiler) emitBlockGasCheck(a *asm.Assembler, blockOOG asm.Label, instrCount int64) {
	a.SubMemImm32(RegGuestBase, -int32(OffsetGas), int32(instrCount))
	a.Jcc(asm.CondS, blockOOG)
}

// emitBlockOutOfGasExit is the prepared block-entry OOG landing pad for GP v0.8.0.
// It reports the block start PC, matching block-based charging semantics.
func emitBlockOutOfGasExit(a *asm.Assembler, blockOOG asm.Label, blockStartPC PVM.ProgramCounter) {
	_ = a.BindLabel(blockOOG)
	a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(blockStartPC))
	a.MovImm64ToReg(RegScratch, uint64(PVM.ExitOOG))
	a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
	a.Jmp(a.ExitTrampoline())
}
