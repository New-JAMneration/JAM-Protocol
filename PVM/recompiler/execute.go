//go:build linux && amd64 && cgo

package recompiler

import (
	"runtime"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
	x86_signal_linux "github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/x86signal"
)

// ExecuteBlock runs a single compiled basic block through the JIT entry/exit
// trampoline and returns the ExitReason. The caller is responsible for
// interpreting the exit reason (halt, panic, page fault, host call, OOG)
// and driving the execution loop accordingly.
//
// This function:
//  1. Locks the current goroutine to its OS thread (required for TLS)
//  2. Sets the per-thread guest base pointer for the signal handler
//  3. Builds and caches the entry trampoline
//  4. Calls into native code via callNative
//  5. Reads the exit reason from the control region
func ExecuteBlock(ctx *JITContext, block *CompiledBlock) PVM.ExitReason {
	x86_signal_linux.SetupSignalHandler()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return executeBlockLocked(ctx, block)
}

// executeBlockLocked runs one compiled block without locking the OS thread.
// Callers (BlockBasedInvoke, DebugSingleStepInvoke) must hold LockOSThread
// for the entire native execution loop; tests and standalone callers use ExecuteBlock.
func executeBlockLocked(ctx *JITContext, block *CompiledBlock) PVM.ExitReason {
	if jitProfile {
		jm.roundTrips.Add(1)
	}
	trampolineAddr, err := ctx.getTrampolineAddr()
	if err != nil {
		ctx.WriteExitReason(PVM.ExitPanic)
		return PVM.ExitPanic
	}

	em := ctx.executableMem
	codeStart := em.GetPtr(0)
	// Use the arena capacity, not Used(): a shared cached arena can be grown by
	// another goroutine compiling concurrently, so a Used()-based upper bound could
	// exclude newly written code and misclassify a fault in it.
	codeEnd := codeStart + uintptr(em.Size())
	x86_signal_linux.SetFaultWindow(ctx.GuestBase(), codeStart, codeEnd)
	defer x86_signal_linux.ClearFaultWindow()

	callNative(ctx.GuestBase(), block.NativeAddr, trampolineAddr)

	return ctx.ReadExitReason()
}

// getTrampolineAddr returns the cached entry trampoline address,
// emitting it into executable memory on first call.
func (ctx *JITContext) getTrampolineAddr() (uintptr, error) {
	if ctx.trampolineAddr != 0 {
		return ctx.trampolineAddr, nil
	}

	em := ctx.executableMem
	if em == nil {
		return 0, errNoExecMem
	}

	a := asm.NewAssembler()
	EmitEntryTrampoline(a)
	code, err := a.Finalize()
	if err != nil {
		return 0, err
	}

	offset, err := em.Write(code)
	if err != nil {
		return 0, err
	}

	ctx.trampolineAddr = em.GetPtr(offset)
	return ctx.trampolineAddr, nil
}

type execError string

func (e execError) Error() string { return string(e) }

const errNoExecMem = execError("executable memory not initialized")


// DjumpCallID is the sentinel host-call ID for indirect jumps (jump_ind, load_imm_jump_ind).
const DjumpCallID = 0xFE

// IsDjumpExit returns true if the exit reason is an indirect jump request.
func IsDjumpExit(reason PVM.ExitReason) bool {
	return reason.GetReasonType() == PVM.HOST_CALL && reason.GetHostCallID() == DjumpCallID
}
