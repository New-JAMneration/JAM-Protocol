//go:build linux && amd64 && cgo && !pvmtrace

package recompiler

import PVM "github.com/New-JAMneration/JAM-Protocol/PVM"

// MachineInvoke: run until non-CONTINUE exit. Forwards to BlockBasedInvoke
// (pre-decoded blocks → native JIT; symmetrical to interpreter
// BlockBasedInvokeDecodedBlocks). Trace routing in invoke_mode_trace.go.
func (r *Recompiler) MachineInvoke(pc PVM.ProgramCounter) (PVM.ExitReason, PVM.ProgramCounter) {
	return r.BlockBasedInvoke(pc)
}
