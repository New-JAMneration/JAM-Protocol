//go:build linux && amd64

package recompiler

import (
	"fmt"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

// emitJmpNativeAddr jumps to an absolute native code address (block linking).
func emitJmpNativeAddr(a *asm.Assembler, addr uintptr) {
	a.MovImm64ToReg(RegScratch, uint64(addr))
	a.JmpReg(RegScratch)
}

// emitFallthroughEpilogue emits block epilogue: native JMP when linkTarget is
// known, otherwise emitExitToPC back to the Go dispatcher.
func emitFallthroughEpilogue(a *asm.Assembler, fallthroughPC PVM.ProgramCounter, linkTarget *CompiledBlock) {
	if linkTarget != nil {
		emitJmpNativeAddr(a, linkTarget.NativeAddr)
		return
	}
	emitExitToPC(a, fallthroughPC)
}

// emitLinkOrExit emits a native JMP directly into a pre-compiled target block
// (block linking) when link is non-nil and its start PC matches targetPC;
// otherwise it falls back to emitExitToPC (one Go-dispatcher round-trip). The
// PVMStartPC guard makes a mismatched/stale link safe — at worst we don't link.
//
// Linked blocks keep the PVM registers live in x86 registers (the entry
// trampoline loads them once per segment), so jumping straight to a block's
// NativeAddr — which begins with its gas check + first instruction, not a
// register reload — is exactly what the Go dispatcher would have run, minus the
// trampoline round-trip.
func emitLinkOrExit(a *asm.Assembler, link *CompiledBlock, targetPC PVM.ProgramCounter) {
	if link != nil && link.PVMStartPC == targetPC {
		emitJmpNativeAddr(a, link.NativeAddr)
		return
	}
	emitExitToPC(a, targetPC)
}

// staticBranchTarget returns the compile-time-known jump/branch target PC of a
// terminator instruction, plus whether the instruction has one. Indirect jumps
// (jump_ind / load_imm_jump_ind) compute their target at runtime and are NOT
// included — they are handled by native djump dispatch (djump_native.go).
//
// Target immediate index matches the per-opcode decode helpers:
//   - jump (40):                Imm[0]
//   - load_imm_jump (80) /
//     branch_*_imm (81–90):     Imm[1]   (Imm[0] is the compared value)
//   - branch_* (170–175):       Imm[0]
func staticBranchTarget(instr *PVM.InstrMeta) (PVM.ProgramCounter, bool) {
	switch instr.Opcode {
	case 40:
		return PVM.ProgramCounter(instr.Imm[0]), true
	case 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90:
		return PVM.ProgramCounter(instr.Imm[1]), true
	case 170, 171, 172, 173, 174, 175:
		return PVM.ProgramCounter(instr.Imm[0]), true
	default:
		return 0, false
	}
}

// compileForLink links only to an already-compiled block; uncompiled targets
// return an error and the caller falls back to emitExitToPC (lazy compile via
// the Go dispatcher). We no longer eagerly pre-compile cold successors: JIT
// profiling showed codegen dominates (~93% of run) and eager pre-compile built
// ~38% more blocks than ever ran.
func (c *Compiler) compileForLink(pc PVM.ProgramCounter, _ int) (*CompiledBlock, error) {
	if block := c.cache.Get(pc); block != nil {
		return block, nil
	}
	return nil, fmt.Errorf("link target PC=%d not yet compiled", pc)
}
