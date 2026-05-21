package PVM

// Ω⟨X⟩
type (
	Omega  func(OmegaInput) OmegaOutput
	Omegas []Omega
)

type OmegaInput struct {
	Operation OperationType // operation type
	VM        *VMState      // VM state (registers, memory, gas) — shared by Interpreter and Recompiler
	Addition  HostCallArgs  // Extra parameter for each host-call function
	HostCalls Omegas        // host-call functions (needed for nested invocations)
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

// GetOmega is the exported alias used by PVM/recompiler.
func GetOmega(omegas Omegas, operation OperationType) Omega {
	return getOmega(omegas, operation)
}

type Psi_H_ReturnType struct {
	ExitReason ExitReason   // exit reason
	Counter    uint32       // new instruction counter
	VM         *VMState     // final VM state
	Addition   HostCallArgs // addition host-call context
}
