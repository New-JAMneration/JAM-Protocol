//go:build linux && amd64 && cgo

package recompiler

import (
	"context"
	"runtime/pprof"
	"time"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/PVMtrace"
	x86_signal_linux "github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/x86signal"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func Psi_M_recompiler(
	code PVM.StandardCodeFormat,
	counter PVM.ProgramCounter,
	gas types.Gas,
	argument PVM.Argument,
	omegas PVM.Omegas,
	addition PVM.HostCallArgs,
) PVM.Psi_M_ReturnType {
	// Tag goroutine "phase:pvm" for `pprof -tagfocus=phase:pvm` (PVM-only view).
	if jitProfile {
		pprof.SetGoroutineLabels(pprof.WithLabels(context.Background(), pprof.Labels("phase", "pvm")))
		defer pprof.SetGoroutineLabels(context.Background())
	}

	var tSetup time.Time
	if jitProfile {
		jm.invokes.Add(1)
		tSetup = time.Now()
		tInvoke := tSetup
		defer func() { jm.invokeNanos.Add(int64(time.Since(tInvoke))) }()
	}

	ctx, err := NewJITContext()
	if err != nil {
		return jitPanicResult(addition)
	}
	defer ctx.Close()

	programCode, registers, err := ctx.InitFromProgram(code, argument)
	if err != nil {
		return jitPanicResult(addition)
	}
	if jitProfile {
		jm.setupNanos.Add(int64(time.Since(tSetup)))
	}

	var tDeblob time.Time
	if jitProfile {
		tDeblob = time.Now()
	}
	program, exitReason := PVM.GetOrDeblobProgram(addition.CodeHash, programCode)
	if jitProfile {
		jm.deblobNanos.Add(int64(time.Since(tDeblob)))
	}
	if exitReason != PVM.ExitContinue {
		return jitPanicResult(addition)
	}
	addition.Program = program

	// Cross-invocation cache: reuse the compiled native code arena for this
	// CodeHash. Uncached (zero-hash) artifacts are throwaway — close after use.
	cp, cached, err := acquireCompiledProgram(addition.CodeHash, program)
	if err != nil {
		return jitPanicResult(addition)
	}
	if cached {
		// Hold a reference for this whole invocation so eviction can't Munmap the
		// arena while we execute in it; release when done.
		defer releaseCompiledProgram(cp)
	} else {
		defer cp.close() // zero-hash throwaway: free its arena
	}
	cp.bindContext(ctx)

	ctx.WriteRegisters(registers)
	ctx.WriteGas(PVM.Gas(gas))
	ctx.WriteExitReason(PVM.ExitContinue)
	ctx.WriteExitPC(counter)

	var trace *PVMtrace.Trace
	if addition.AccumulateTrace != nil {
		at := addition.AccumulateTrace
		trace = PVMtrace.NewTrace(
			uint32(at.ServiceID),
			[32]byte(at.CodeHash),
			uint64(at.Timeslot),
			programCode,
			PVMtrace.InitialState{PC: uint64(counter), Gas: int64(gas), Regs: registers},
			"accumulate",
			PVMtrace.BackendRecompiler,
		)
	}
	var psiHResult PVM.Psi_H_ReturnType
	defer func() {
		if trace == nil {
			return
		}
		var regs [13]uint64
		var g int64
		if psiHResult.VM != nil && psiHResult.VM.Registers != nil {
			copy(regs[:], (*psiHResult.VM.Registers)[:])
		}
		if psiHResult.VM != nil && psiHResult.VM.Gas != nil {
			g = int64(*psiHResult.VM.Gas)
		}
		_ = trace.Close(PVMtrace.FinalState{
			PC:         uint64(psiHResult.Counter),
			Gas:        g,
			ExitReason: psiHResult.ExitReason.String(),
			Regs:       regs,
		})
	}()

	r := newRecompiler(cp, ctx)
	r.Trace = trace
	host := newHost(r, addition, omegas)
	var tRun time.Time
	if jitProfile {
		tRun = time.Now()
	}
	psiHResult = host.HostCall(counter)
	if jitProfile {
		jm.runNanos.Add(int64(time.Since(tRun)))
	}

	g, v, a := PVM.R(gas, psiHResult)
	return PVM.Psi_M_ReturnType{
		Gas:           types.Gas(g),
		ReasonOrBytes: v,
		Addition:      a,
	}
}

func jitPanicResult(addition PVM.HostCallArgs) PVM.Psi_M_ReturnType {
	return PVM.Psi_M_ReturnType{
		Gas:           0,
		ReasonOrBytes: PVM.ExitPanic,
		Addition:      addition,
	}
}

func init() {
	x86_signal_linux.SetupSignalHandler()
	PVM.Psi_M_recompilerHook = Psi_M_recompiler
}
