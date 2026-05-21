//go:build linux && amd64

package recompiler

import (
	"fmt"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

// Four inline instructions per PVM instruction (GP v0.7.2):
//   - MovMemToReg: load current gas
//   - TestRegReg + Jcc: branch to per-instruction OOG label if Gas < 0
//   - SubMemImm32: subtract 1 gas
// Block epilogue also emits one emitOutOfGasExit landing pad per instruction.
//
// If/when we switch to blockBased gas charging in GP v0.8.0, this helper
// should stop being called from the per-instruction compile loop.
func (c *Compiler) emitGasCheck(a *asm.Assembler, instrPC PVM.ProgramCounter) {
	// per-instruction (GP v0.7.2): load current gas before instrPC executes.
	a.MovMemToReg(RegScratch, RegGuestBase, -int32(OffsetGas))
	// per-instruction (GP v0.7.2): if Gas < 0, exit OOG at this instruction PC.
	a.TestRegReg(RegScratch, RegScratch)
	a.Jcc(asm.CondS, outOfGasLabel(instrPC))
	// per-instruction (GP v0.7.2): charge exactly 1 gas for the current instruction.
	a.SubMemImm32(RegGuestBase, -int32(OffsetGas), 1)
}

// emitOutOfGasExit emits the temporary GP v0.7.2 per-instruction OOG landing pad.
// Each instruction gets its own exit label so ExitPC matches interpreter semantics.
//
// TODO: when switch to blockBased gas charging in GP v0.8.0
// the per-instruction callers should be commented out and replaced by a single block-entry OOG exit.
func emitOutOfGasExit(a *asm.Assembler, instrPC PVM.ProgramCounter) {
	_ = a.BindLabel(outOfGasLabel(instrPC))
	// per-instruction (GP v0.7.2): report the exact instruction PC that failed.
	a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(instrPC))
	a.MovImm64ToReg(RegScratch, uint64(PVM.ExitOOG))
	a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
	a.Jmp("exit_trampoline")
}

func outOfGasLabel(instrPC PVM.ProgramCounter) string {
	return fmt.Sprintf("out_of_gas_pc_%08x", uint32(instrPC))
}

// emitBlockGasCheck is the prepared block-based gas charging path for GP v0.8.0.
// It is intentionally not called yet; `compiler.go` keeps the call sites commented
// out until the JIT path switches from current per-instruction semantics.
func (c *Compiler) emitBlockGasCheck(a *asm.Assembler, blockStartPC PVM.ProgramCounter, instrCount int64) {
	a.SubMemImm32(RegGuestBase, -int32(OffsetGas), int32(instrCount))
	a.Jcc(asm.CondS, blockOutOfGasLabel(blockStartPC))
}

// emitBlockOutOfGasExit is the prepared block-entry OOG landing pad for GP v0.8.0.
// It reports the block start PC, matching block-based charging semantics.
func emitBlockOutOfGasExit(a *asm.Assembler, blockStartPC PVM.ProgramCounter) {
	_ = a.BindLabel(blockOutOfGasLabel(blockStartPC))
	a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(blockStartPC))
	a.MovImm64ToReg(RegScratch, uint64(PVM.ExitOOG))
	a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
	a.Jmp("exit_trampoline")
}

func blockOutOfGasLabel(blockStartPC PVM.ProgramCounter) string {
	return fmt.Sprintf("out_of_gas_block_%08x", uint32(blockStartPC))
}

// emitExitWithReason64 emits an inline exit sequence that stores a 64-bit ExitReason
// then jumps to exit_trampoline. Used when the reason doesn't fit in imm32.
func emitExitWithReason64(a *asm.Assembler, label string, reason PVM.ExitReason) {
	_ = a.BindLabel(label)
	a.MovImm64ToReg(RegScratch, uint64(reason))
	a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
	a.Jmp("exit_trampoline")
}
