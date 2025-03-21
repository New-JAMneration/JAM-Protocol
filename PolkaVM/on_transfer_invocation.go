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

	if account, accountExists := input.ServiceAccounts[input.ServiceID]; accountExists {
		if account.ServiceInfo.CodeHash == emptyArray || len(input.DeferredTransfers) == 0 {
			return input.ServiceAccounts[input.ServiceID], 0
		}

		encoder := types.Encoder{}
		// Psi_M arguments
		gasLimits := types.Gas(0)
		balances := types.U64(0)
		serialized := []byte{}
		F := Omegas{}

		s := input.ServiceAccounts[input.ServiceID]
		s.ServiceInfo.Balance += balances

		// timeslot(t) bytes
		timeslotSer, err := encoder.Encode(input.Timeslot)
		if err != nil {
			log.Println("OnTransferInvoke encode TimeSlot error")
		}
		deferredTransferSer, err := encoder.Encode(input.DeferredTransfers)
		if err != nil {
			log.Println("OnTransferInvoke encode DeferredTransfer error")
		}

		// serviceID(s) bytes
		serviceIDSer, err := encoder.Encode(input.ServiceID)
		if err != nil {
			log.Println("OnTransferInvoke encode ServiceID error")
		}

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
		F[26] = onTransferHostCallException
		// Psi_M the last arugment : standardProgram type will be removed
		result := Psi_M(account.ServiceInfo.CodeHash[:], 10, Gas(gasLimits), serialized, F, nil, StandardProgram{})
		if serviceAccount, isServiceAccount := result.Addition.(types.ServiceAccount); isServiceAccount {
			account = serviceAccount
		}
	} else {
		log.Println("OnTransferInvoke account not exists")
	}

	return account, types.Gas(result.Gas)
}
