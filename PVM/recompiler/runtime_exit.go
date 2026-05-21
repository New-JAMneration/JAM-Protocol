//go:build linux && amd64

package recompiler

import (
	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

// sbrkOpcode is opcode 101 (heap expansion). Expand paths exit to Go mid-block;
// resume must compile a suffix from fallthroughPC, not re-run the block head.
const sbrkOpcode = 101

// emitRuntimeExit stores ExitReason and ExitPC (the PVM PC to resume at) and
// jumps to the shared exit trampoline. Matches ecalli: exitPC is fallthrough,
// not the exiting instruction's PC.
func emitRuntimeExit(a *asm.Assembler, exitReason uint64, exitPC PVM.ProgramCounter) {
	a.MovImm64ToReg(RegScratch, exitReason)
	a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
	a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(exitPC))
	a.Jmp("exit_trampoline")
}
