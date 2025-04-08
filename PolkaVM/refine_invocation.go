package PolkaVM

import (
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

type RefineInput struct {
	WorkItemIndex       uint                    // i
	WorkPackage         types.WorkPackage       // p
	AuthOutput          types.ByteSequence      // o
	ImportSegments      [][]types.ExportSegment // bold{i}
	ExportSegmentOffset uint                    // zeta
	ServiceAccounts     types.ServiceAccountState
}

type RefineOutput struct {
	WorkResult    types.WorkExecResultType
	RefineOutput  []byte
	ExportSegment []types.ExportSegment
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
	// lookupData = Λ(δ[w_s], (p_x)_t, w_c)
	lookupData := service_account.HistoricalLookup(account, input.WorkPackage.Context.LookupAnchorSlot, workItem.CodeHash)

	if accountExists || (accountExists && len(lookupData) == 0) {
		return RefineOutput{
			WorkResult:    types.WorkExecResultBadCode,
			ExportSegment: []types.ExportSegment{},
			Gas:           0,
		}
	}

	// check BIG
	if len(lookupData) > types.MaxServiceCodeSize {
		return RefineOutput{
			WorkResult:    types.WorkExecResultCodeOversize,
			ExportSegment: []types.ExportSegment{},
			Gas:           0,
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
	_, code, err := service_account.DecodeMetaCode(lookupData)
	if err != nil {
		log.Fatalf("refine invoke (Psi_R) decode metaCode error : %v", err)
	}

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
			ExportSegment:    []types.ExportSegment{},
			TimeSlot:         input.WorkPackage.Context.LookupAnchorSlot,
		},
	}
	//	WorkExecResultOutOfGas                        = "out-of-gas"
	// WorkExecResultPanic
	// result = u, r, (m,e)
	result := Psi_M(StandardCodeFormat(code), 0, Gas(workItem.RefineGasLimit), a, F, addition)

	if result.ReasonOrBytes == PANIC {
		return RefineOutput{
			WorkResult:    types.WorkExecResultPanic,
			RefineOutput:  []byte{},
			ExportSegment: []types.ExportSegment{},
			Gas:           types.Gas(result.Gas),
		}
	}
	if result.ReasonOrBytes == OUT_OF_GAS {
		return RefineOutput{
			WorkResult:    types.WorkExecResultOutOfGas,
			RefineOutput:  []byte{},
			ExportSegment: []types.ExportSegment{},
			Gas:           types.Gas(result.Gas),
		}
	}
	// otherwise
	var refineOutput []byte
	if output, isByteSlice := result.ReasonOrBytes.([]byte); isByteSlice {
		refineOutput = output
	}

	return RefineOutput{
		WorkResult:    types.WorkExecResultOk,
		RefineOutput:  refineOutput,
		ExportSegment: result.Addition.ExportSegment,
		Gas:           types.Gas(result.Gas),
	}
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
