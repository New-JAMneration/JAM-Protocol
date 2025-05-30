package PolkaVM

import (
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
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
	account, accountExists := input.ServiceAccounts[input.ServiceID]
	if !accountExists {
		log.Fatalf("OnTransferInvoke serviceAccount : %d not exists", input.ServiceID)
		return account, 0
	}

	codeHash := account.ServiceInfo.CodeHash
	programCode, programCodeExists := account.PreimageLookup[codeHash]
	if !programCodeExists || len(programCode) == 0 || len(programCode) > types.MaxServiceCodeSize || len(input.DeferredTransfers) == 0 {
		return account, 0
	}

	encoder := types.NewEncoder()
	// Psi_M arguments
	gasLimits := types.Gas(0)
	balances := types.U64(0)

	F := Omegas{}

	account.ServiceInfo.Balance += balances

	// timeslot(t) bytes
	timeslotSer, _ := encoder.Encode(input.Timeslot)
	deferredTransferSer, _ := encoder.EncodeUint(uint64(len(input.DeferredTransfers)))
	// serviceID(s) bytes
	serviceIDSer, _ := encoder.Encode(input.ServiceID)

	var serialized []byte
	serialized = append(serialized, timeslotSer...)
	serialized = append(serialized, serviceIDSer...)
	serialized = append(serialized, deferredTransferSer...)

	for _, deferredTransfer := range input.DeferredTransfers {
		gasLimits += deferredTransfer.GasLimit
		balances += deferredTransfer.Balance
	}

	store := store.GetInstance()
	eta0 := store.GetPriorStates().GetEta()[0]

	// omegas
	F[LookupOp] = HostCallFunctions[LookupOp]
	F[FetchOp] = HostCallFunctions[FetchOp] // added 0.6.6
	F[ReadOp] = HostCallFunctions[ReadOp]
	F[WriteOp] = HostCallFunctions[WriteOp]
	F[GasOp] = HostCallFunctions[GasOp]
	F[InfoOp] = HostCallFunctions[InfoOp]
	F[OperationType(len(HostCallFunctions)-1)] = onTransferHostCallException
	// addition, on-transfer only uses GeneralArgs
	addition := HostCallArgs{
		GeneralArgs: GeneralArgs{
			ServiceAccount:      input.ServiceAccounts[input.ServiceID],
			ServiceId:           input.ServiceID,
			ServiceAccountState: input.ServiceAccounts,
		},
		AccumulateArgs: AccumulateArgs{
			Eta: eta0,
		},
		// TODO: double-check whether Eta0 is a typo
		OnTransferArgs: OnTransferArgs{
			DeferredTransfer: input.DeferredTransfers,
		},
	}

	result := Psi_M(StandardCodeFormat(programCode), 10, gasLimits, serialized, F, addition)
	account = result.Addition.ServiceAccount

	return account, types.Gas(result.Gas)
}
