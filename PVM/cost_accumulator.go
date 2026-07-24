package PVM

import "github.com/New-JAMneration/JAM-Protocol/internal/telemetry"

// InvocationPhase identifies which PVM invocation context is active.
type InvocationPhase uint8

const (
	PhaseIsAuthorized InvocationPhase = iota
	PhaseRefine
	PhaseAccumulate
)

// CostCategory identifies a JIP-3 host-call cost bucket.
type CostCategory uint8

const (
	CatIsAuthHostCalls CostCategory = iota
	CatRefineHistLookup
	CatRefineMachExpunge
	CatRefinePeekPokePg
	CatRefineInvoke
	CatRefineOther
	CatAccumReadWrite
	CatAccumLookup
	CatAccumQSFP
	CatAccumINUE
	CatAccumTransfer
	CatAccumOther
	CatCount
)

// CostAccumulator tracks per-category gas usage during a single PVM invocation.
// One instance per Psi_A / Psi_R / is_authorized call. No mutex needed (single goroutine).
type CostAccumulator struct {
	phase   InvocationPhase
	gasUsed [CatCount]Gas
}

func NewCostAccumulator(phase InvocationPhase) *CostAccumulator {
	return &CostAccumulator{phase: phase}
}

func (ca *CostAccumulator) Add(cat CostCategory, gas Gas) {
	ca.gasUsed[cat] += gas
}

// RecordHostCall classifies the operation and accumulates the gas delta.
func (ca *CostAccumulator) RecordHostCall(op OperationType, gasDelta Gas) {
	cat := classifyOp(ca.phase, op)
	ca.gasUsed[cat] += gasDelta
}

// classifyOp maps (phase, opcode) to the appropriate JIP-3 CostCategory.
func classifyOp(phase InvocationPhase, op OperationType) CostCategory {
	switch phase {
	case PhaseIsAuthorized:
		return CatIsAuthHostCalls

	case PhaseRefine:
		switch op {
		case HistoricalLookupOp:
			return CatRefineHistLookup
		case MachineOp, ExpungeOp:
			return CatRefineMachExpunge
		case PeekOp, PokeOp, PagesOp:
			return CatRefinePeekPokePg
		case InvokeOp:
			return CatRefineInvoke
		default:
			return CatRefineOther
		}

	case PhaseAccumulate:
		switch op {
		case ReadOp, WriteOp:
			return CatAccumReadWrite
		case LookupOp:
			return CatAccumLookup
		case QueryOp, SolicitOp, ForgetOp, ProvideOp:
			return CatAccumQSFP
		case InfoOp, NewOp, UpgradeOp, EjectOp:
			return CatAccumINUE
		case TransferOp:
			return CatAccumTransfer
		default:
			return CatAccumOther
		}
	}
	return CatAccumOther
}

func (ca *CostAccumulator) ToIsAuthorizedCost(totalGas Gas) telemetry.IsAuthorizedCost {
	return telemetry.IsAuthorizedCost{
		Total:     telemetry.ExecCost{GasUsed: uint64(totalGas)},
		HostCalls: telemetry.ExecCost{GasUsed: uint64(ca.gasUsed[CatIsAuthHostCalls])},
	}
}

func (ca *CostAccumulator) ToRefineCost(totalGas Gas) telemetry.RefineCost {
	return telemetry.RefineCost{
		Total:            telemetry.ExecCost{GasUsed: uint64(totalGas)},
		HistoricalLookup: telemetry.ExecCost{GasUsed: uint64(ca.gasUsed[CatRefineHistLookup])},
		MachineExpunge:   telemetry.ExecCost{GasUsed: uint64(ca.gasUsed[CatRefineMachExpunge])},
		PeekPokePages:    telemetry.ExecCost{GasUsed: uint64(ca.gasUsed[CatRefinePeekPokePg])},
		Invoke:           telemetry.ExecCost{GasUsed: uint64(ca.gasUsed[CatRefineInvoke])},
		Other:            telemetry.ExecCost{GasUsed: uint64(ca.gasUsed[CatRefineOther])},
	}
}

func (ca *CostAccumulator) ToAccumulateCost(totalGas Gas) telemetry.AccumulateCost {
	return telemetry.AccumulateCost{
		Total:                     telemetry.ExecCost{GasUsed: uint64(totalGas)},
		ReadWrite:                 telemetry.ExecCost{GasUsed: uint64(ca.gasUsed[CatAccumReadWrite])},
		Lookup:                    telemetry.ExecCost{GasUsed: uint64(ca.gasUsed[CatAccumLookup])},
		QuerySolicitForgetProvide: telemetry.ExecCost{GasUsed: uint64(ca.gasUsed[CatAccumQSFP])},
		InfoNewUpgradeEject:       telemetry.ExecCost{GasUsed: uint64(ca.gasUsed[CatAccumINUE])},
		Transfer:                  telemetry.ExecCost{GasUsed: uint64(ca.gasUsed[CatAccumTransfer])},
		Other:                     telemetry.ExecCost{GasUsed: uint64(ca.gasUsed[CatAccumOther])},
	}
}
