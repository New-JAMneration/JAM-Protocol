//go:build pvmtrace

package interpreter

import PVM "github.com/New-JAMneration/JAM-Protocol/PVM"

// MachineInvoke: trace → DebugSingleStepInvoke; else BlockBasedInvokeDecodedBlocks.
func (h *Host) MachineInvoke(pc PVM.ProgramCounter) (PVM.ExitReason, PVM.ProgramCounter) {
	if h.Interpreter.Trace != nil {
		return h.Interpreter.DebugSingleStepInvoke(pc)
	}
	return h.Interpreter.BlockBasedInvokeDecodedBlocks(pc)
}
