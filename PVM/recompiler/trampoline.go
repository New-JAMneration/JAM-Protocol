//go:build linux && amd64

package recompiler

import (
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

// pvmRegSlot maps each PVM register index to a slot in the control-region array.
// RA (0) and SP (1) use slots 10–11 (R15-72 / R15-64) so division spill paths use disp8.
var pvmRegSlot = [PVMRegCount]int{10, 11, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 12}

// regOffset returns the negative displacement from R15 for PVM register index i.
func regOffset(i int) int32 {
	return -int32(OffsetRegisters) + int32(pvmRegSlot[i])*8
}

// EmitEntryTrampoline generates the Go → JIT entry trampoline.
//
// Calling convention (Go internal ABI, Go 1.17+):
//
//	RAX = guestBase (R15's value)
//	RBX = target native code address to jump to
//
// The trampoline:
//  1. Saves host callee-saved registers
//  2. Sets R15 = guestBase (from RAX)
//  3. Saves return address & RSP into control region (for signal handler)
//  4. Loads 13 PVM registers from control region
//  5. Jumps to target block (address in RBX, saved before overwrite)
//
// On return (via exit trampoline or signal handler jumping to "return_label"):
//  6. Restores host callee-saved registers
//  7. Returns to Go caller
func EmitEntryTrampoline(a *asm.Assembler) {
	// 1. Save host callee-saved registers
	a.Push(asm.RBX)
	a.Push(asm.RBP)
	a.Push(asm.R12)
	a.Push(asm.R13)
	a.Push(asm.R14)
	a.Push(asm.R15)

	// Save target address (RBX) to RCX before we overwrite RBX with PVM T0
	a.MovRegToReg(asm.RCX, asm.RBX)

	// 2. Set R15 = guestBase (passed in RAX)
	a.MovRegToReg(asm.R15, asm.RAX)

	// 3. Save return address and stack pointer for signal handler / exit trampoline
	a.LeaRIPRel(asm.RAX, "return_label")
	a.MovRegToMem(asm.R15, -int32(OffsetReturnAddr), asm.RAX)
	a.MovRegToMem(asm.R15, -int32(OffsetReturnStack), asm.RSP)

	// 4. Load 13 PVM registers from control region [R15 - OffsetRegisters + i*8]
	for i := 0; i < PVMRegCount; i++ {
		a.MovMemToReg(PVMToX86[i], asm.R15, regOffset(i))
	}

	// 5. Jump to target native code (address saved in RCX)
	a.JmpReg(asm.RCX)

	// return_label: exit trampoline or signal handler jumps here
	_ = a.BindLabel("return_label")

	// 6. Restore host callee-saved registers (reverse order)
	a.Pop(asm.R15)
	a.Pop(asm.R14)
	a.Pop(asm.R13)
	a.Pop(asm.R12)
	a.Pop(asm.RBP)
	a.Pop(asm.RBX)

	// 7. Return to Go caller
	a.Ret()
}

// EmitExitTrampoline generates the JIT → Go exit stub.
// This is emitted once and shared by all exit paths (halt, panic, OOG, host call).
// The caller must set ExitReason and ExitPC *before* jumping here.
//
// The trampoline:
//  1. Saves all 13 PVM registers back to the control region
//  2. Restores RSP from control region (ReturnStack)
//  3. Jumps to return_label (ReturnAddr in control region)
func EmitExitTrampoline(a *asm.Assembler) {
	_ = a.BindLabel("exit_trampoline")

	// 1. Store all 13 PVM registers back to control region
	for i := 0; i < PVMRegCount; i++ {
		a.MovRegToMem(asm.R15, regOffset(i), PVMToX86[i])
	}

	// 2. Restore host RSP and jump back to return_label in entry trampoline
	a.MovMemToReg(asm.RSP, asm.R15, -int32(OffsetReturnStack))
	a.JmpMem(asm.R15, -int32(OffsetReturnAddr))
}

// EmitExitWithReason emits a short stub that sets ExitReason to a constant
// and then jumps to the shared exit_trampoline.
// Used for halt, panic, out-of-gas exits.
func EmitExitWithReason(a *asm.Assembler, label string, reason int32) {
	_ = a.BindLabel(label)
	a.MovMemImm32(asm.R15, -int32(OffsetExitReason), reason)
	a.Jmp("exit_trampoline")
}

// EmitHostCallExit emits an inline exit sequence for a specific ecalli instruction.
// It stores ExitReason (encoded as host_call + callID), ExitPC (next PVM PC),
// then jumps to exit_trampoline.
//
// exitReason should be pre-encoded: (uint64(callID) << 8) | ExitHostCall
// nextPC is the PVM PC of the instruction after ecalli.
func EmitHostCallExit(a *asm.Assembler, exitReason int32, nextPC int32) {
	a.MovMemImm32(asm.R15, -int32(OffsetExitReason), exitReason)
	a.MovMemImm32_32(asm.R15, -int32(OffsetExitPC), nextPC)
	a.Jmp("exit_trampoline")
}
