package PolkaVM

import (
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type OnTransferInput struct {
	ServiceAccounts   types.ServiceAccountState
	Timeslot          types.TimeSlot
	ServiceID         types.ServiceId
	DeferredTransfers []types.DeferredTransfer
}

// Psi_T
func OnTransferInvoke(input OnTransferInput) (types.ServiceAccount, types.Gas) {
	emptyArray := types.OpaqueHash{}
	result := Psi_M_ReturnType{}
	account := types.ServiceAccount{}

	account, accountExists := input.ServiceAccounts[input.ServiceID]
	if !accountExists {
		log.Fatalf("OnTransferInvoke serviceAccount : %d not exists", input.ServiceID)
	}

	if account.ServiceInfo.CodeHash == emptyArray || len(input.DeferredTransfers) == 0 {
		return input.ServiceAccounts[input.ServiceID], 0
	}

	encoder := types.NewEncoder()
	// Psi_M arguments
	gasLimits := types.Gas(0)
	balances := types.U64(0)

	F := Omegas{}

	s := input.ServiceAccounts[input.ServiceID]
	s.ServiceInfo.Balance += balances

	// timeslot(t) bytes
	timeslotSer, _ := encoder.Encode(input.Timeslot)
	deferredTransferSer, _ := encoder.Encode(input.DeferredTransfers)
	// serviceID(s) bytes
	serviceIDSer, _ := encoder.Encode(input.ServiceID)

	var serialized []byte
	serialized = append(timeslotSer, serviceIDSer...)
	serialized = append(serialized, deferredTransferSer...)

	for _, deferredTransfer := range input.DeferredTransfers {
		gasLimits += deferredTransfer.GasLimit
		balances += deferredTransfer.Balance
	}

	// omegas
	F[LookupOp] = hostCallFunctions[0]
	F[ReadOp] = hostCallFunctions[ReadOp]
	F[WriteOp] = hostCallFunctions[WriteOp]
	F[GasOp] = hostCallFunctions[GasOp]
	F[InfoOp] = hostCallFunctions[InfoOp]
	F[27] = onTransferHostCallException
	// addition, on-transfer only uses GeneralArgs
	addition := HostCallArgs{
		GeneralArgs: GeneralArgs{
			ServiceAccount:      input.ServiceAccounts[input.ServiceID],
			ServiceId:           input.ServiceID,
			ServiceAccountState: input.ServiceAccounts,
		},
	}

	result = Psi_M(account.ServiceInfo.CodeHash[:], 10, Gas(gasLimits), serialized, F, addition)
	account = result.Addition.ServiceAccount

	return account, types.Gas(result.Gas)
}
