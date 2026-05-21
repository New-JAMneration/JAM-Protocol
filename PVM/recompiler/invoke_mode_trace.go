//go:build linux && amd64 && trace

package recompiler

import PVM "github.com/New-JAMneration/JAM-Protocol/PVM"

// MachineInvoke runs native PVM execution. When trace is active, uses debug single-step
// mode to produce per-instruction streams aligned with the interpreter trace.
func (r *Recompiler) MachineInvoke(pc PVM.ProgramCounter) (PVM.ExitReason, PVM.ProgramCounter) {
	if r.Trace != nil {
		return r.DebugSingleStepInvoke(pc)
	}
	if PVM.RecompilerDebugModeRuntime == PVM.RecompilerDebugSingleStep {
		return r.DebugSingleStepInvoke(pc)
	}
	return r.BlockBasedInvoke(pc)
}
