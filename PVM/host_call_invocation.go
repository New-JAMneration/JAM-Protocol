package PVM

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

// Ω⟨X⟩
type (
	Omega  func(OmegaInput) OmegaOutput
	Omegas []Omega
)

type OmegaInput struct {
	Operation OperationType // operation type
	Gas       Gas           // gas counter
	Registers Registers     // PVM registers
	Memory    Memory        // memory
	Addition  HostCallArgs  // Extra parameter for each host-call function
}
type OmegaOutput struct {
	ExitReason   ExitReason   // Exit reason
	NewGas       Gas          // New Gas
	NewRegisters Registers    // New Register
	NewMemory    Memory       // New Memory
	Addition     HostCallArgs // addition host-call context
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
func HostCall(pc ProgramCounter, gas types.Gas, reg Registers, ram Memory, omegas Omegas, addition HostCallArgs, instrCount uint64,
) (psi_result Psi_H_ReturnType) {
	var exitReason ExitReason
	var pcPrime ProgramCounter
	var gasPrime Gas
	var regPrime Registers
	var memPrime Memory

	if GasChargingMode == "blockBased" {
		exitReason, pcPrime, gasPrime, regPrime, memPrime = BlockBasedInvoke(addition.Program, pc, Gas(gas), reg, ram)
	} else {
		exitReason, pcPrime, gasPrime, regPrime, memPrime = SingleStepInvoke(addition.Program, pc, Gas(gas), reg, ram)
	}

	switch exitReason.GetReasonType() {
	case HALT, PANIC, OUT_OF_GAS, PAGE_FAULT:
		psi_result.ExitReason = exitReason
		psi_result.Counter = uint32(pcPrime)
		psi_result.Gas = gasPrime
		psi_result.Reg = regPrime
		psi_result.Ram = memPrime
		psi_result.Addition = addition
		return
	}

	// reason.Reason == HOST_CALL
	var input OmegaInput
	input.Operation = OperationType(exitReason.GetHostCallID())
	input.Gas = gasPrime
	input.Registers = regPrime
	input.Memory = ram
	input.Addition = addition

	omega := getOmega(omegas, input.Operation)
	if omega == nil {
		if gasPrime < 0 {
			omega = hostCallOutOfGas
		} else {
			omega = hostCallException
		}
	}
	omegaResult := omega(input)
	pvmLogger.Debugf("%s host-call return: %d, gas : %d -> %d\nRegisters: %v\n",
		hostCallName[input.Operation], omegaResult.ExitReason.GetReasonType(), gasPrime, omegaResult.NewGas, omegaResult.NewRegisters)
	switch omegaResult.ExitReason {
	case ExitContinue:
		skipLength := ProgramCounter(skip(int(pcPrime), addition.Program.Bitmasks))
		return HostCall(pcPrime+skipLength+1, types.Gas(omegaResult.NewGas), omegaResult.NewRegisters, omegaResult.NewMemory, omegas, omegaResult.Addition, instrCount)
	default: // PANIC, OUT_OF_GAS, HALT
		psi_result.ExitReason = omegaResult.ExitReason
		psi_result.Counter = uint32(pcPrime)
		psi_result.Gas = omegaResult.NewGas
		psi_result.Reg = omegaResult.NewRegisters
		psi_result.Ram = omegaResult.NewMemory
		psi_result.Addition = omegaResult.Addition
		return
	}
}
