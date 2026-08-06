//go:build linux && amd64 && cgo

package recompiler

import (
	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

// emitRuntimeExit stores ExitReason and ExitPC (the PVM PC to resume at) and
// jumps to the shared exit trampoline. Matches ecalli: exitPC is fallthrough,
// not the exiting instruction's PC.
func emitRuntimeExit(a *asm.Assembler, exitReason uint64, exitPC PVM.ProgramCounter) {
	a.MovImm64ToReg(RegScratch, exitReason)
	a.MovRegToMem(RegGuestBase, -int32(OffsetExitReason), RegScratch)
	a.MovMemImm32_32(RegGuestBase, -int32(OffsetExitPC), int32(exitPC))
	a.Jmp(a.ExitTrampoline())
}
