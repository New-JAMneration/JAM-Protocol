// Package interpreter is the cross-platform (pure-Go) PVM execution backend.
//
// It owns the host-call orchestration layer (Host / Host.HostCall), the Psi_M
// interpreter driver, and the accumulate-trace wiring. The instruction
// execution engine itself (the Interpreter type, the per-opcode semantics, the
// decode/block driver, the paged Memory accessor, omega host-calls, R(), and
// the Psi_M dispatcher) stays in package PVM (core): per-opcode semantics are
// shared with the inner-invoke executor, and keeping the Interpreter type in
// core avoids an import cycle (core's decoded InstrMeta carries
// func(*PVM.Interpreter, ...) execution pointers).
//
// This package registers PVM.Psi_M_interpreterHook in init(); every binary that
// runs the PVM must blank-import it (see Psi_M's panic guard).
package interpreter

import (
	"encoding/json"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/logger"
)

var pvmLogger = logger.GetLogger("pvm")

type Host struct {
	Interpreter PVM.Interpreter
	Addition    PVM.HostCallArgs
	HostCalls   PVM.Omegas
	// CostAccum *PVM.CostAccumulator // reserved for future use (not wired yet)
}

func NewHost(program *PVM.Program, registers PVM.Registers, memory *PVM.Memory, gas PVM.Gas, addition PVM.HostCallArgs, hostCalls PVM.Omegas) *Host {
	return &Host{
		Interpreter: PVM.Interpreter{
			Program:   program,
			Registers: registers,
			Memory:    memory,
			Gas:       gas,
		},
		Addition:  addition,
		HostCalls: hostCalls,
	}
}

// (A.34) Ψ_H
func (h *Host) HostCall(pc PVM.ProgramCounter, instrCount uint64) (psi_result PVM.Psi_H_ReturnType) {
	for {
		var exitReason PVM.ExitReason
		var pcPrime PVM.ProgramCounter

		exitReason, pcPrime = h.Interpreter.SingleStepInvokeDecodedBlocks(pc)

		switch exitReason.GetReasonType() {
		case PVM.HALT, PVM.PANIC, PVM.OUT_OF_GAS, PVM.PAGE_FAULT:
			psi_result.ExitReason = exitReason
			psi_result.Counter = uint32(pcPrime)
			psi_result.VM = &PVM.VMState{
				Registers: &h.Interpreter.Registers,
				Mem:       PVM.NewPagedGuestMemory(h.Interpreter.Memory),
				Gas:       &h.Interpreter.Gas,
			}
			psi_result.Addition = h.Addition
			return
		}

		// reason.Reason == HOST_CALL
		baseMem := PVM.NewPagedGuestMemory(h.Interpreter.Memory)
		tracedMem, memDetails := wrapGuestMemoryForHostCallTrace(baseMem)

		var input PVM.OmegaInput
		input.Operation = PVM.OperationType(exitReason.GetHostCallID())
		input.VM = &PVM.VMState{
			Registers: &h.Interpreter.Registers,
			Mem:       tracedMem,
			Gas:       &h.Interpreter.Gas,
		}
		input.Addition = h.Addition
		input.HostCalls = h.HostCalls

		omega := PVM.GetOmega(h.HostCalls, input.Operation)
		if omega == nil {
			if h.Interpreter.Gas < 0 {
				omega = PVM.HostCallOutOfGas
			} else {
				omega = PVM.HostCallException
			}
		}

		rin, gIn, captured := h.beginAccumulateHostCallTrace()
		ecalliPC := PVM.HostCallInstrPC(h.Interpreter.Program, pcPrime)

		omegaResult := omega(input)

		var rout [13]uint64
		var details json.RawMessage
		if captured {
			copy(rout[:], h.Interpreter.Registers[:])
			details = memDetails()
		}
		h.recordAccumulateHostCallAfterOmega(
			input.Operation, rin, gIn, rout, int64(h.Interpreter.Gas),
			captured, uint64(ecalliPC), omegaResult.ExitReason, details,
		)

		pvmLogger.Debugf("%s host-call return: %d, gas : %d\nRegisters: %v\n",
			PVM.HostCallName(int(input.Operation)), omegaResult.ExitReason.GetReasonType(), h.Interpreter.Gas, h.Interpreter.Registers)

		switch omegaResult.ExitReason {
		case PVM.ExitContinue:
			h.Addition = omegaResult.Addition
			// SingleStepInvokeDecodedBlocks already returns the next instruction PC
			// (ecalli.PC + skipLen + 1 = fallthrough PC), so no skip needed.
			pc = pcPrime
			continue
		default:
			psi_result.ExitReason = omegaResult.ExitReason
			psi_result.Counter = uint32(pcPrime)
			psi_result.VM = &PVM.VMState{
				Registers: &h.Interpreter.Registers,
				Mem:       PVM.NewPagedGuestMemory(h.Interpreter.Memory),
				Gas:       &h.Interpreter.Gas,
			}
			psi_result.Addition = omegaResult.Addition
			return
		}
	}
}
