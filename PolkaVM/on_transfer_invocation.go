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

	if account, accountExists := input.ServiceAccounts[input.ServiceID]; accountExists {
		if account.ServiceInfo.CodeHash == emptyArray || len(input.DeferredTransfers) == 0 {
			return input.ServiceAccounts[input.ServiceID], 0
		}

		encoder := types.Encoder{}
		// Psi_M arguments
		gasLimits := types.Gas(0)
		balances := types.U64(0)
		serialized := []byte{}

		s := input.ServiceAccounts[input.ServiceID]
		s.ServiceInfo.Balance += balances
		F := hostCallFunctions

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
		serialized = append(timeslotSer, serviceIDSer...)
		serialized = append(serialized, deferredTransferSer...)

		for _, deferredTransfer := range input.DeferredTransfers {
			gasLimits += deferredTransfer.GasLimit
			balances += deferredTransfer.Balance
		}

		// !warning: should input service count blob, not codehash, how to get code blob ?
		result := Psi_M(account.ServiceInfo.CodeHash[:], 10, Gas(gasLimits), serialized, F, nil, s)

		// Psi_M
	} else {
		log.Println("OnTransferInvoke account not exists")
	}

	return types.ServiceAccount{}, 0
}
