//go:build linux && amd64

package recompiler

import (
	"fmt"
	"runtime"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/PVMtrace"
	x86_signal_linux "github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/x86signal"
)

// Recompiler is the machine layer of the JIT backend, symmetrical to
// PVM.Interpreter on the interpreter backend. It owns compilation and
// native execution over a JITContext, but not host-call dispatch state.
// Host-call orchestration (OOG, HALT, PANIC handling, sbrk, Omega
// dispatch) lives in host, which drives Recompiler.BlockBasedInvoke in
// a loop.
type Recompiler struct {
	compiler *Compiler
	program  *PVM.Program
	ctx      *JITContext
	Trace    *PVMtrace.Trace
}

// NewRecompiler creates a fresh Recompiler bound to the given program
// and JIT context. The caller owns the ctx lifetime (Close/Release).
func NewRecompiler(program *PVM.Program, ctx *JITContext) *Recompiler {
	cache := NewCodeCache()
	return &Recompiler{
		compiler: NewCompiler(program, ctx, cache),
		program:  program,
		ctx:      ctx,
	}
}

// BlockBasedInvoke runs one or more compiled basic blocks until the
// native side signals a non-CONTINUE exit.
func (r *Recompiler) BlockBasedInvoke(pc PVM.ProgramCounter) (PVM.ExitReason, PVM.ProgramCounter) {
	x86_signal_linux.SetupSignalHandler()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if jitProfile {
		jm.lockCalls.Add(1)
	}

	for {
		blockStartPC := pc
		block, err := r.lookupOrCompileBlock(pc)
		if err != nil {
			return PVM.ExitPanic, 0
		}

		r.ctx.WriteExitPC(pc)
		r.ctx.WriteExitReason(PVM.ExitContinue)

		if r.Trace != nil {
			r.Trace.RecordBlockEntry(uint64(pc), int64(r.ctx.ReadGas()))
		}

		exitReason := executeBlockLocked(r.ctx, block)
		exitPC := r.ctx.ReadExitPC()

		if exitReason.GetReasonType() == PVM.HOST_CALL && exitReason.GetHostCallID() == DjumpCallID {
			exitReason, pc = r.resolveDjump(blockStartPC, uint32(exitPC))
			switch exitReason.GetReasonType() {
			case PVM.CONTINUE:
				continue
			case PVM.HALT, PVM.PANIC:
				return exitReason, djumpInstrPC(r.program, blockStartPC)
			default:
				return exitReason, 0
			}
		}

		if IsSbrkExit(exitReason) {
			instr, ok := r.sbrkInstrForRuntimeExit(exitPC)
			if !ok {
				return PVM.ExitPanic, 0
			}
			exitReason, pc = r.resolveSbrk(instr)
			switch exitReason.GetReasonType() {
			case PVM.CONTINUE:
				continue
			case PVM.HALT, PVM.PANIC:
				return exitReason, r.ctx.ReadExitPC()
			default:
				return exitReason, exitPC
			}
		}

		switch exitReason.GetReasonType() {
		case PVM.CONTINUE:
			pc = exitPC
			continue
		case PVM.HALT, PVM.PANIC:
			return exitReason, r.ctx.ReadExitPC()
		case PVM.OUT_OF_GAS, PVM.PAGE_FAULT, PVM.HOST_CALL:
			return exitReason, exitPC
		default:
			return PVM.ExitPanic, 0
		}
	}
}

func (r *Recompiler) resolveDjump(blockStartPC PVM.ProgramCounter, jumpAddr uint32) (PVM.ExitReason, PVM.ProgramCounter) {
	if jitProfile {
		jm.djumpResolves.Add(1)
	}
	instrPC := djumpInstrPC(r.program, blockStartPC)
	return PVM.DjumpResolve(instrPC, jumpAddr, r.program.JumpTable, r.program.Bitmasks)
}

// djumpInstrPC returns the PC of the terminating djump instruction in the block.
// jump_ind and load_imm_jump_ind are block terminators, so EndPC is the instruction PC.
func djumpInstrPC(program *PVM.Program, blockStartPC PVM.ProgramCounter) PVM.ProgramCounter {
	if blockMeta := program.LookupBlock(blockStartPC); blockMeta != nil {
		return blockMeta.EndPC
	}
	return blockStartPC
}

func (r *Recompiler) Program() *PVM.Program {
	return r.program
}

func (r *Recompiler) Ctx() *JITContext {
	return r.ctx
}

func (r *Recompiler) lookupOrCompileBlock(pc PVM.ProgramCounter) (*CompiledBlock, error) {
	if block := r.compiler.cache.Get(pc); block != nil {
		return block, nil
	}

	block, err := r.compiler.CompileBasicBlock(pc)
	if err != nil {
		return nil, fmt.Errorf("compile block at PC=%d: %w", pc, err)
	}
	r.compiler.cache.Put(block)
	return block, nil
}
