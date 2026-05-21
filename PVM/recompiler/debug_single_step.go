//go:build linux && amd64 && trace

package recompiler

import (
	"runtime"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	x86_signal_linux "github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/x86signal"
)

// DebugSingleStepInvoke runs the recompiler in debug single-step mode.
// Each PVM instruction is compiled and executed individually, with a trampoline
// back to Go after every instruction. This produces per-instruction trace streams
// identical in semantics to the interpreter trace.
//
// This mode does NOT online-compare with the interpreter. It only records
// recompiler dynamic streams for offline comparison via pvm-diff.
func (r *Recompiler) DebugSingleStepInvoke(pc PVM.ProgramCounter) (PVM.ExitReason, PVM.ProgramCounter) {
	x86_signal_linux.SetupSignalHandler()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	trace := r.Trace

	for {
		idx := r.program.InstrIdxAt[pc]
		if idx < 0 {
			return PVM.ExitPanic, 0
		}
		instr := &r.program.Instrs[int(idx)]

		block, err := r.compiler.CompileSingleInstruction(instr)
		if err != nil {
			return PVM.ExitPanic, 0
		}

		var src1Val, src2Val uint64
		if instr.Src[0] != 0xff {
			src1Val = r.ctx.ReadRegister(instr.Src[0])
		}
		if instr.Src[1] != 0xff {
			src2Val = r.ctx.ReadRegister(instr.Src[1])
		}

		r.ctx.ClearMemAccess()
		r.ctx.WriteExitPC(pc)
		r.ctx.WriteExitReason(PVM.ExitContinue)

		exitReason := executeBlockLocked(r.ctx, block)
		exitPC := r.ctx.ReadExitPC()

		if IsSbrkExit(exitReason) {
			exitReason = HandleSbrk(r.ctx, instr.Dst, instr.Src[0])
			if exitReason != PVM.ExitContinue {
				if trace != nil {
					var dstVal, src1Val, src2Val uint64
					if instr.Dst != 0xff {
						dstVal = r.ctx.ReadRegister(instr.Dst)
					}
					if instr.Src[0] != 0xff {
						src1Val = r.ctx.ReadRegister(instr.Src[0])
					}
					if instr.Src[1] != 0xff {
						src2Val = r.ctx.ReadRegister(instr.Src[1])
					}
					trace.RecordStep(
						uint32(instr.PC), instr.Opcode,
						instr.Dst, instr.Src[0], instr.Src[1],
						dstVal, src1Val, src2Val,
						int64(r.ctx.ReadGas()),
						0, 0, 0, 0,
					)
				}
				switch exitReason.GetReasonType() {
				case PVM.HALT, PVM.PANIC:
					return exitReason, r.ctx.ReadExitPC()
				default:
					return exitReason, exitPC
				}
			}
			exitPC = fallthroughPC(instr)
			exitReason = PVM.ExitContinue
		}

		if IsDjumpExit(exitReason) {
			exitReason, exitPC = r.resolveDjump(instr.PC, uint32(exitPC))
			if exitReason != PVM.ExitContinue {
				if trace != nil {
					var dstVal uint64
					if instr.Dst != 0xff {
						dstVal = r.ctx.ReadRegister(instr.Dst)
					}
					trace.RecordStep(
						uint32(instr.PC), instr.Opcode,
						instr.Dst, instr.Src[0], instr.Src[1],
						dstVal, src1Val, src2Val,
						int64(r.ctx.ReadGas()),
						0, 0, 0, 0,
					)
				}
				switch exitReason.GetReasonType() {
				case PVM.HALT, PVM.PANIC:
					return exitReason, instr.PC
				default:
					return exitReason, exitPC
				}
			}
		}

		var dstVal uint64
		if instr.Dst != 0xff {
			dstVal = r.ctx.ReadRegister(instr.Dst)
		}
		gas := r.ctx.ReadGas()

		var loadAddr, storeAddr uint32
		var loadVal, storeVal uint64
		if r.ctx.HasMemAccess() {
			addr, _ := r.ctx.ReadMemAccess()
			loadAddr, loadVal, storeAddr, storeVal = traceRecordedMemAccess(r.ctx, instr.Opcode, addr)
		}

		if trace != nil {
			trace.RecordStep(
				uint32(instr.PC), instr.Opcode,
				instr.Dst, instr.Src[0], instr.Src[1],
				dstVal, src1Val, src2Val,
				int64(gas),
				loadAddr, loadVal,
				storeAddr, storeVal,
			)
		}

		switch exitReason.GetReasonType() {
		case PVM.CONTINUE:
			pc = exitPC
			continue
		case PVM.HALT, PVM.PANIC:
			return exitReason, r.ctx.ReadExitPC()
		case PVM.OUT_OF_GAS, PVM.PAGE_FAULT:
			return exitReason, exitPC
		case PVM.HOST_CALL:
			return exitReason, exitPC
		default:
			return PVM.ExitPanic, 0
		}
	}
}
