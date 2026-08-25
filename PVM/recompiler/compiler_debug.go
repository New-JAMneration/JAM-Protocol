//go:build linux && amd64 && cgo && pvmtrace

package recompiler

import (
	"fmt"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
)

// CompileBlockInstruction compiles one pre-decoded instruction as a native block
// with A.7 block gas (same model as CompileBasicBlock). Used by
// DebugSingleStepInvoke: one instruction per native run, then trampoline back
// to Go for trace capture — not per-instruction (0.7.2) gas.
func (c *Compiler) CompileBlockInstruction(instr *PVM.InstrMeta) (*CompiledBlock, error) {
	a := c.asm
	a.Reset()
	c.memPanicPads = c.memPanicPads[:0]
	if err := c.ensureExitTrampoline(); err != nil {
		return nil, err
	}
	a.SetExitJmp(c.jmpExit)

	// Trace single-step must not chain blocks natively; each step returns to Go.
	c.singleStep = true
	c.linkFallthrough = nil
	c.linkTaken = nil

	pc := instr.PC
	fallthroughPC := fallthroughPC(instr)
	gasCost := c.blockGasCostAt(pc)

	oog := a.NewLabel()
	c.emitBlockGasCheck(a, oog, gasCost)

	handler := opcodeHandlers[instr.Opcode]
	if handler == nil {
		return nil, fmt.Errorf("unsupported opcode %d at PC=%d", instr.Opcode, pc)
	}
	if err := handler(c, a, instr); err != nil {
		return nil, fmt.Errorf("emit instruction at PC=%d: %w", pc, err)
	}

	// Reached only when the handler falls through (CONTINUE). Trap/ecalli/halt
	// jump to the trampoline inside the handler and never clear here.
	if PVM.IsBlockTerminator(instr.Opcode) {
		emitGasCharged(a, false)
	}
	emitExitToPC(a, fallthroughPC)
	c.emitDeferredMemPanicPads(a)
	emitBlockOutOfGasExit(a, oog, pc, gasCost)

	code, err := a.Finalize()
	if err != nil {
		return nil, fmt.Errorf("finalize block instruction at PC=%d: %w", pc, err)
	}

	em := c.ctx.executableMem
	if em == nil {
		return nil, fmt.Errorf("executable memory not initialized")
	}

	offset, err := em.Write(code)
	if err != nil {
		return nil, fmt.Errorf("write native code: %w", err)
	}

	return &CompiledBlock{
		PVMStartPC:   pc,
		PVMEndPC:     fallthroughPC,
		NativeAddr:   em.GetPtr(offset),
		NativeOffset: offset,
		NativeSize:   len(code),
		GasCost:      gasCost,
		InstrCount:   1,
	}, nil
}
