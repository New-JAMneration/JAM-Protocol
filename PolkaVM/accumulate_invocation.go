package PolkaVM

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

func Psi_I(workPackage types.WorkPackage, cores types.CoreIndex) (types.WorkExecResultType, types.ByteSequence, Gas) {
	return types.WorkExecResultOk, nil, 0
}

func Psi_R(i types.U64, p types.WorkPackage, o types.ByteSequence, iBar [][]types.ExportSegment, x types.U64) RefineOutput {
	return RefineOutput{}
}

type RefineOutput struct {
	WorkResult   types.WorkExecResultType
	RefineOutput []byte
	EportSegment []types.ExportSegment
	Gas          types.Gas
}

// (B.8) Ψ_A
func Psi_A(
	partialState types.PartialStateSet,
	time types.TimeSlot,
	serviceId types.ServiceId,
	gas types.Gas,
	operand []types.Operand,
) (
	psi_result Psi_A_ReturnType,
) {
	// if partialState.ServiceAccounts[serviceId].CodeHash != nil {
	// } else {
	// }
	return Psi_A_ReturnType{}
}

type Psi_A_ReturnType struct {
	PartialStateSet   types.PartialStateSet
	DeferredTransfers []types.DeferredTransfer
	Result            *types.OpaqueHash
	Gas               types.Gas
}

// (B.6)
type ResultContext struct {
	ServiceId         types.ServiceId
	PartialState      types.PartialStateSet
	ImportServiceId   types.ServiceId
	DeferredTransfers []types.DeferredTransfer
	Exception         types.OpaqueHash
}
