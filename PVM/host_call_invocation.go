package PVM

// Ω⟨X⟩
type (
	Omega  func(OmegaInput) OmegaOutput
	Omegas []Omega
)

type OmegaInput struct {
	Operation   OperationType // operation type
	Interpreter *Interpreter  // interpreter state (registers, memory, gas)
	Addition    HostCallArgs  // Extra parameter for each host-call function
	HostCalls   Omegas        // host-call functions (needed for nested invocations)
}
type OmegaOutput struct {
	ExitReason ExitReason // Exit reason
	// NewGas       Gas          // New Gas
	// NewRegisters Registers    // New Register
	// NewMemory    Memory       // New Memory
	Addition HostCallArgs // addition host-call context
}

var (
	AccumulateOmegas   Omegas
	RefineOmegas       Omegas
	IsAuthorizedOmegas Omegas
)

func init() {
	// is authorized host-call functions
	IsAuthorizedOmegas = make(Omegas, len(HostCallFunctions))
	IsAuthorizedOmegas[GasOp] = HostCallFunctions[GasOp]
	IsAuthorizedOmegas[FetchOp] = HostCallFunctions[FetchOp]
	IsAuthorizedOmegas[100] = logHostCall

	// accumulate host-call functions
	AccumulateOmegas = make(Omegas, len(HostCallFunctions))
	AccumulateOmegas[GasOp] = HostCallFunctions[GasOp]
	AccumulateOmegas[FetchOp] = HostCallFunctions[FetchOp]
	AccumulateOmegas[ReadOp] = readWrapWithG
	AccumulateOmegas[WriteOp] = writeWrapWithG
	AccumulateOmegas[LookupOp] = lookupWrapWithG
	AccumulateOmegas[InfoOp] = infoWrapWithG
	AccumulateOmegas[BlessOp] = HostCallFunctions[BlessOp]
	AccumulateOmegas[AssignOp] = HostCallFunctions[AssignOp]
	AccumulateOmegas[DesignateOp] = HostCallFunctions[DesignateOp]
	AccumulateOmegas[CheckpointOp] = HostCallFunctions[CheckpointOp]
	AccumulateOmegas[NewOp] = HostCallFunctions[NewOp]
	AccumulateOmegas[UpgradeOp] = HostCallFunctions[UpgradeOp]
	AccumulateOmegas[TransferOp] = HostCallFunctions[TransferOp]
	AccumulateOmegas[EjectOp] = HostCallFunctions[EjectOp]
	AccumulateOmegas[QueryOp] = HostCallFunctions[QueryOp]
	AccumulateOmegas[SolicitOp] = HostCallFunctions[SolicitOp]
	AccumulateOmegas[ForgetOp] = HostCallFunctions[ForgetOp]
	AccumulateOmegas[YieldOp] = HostCallFunctions[YieldOp]
	AccumulateOmegas[ProvideOp] = HostCallFunctions[ProvideOp]
	AccumulateOmegas[100] = logHostCall

	// refine host-call functions
	RefineOmegas = make(Omegas, len(HostCallFunctions))
	RefineOmegas[GasOp] = HostCallFunctions[GasOp]
	RefineOmegas[FetchOp] = HostCallFunctions[FetchOp]
	RefineOmegas[HistoricalLookupOp] = HostCallFunctions[HistoricalLookupOp]
	RefineOmegas[ExportOp] = HostCallFunctions[ExportOp]
	RefineOmegas[MachineOp] = HostCallFunctions[MachineOp]
	RefineOmegas[PeekOp] = HostCallFunctions[PeekOp]
	RefineOmegas[PokeOp] = HostCallFunctions[PokeOp]
	RefineOmegas[PagesOp] = HostCallFunctions[PagesOp]
	RefineOmegas[InvokeOp] = HostCallFunctions[InvokeOp]
	RefineOmegas[ExpungeOp] = HostCallFunctions[ExpungeOp]
	RefineOmegas[100] = logHostCall
}

func getOmega(omegas Omegas, operation OperationType) Omega {
	if operation < 0 || int(operation) >= len(omegas) {
		return nil
	}
	return omegas[operation]
}

type Psi_H_ReturnType struct {
	ExitReason ExitReason   // exit reason
	Counter    uint32       // new instruction counter
	Gas        Gas          // gas remain
	Reg        Registers    // new registers
	Ram        Memory       // new memory
	Addition   HostCallArgs // addition host-call context
}

// (A.34) Ψ_H
func (h *Host) HostCall(pc ProgramCounter, instrCount uint64) (psi_result Psi_H_ReturnType) {
	for {
		var exitReason ExitReason
		var pcPrime ProgramCounter
		if GasChargingMode == "blockBased" {
			exitReason, pcPrime = h.Interpreter.BlockBasedInvoke(pc)
		} else {
			exitReason, pcPrime = h.Interpreter.SingleStepInvoke(pc)
		}

		switch exitReason.GetReasonType() {
		case HALT, PANIC, OUT_OF_GAS, PAGE_FAULT:
			psi_result.ExitReason = exitReason
			psi_result.Counter = uint32(pcPrime)
			psi_result.Gas = h.Interpreter.Gas
			psi_result.Reg = h.Interpreter.Registers
			psi_result.Ram = *h.Interpreter.Memory
			psi_result.Addition = h.Addition
			return
		}

		// reason.Reason == HOST_CALL
		var input OmegaInput
		input.Operation = OperationType(exitReason.GetHostCallID())
		input.Interpreter = &h.Interpreter
		input.Addition = h.Addition
		input.HostCalls = h.HostCalls

		omega := getOmega(h.HostCalls, input.Operation)
		if omega == nil {
			if h.Interpreter.Gas < 0 {
				omega = hostCallOutOfGas
			} else {
				omega = hostCallException
			}
		}
		omegaResult := omega(input)
		pvmLogger.Debugf("%s host-call return: %d, gas : %d\nRegisters: %v\n",
			hostCallName[input.Operation], omegaResult.ExitReason.GetReasonType(), h.Interpreter.Gas, h.Interpreter.Registers)

		switch omegaResult.ExitReason {
		case ExitContinue:
			h.Addition = omegaResult.Addition
			skipLength := ProgramCounter(skip(int(pcPrime), h.Addition.Program.Bitmasks))
			pc = pcPrime + skipLength + 1
			continue
		default:
			psi_result.ExitReason = omegaResult.ExitReason
			psi_result.Counter = uint32(pcPrime)
			psi_result.Gas = h.Interpreter.Gas
			psi_result.Reg = h.Interpreter.Registers
			psi_result.Ram = *h.Interpreter.Memory
			psi_result.Addition = omegaResult.Addition
			return
		}
	}
}
