//go:build !pvmtrace

package interpreter

import PVM "github.com/New-JAMneration/JAM-Protocol/PVM"

// MachineInvoke: run until non-CONTINUE exit. Forwards to
// BlockBasedInvokeDecodedBlocks (pre-decoded Go path). Trace in invoke_mode_trace.go.
func (h *Host) MachineInvoke(pc PVM.ProgramCounter) (PVM.ExitReason, PVM.ProgramCounter) {
	return h.Interpreter.BlockBasedInvokeDecodedBlocks(pc)
}
