package PolkaVM

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

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
	ServiceId         types.ServiceId          // s
	PartialState      types.PartialStateSet    // u
	ImportServiceId   types.ServiceId          // i
	DeferredTransfers []types.DeferredTransfer // t
	Exception         types.OpaqueHash         // y
}
