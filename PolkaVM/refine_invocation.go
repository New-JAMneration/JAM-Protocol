package PolkaVM

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

type ExportSegment [types.SegmentSize]byte

type RefineInput struct {
	WorkItemIndex       uint               // i
	WorkPackage         types.WorkPackage  // p
	AuthOutput          types.ByteSequence // o
	ImportSegments      [][]ExportSegment  // bold{i}
	ExportSegmentOffset uint               // zeta
	ServiceAccounts     types.ServiceAccountState
}

type RefineOutput struct {
	WorkResult    types.WorkExecResultType
	WorkReport    types.WorkReport
	ExportSegment []ExportSegment
	Gas           types.Gas
}

// B.4 M
type IntegratedPVMType struct {
	ProgramCode ProgramCode    // p
	Memory      Memory         // u
	PC          ProgramCounter // i
}

type IntegratedPVMMap map[uint64]IntegratedPVMType

// 14.17 P_n : Y -> Y_(k*n)
func zeroPadding(seq types.ByteSequence, n int) types.ByteSequence {
	// [0,0,0, ...]_((|x|+n-1) % n) + 1 ... n
	sub := (len(seq)+n-1)%n + 1
	padding := make([]byte, n-sub)
	return append(seq, padding...)
}

func RefineInvoke(input RefineInput) RefineOutput {
	// p_w[i]
	workItem := input.WorkPackage.Items[input.WorkItemIndex]

	// check BAD
	account, accountExists := input.ServiceAccounts[workItem.Service]
	if accountExists || (accountExists && len(service_account.HistoricalLookupFunction(account, input.WorkPackage.Context.LookupAnchorSlot, workItem.CodeHash)) == 0) {
		return RefineOutput{
			WorkResult: types.WorkExecResultBadCode,
			// ImportSegment: types.ImportSpec{},
			Gas: 0,
		}
	}

	// check BIG
	if len(service_account.HistoricalLookupFunction(account, input.WorkPackage.Context.LookupAnchorSlot, workItem.CodeHash)) > types.MaxServiceCodeSize {
		return RefineOutput{
			WorkResult: types.WorkExecResultCodeOversize,
			// ExportSegment: types.ImportSpec{},
			Gas: 0,
		}
	}

	// otherwise
	var a []byte
	encoder := types.NewEncoder()
	// w_s
	encoded, _ := encoder.Encode(workItem.CodeHash)
	a = append(a, encoded...)
	// w_y
	encoded, _ = encoder.Encode(workItem.Payload)
	a = append(a, encoded...)
	// H(p)
	encoded, _ = encoder.Encode(input.WorkPackage)
	h := hash.Blake2bHash(encoded)
	a = append(a, h[:]...)
	// p_x
	encoded, _ = encoder.Encode(input.WorkPackage.Context)
	a = append(a, encoded...)
	// p_u
	encoded, _ = encoder.Encode(input.WorkPackage.Authorizer.CodeHash)
	a = append(a, encoded...)

	// E(m, c)
	// lookupData := service_account.HistoricalLookupFunction(account, input.WorkPackage.Context.LookupAnchorSlot, workItem.CodeHash)

	F := Omegas{}
	F[HistoricalLookupOp] = HostCallFunctions[HistoricalLookupOp]
	F[FetchOp] = HostCallFunctions[FetchOp]
	F[ExportOp] = HostCallFunctions[ExportOp]
	F[GasOp] = HostCallFunctions[GasOp]
	F[MachineOp] = HostCallFunctions[MachineOp]
	F[PeekOp] = HostCallFunctions[PeekOp]
	F[ZeroOp] = HostCallFunctions[ZeroOp]
	F[PokeOp] = HostCallFunctions[PokeOp]
	F[VoidOp] = HostCallFunctions[VoidOp]
	F[InvokeOp] = HostCallFunctions[InvokeOp]
	F[ExpungeOp] = HostCallFunctions[ExpungeOp]
	F[27] = RefineHostCallException

	// addition
	// Though Psi_M addition input is nil, still need the RefineInput for historical_lookup op (only for historical_lookup)
	addition := HostCallArgs{
		// GeneralArgs is only for historical_lookup op
		GeneralArgs: GeneralArgs{
			ServiceId:           workItem.Service,
			ServiceAccountState: input.ServiceAccounts,
		},
		RefineArgs: RefineArgs{
			RefineInput:      input,
			IntegratedPVMMap: IntegratedPVMMap{},
			ExportSegment:    types.ExportSegment{},
			TimeSlot:         input.WorkPackage.Context.LookupAnchorSlot,
		},
	}

	// result = u, r, (m,e)
	// TODO : get code from historical_lookup
	// TODO : r type : ExecResult or bytes or report (?)
	result := Psi_M(code, 0, Gas(workItem.RefineGasLimit), a, F, addition)
	exitReason := result.ReasonOrBytes.(*PVMExitReason).Reason
	if exitReason == PANIC || exitReason == OUT_OF_GAS {
		return RefineOutput{
			WorkResult:    types.WorkExecResultOk,
			WorkReport:    result.ReasonOrBytes,
			ExportSegment: []ExportSegment{},
			Gas:           types.Gas(result.Gas),
		}
	}

	return RefineOutput{
		WorkResult:    types.WorkExecResultOk,
		WorkReport:    result.ReasonOrBytes,
		ExportSegment: result.Addition.ExportSegment,
		Gas:           types.Gas(result.Gas),
	}

	return RefineOutput{}
}

func RefineHostCallException(input OmegaInput) (output OmegaOutput) {
	input.Registers[7] = WHAT
	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       input.Gas - 10,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}
