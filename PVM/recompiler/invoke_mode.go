//go:build linux && amd64 && !trace

package recompiler

import PVM "github.com/New-JAMneration/JAM-Protocol/PVM"

// MachineInvoke runs native PVM execution until a non-CONTINUE exit.
func (r *Recompiler) MachineInvoke(pc PVM.ProgramCounter) (PVM.ExitReason, PVM.ProgramCounter) {
	return r.BlockBasedInvoke(pc)
}
