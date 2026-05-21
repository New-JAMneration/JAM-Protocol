//go:build linux && amd64 && trace

package recompiler

import (
	"fmt"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
)

// CompileSingleInstruction compiles a single PVM instruction as a debug block.
// After execution, native code trampolines back to Go with ExitPC set to the
// fallthrough PC (for non-branch instructions) or the computed target PC (for branches).
//
// This enables debug single-step mode where each instruction acts as its own block,
// allowing per-instruction trace capture from the recompiler backend.
func (c *Compiler) CompileSingleInstruction(instr *PVM.InstrMeta) (*CompiledBlock, error) {
	a := c.asm
	a.Reset()

	pc := instr.PC
	fallthroughPC := fallthroughPC(instr)

	c.emitGasCheck(a, instr.PC)

	handler := opcodeHandlers[instr.Opcode]
	if handler == nil {
		return nil, fmt.Errorf("unsupported opcode %d at PC=%d", instr.Opcode, pc)
	}
	if err := handler(c, a, instr); err != nil {
		return nil, fmt.Errorf("emit instruction at PC=%d: %w", pc, err)
	}

	emitExitToPC(a, fallthroughPC)
	EmitExitTrampoline(a)
	emitOutOfGasExit(a, instr.PC)

	code, err := a.Finalize()
	if err != nil {
		return nil, fmt.Errorf("finalize single instruction at PC=%d: %w", pc, err)
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
		GasCost:      1,
		InstrCount:   1,
	}, nil
}
