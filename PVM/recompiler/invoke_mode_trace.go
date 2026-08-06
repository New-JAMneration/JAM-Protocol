//go:build linux && amd64 && cgo && trace

package recompiler

import PVM "github.com/New-JAMneration/JAM-Protocol/PVM"

// MachineInvoke: trace/debug → DebugSingleStepInvoke; else BlockBasedInvoke.
func (r *Recompiler) MachineInvoke(pc PVM.ProgramCounter) (PVM.ExitReason, PVM.ProgramCounter) {
	if r.Trace != nil {
		return r.DebugSingleStepInvoke(pc)
	}
	if PVM.RecompilerDebugModeRuntime == PVM.RecompilerDebugSingleStep {
		return r.DebugSingleStepInvoke(pc)
	}
	return r.BlockBasedInvoke(pc)
}
