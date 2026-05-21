//go:build trace

package interpreter

import (
	"encoding/json"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/PVMtrace"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func wireInterpreterAccumulateTrace(host *Host, programCode []byte, counter PVM.ProgramCounter, gas types.Gas, registers PVM.Registers, addition PVM.HostCallArgs) func(PVM.Psi_H_ReturnType) {
	var trace *PVMtrace.Trace
	if addition.AccumulateTrace != nil {
		init := PVMtrace.InitialState{
			PC:   uint64(counter),
			Gas:  int64(gas),
			Regs: registers,
		}
		ctx := addition.AccumulateTrace
		trace = PVMtrace.NewTrace(
			uint32(ctx.ServiceID),
			[32]byte(ctx.CodeHash),
			uint64(ctx.Timeslot),
			programCode,
			init,
			"accumulate",
			PVMtrace.BackendInterpreter,
		)
	}
	host.Interpreter.Trace = trace
	return func(h PVM.Psi_H_ReturnType) {
		if trace == nil {
			return
		}
		var regs [13]uint64
		var finalGas int64
		if h.VM != nil && h.VM.Registers != nil {
			copy(regs[:], (*h.VM.Registers)[:])
		}
		if h.VM != nil && h.VM.Gas != nil {
			finalGas = int64(*h.VM.Gas)
		}
		_ = trace.Close(PVMtrace.FinalState{
			PC:         uint64(h.Counter),
			Gas:        finalGas,
			ExitReason: h.ExitReason.String(),
			Regs:       regs,
		})
	}
}

func (h *Host) beginAccumulateHostCallTrace() (rin [13]uint64, gIn int64, ok bool) {
	if h.Interpreter.Trace == nil {
		return [13]uint64{}, 0, false
	}
	copy(rin[:], h.Interpreter.Registers[:])
	return rin, int64(h.Interpreter.Gas), true
}

func wrapGuestMemoryForHostCallTrace(mem PVM.GuestMemory) (PVM.GuestMemory, func() json.RawMessage) {
	detailsFn, wrapped := PVM.BeginHostCallMemTrace(mem)
	return wrapped, detailsFn
}

func (h *Host) recordAccumulateHostCallAfterOmega(op PVM.OperationType, rin [13]uint64, gIn int64, rout [13]uint64, gOut int64, captured bool, ecalliPC uint64, omegaExit PVM.ExitReason, details json.RawMessage) {
	if !captured || h.Interpreter.Trace == nil {
		return
	}
	trace := h.Interpreter.Trace
	step := int(trace.Steps()) - 1
	if step < 0 {
		step = 0
	}
	trace.RecordHostCall(PVMtrace.HostCallRecord{
		Step:       step,
		PC:         ecalliPC,
		Op:         uint64(op),
		OpName:     PVM.HostCallName(int(op)),
		RegsIn:     rin,
		RegsOut:    rout,
		GasIn:      gIn,
		GasOut:     gOut,
		ExitReason: omegaExit.String(),
		Details:    details,
	})
}
