//go:build linux && amd64

package recompiler

import (
	"encoding/json"
	"time"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/PVMtrace"
)

// host is the host-call dispatch layer for the JIT backend, symmetrical
// to PVM.Host on the interpreter backend.
//
// Layering recap:
//
//	host       ~  PVM.Host           (host-call dispatch)
//	Recompiler ~  PVM.Interpreter    (machine)
type host struct {
	recomp    *Recompiler
	Addition  PVM.HostCallArgs
	HostCalls PVM.Omegas
}

func newHost(r *Recompiler, addition PVM.HostCallArgs, hostCalls PVM.Omegas) *host {
	return &host{
		recomp:    r,
		Addition:  addition,
		HostCalls: hostCalls,
	}
}

func (h *host) HostCall(pc PVM.ProgramCounter) PVM.Psi_H_ReturnType {
	ctx := h.recomp.Ctx()
	var vm PVM.VMState
	regsBuf, gasBuf := vm.InlineSnapshotPtrs()

	snapshot := func() {
		ctx.ReadRegistersInto(regsBuf)
		ctx.ReadGasInto(gasBuf)
		vm.Mem = ctx.GuestMemory()
		vm.BindInlineSnapshot()
	}

	for {
		exitReason, pcPrime := h.recomp.MachineInvoke(pc)
		switch exitReason.GetReasonType() {
		case PVM.HALT, PVM.PANIC, PVM.OUT_OF_GAS, PVM.PAGE_FAULT:
			snapshot()
			return PVM.Psi_H_ReturnType{
				ExitReason: exitReason,
				Counter:    uint32(pcPrime),
				VM:         escapeVMState(&vm),
				Addition:   h.Addition,
			}
		}

		snapshot()

		// unreachable: sbrk is resolved inside BlockBasedInvoke / DebugSingleStepInvoke

		input := PVM.OmegaInput{
			Operation: PVM.OperationType(exitReason.GetHostCallID()),
			VM:        &vm,
			Addition:  h.Addition,
			HostCalls: h.HostCalls,
		}

		omega := PVM.GetOmega(h.HostCalls, input.Operation)
		PVMtrace.LogHostCallDispatchEnv(uint32(*h.Addition.ServiceID), PVM.HostCallName(int(input.Operation)), regsBuf, int64(*gasBuf))
		if omega == nil {
			if *gasBuf < 0 {
				omega = PVM.HostCallOutOfGas
			} else {
				omega = PVM.HostCallException
			}
		}

		var rin, rout [13]uint64
		var gIn int64
		var memDetails func() json.RawMessage
		trace := h.recomp.Trace
		if trace != nil {
			copy(rin[:], regsBuf[:])
			gIn = int64(*gasBuf)
			vm.Mem, memDetails = wrapGuestMemoryForHostCallTrace(ctx)
		}

		var tHost time.Time
		if jitProfile {
			tHost = time.Now()
		}
		omegaResult := omega(input)
		if jitProfile {
			jm.hostNanos.Add(int64(time.Since(tHost)))
			jm.hostCalls.Add(1)
		}
		regs, gas := vm.InlineSnapshotValues()
		ctx.WriteRegisters(regs)
		ctx.WriteGas(gas)

		if trace != nil {
			copy(rout[:], regs[:])
			step := int(trace.Steps()) - 1
			if step < 0 {
				step = 0
			}
			trace.RecordHostCall(PVMtrace.HostCallRecord{
				Step:       step,
				PC:         uint64(PVM.HostCallInstrPC(h.recomp.program, pcPrime)),
				Op:         uint64(input.Operation),
				OpName:     PVM.HostCallName(int(input.Operation)),
				RegsIn:     rin,
				RegsOut:    rout,
				GasIn:      gIn,
				GasOut:     int64(gas),
				ExitReason: omegaResult.ExitReason.String(),
				Details:    memDetails(),
			})
		}

		switch omegaResult.ExitReason {
		case PVM.ExitContinue:
			h.Addition = omegaResult.Addition
			pc = pcPrime
			continue
		default:
			return PVM.Psi_H_ReturnType{
				ExitReason: omegaResult.ExitReason,
				Counter:    uint32(pcPrime),
				VM:         escapeVMState(&vm),
				Addition:   omegaResult.Addition,
			}
		}
	}
}

func escapeVMState(local *PVM.VMState) *PVM.VMState {
	heap := *local
	heap.BindInlineSnapshot()
	return &heap
}
