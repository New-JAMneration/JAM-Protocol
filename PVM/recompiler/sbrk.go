//go:build linux && amd64

package recompiler

import PVM "github.com/New-JAMneration/JAM-Protocol/PVM"

// resolveSbrk runs HandleSbrk for the instruction at instr and returns the
// fallthrough PC on success.
func (r *Recompiler) resolveSbrk(instr *PVM.InstrMeta) (PVM.ExitReason, PVM.ProgramCounter) {
	exitReason := HandleSbrk(r.ctx, instr.Dst, instr.Src[0])
	if exitReason != PVM.ExitContinue {
		return exitReason, 0
	}
	return PVM.ExitContinue, fallthroughPC(instr)
}

// sbrkInstrForRuntimeExit resolves the sbrk InstrMeta from an expand exit's ExitPC.
// ExitPC is fallthrough (like ecalli). If exitPC still points at the sbrk opcode
// (legacy emit), that path is accepted too.
func (r *Recompiler) sbrkInstrForRuntimeExit(exitPC PVM.ProgramCounter) (*PVM.InstrMeta, bool) {
	idx := r.program.InstrIdxAt[exitPC]
	if idx < 0 {
		return nil, false
	}
	instr := &r.program.Instrs[int(idx)]
	if instr.Opcode == sbrkOpcode && instr.PC == exitPC {
		return instr, true
	}
	if int(idx) > 0 {
		prev := &r.program.Instrs[int(idx)-1]
		if prev.Opcode == sbrkOpcode && fallthroughPC(prev) == exitPC {
			return prev, true
		}
	}
	return nil, false
}
