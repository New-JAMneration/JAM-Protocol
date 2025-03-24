package PolkaVM

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// (B.8) Ψ_A
func Psi_A(
	code StandardCodeFormat,
	partialState types.PartialStateSet,
	time types.TimeSlot,
	serviceId types.ServiceId,
	gas types.Gas,
	operand []types.Operand,
) (
	psi_result Psi_A_ReturnType,
) {
	c := partialState.ServiceAccounts[serviceId].ServiceInfo.CodeHash
	if storedCode, ok := partialState.ServiceAccounts[serviceId].PreimageLookup[c]; !ok {
		return Psi_A_ReturnType{
			PartialStateSet:   I(partialState, serviceId).PartialState,
			DeferredTransfers: []types.DeferredTransfer{},
			Result:            nil,
			Gas:               0,
		}
	} else {
		serialized := utilities.SerializeFixedLength(types.U32(time), 4)
		serialized = append(serialized, utilities.SerializeFixedLength(types.U32(serviceId), 4)...)
		serialized = append(serialized, utilities.SerializeU64(types.U64(len(operand)))...)
		for _, op := range operand {
			serialized = append(serialized, utilities.SerializeByteSequence(op.Hash[:])...)
			serialized = append(serialized, utilities.SerializeByteSequence(op.ExportsRoot[:])...)
			serialized = append(serialized, utilities.SerializeByteSequence(op.AuthorizerHash[:])...)
			serialized = append(serialized, utilities.SerializeByteSequence(op.AuthOutput)...)
			serialized = append(serialized, utilities.SerializeByteSequence(op.PayloadHash[:])...)
			serialized = append(serialized, utilities.SerializeWorkExecResult(op.Result)...)
		}

		F := Omegas{}
		F[ReadOp] = wrapWithG(hostCallFunctions[ReadOp])
		F[WriteOp] = wrapWithG(hostCallFunctions[WriteOp])
		F[LookupOp] = wrapWithG(hostCallFunctions[LookupOp])
		F[GasOp] = hostCallFunctions[GasOp]
		F[InfoOp] = wrapWithG(hostCallFunctions[InfoOp])
		F[BlessOp] = hostCallFunctions[BlessOp]
		F[AssignOp] = hostCallFunctions[AssignOp]
		F[DesignateOp] = hostCallFunctions[DesignateOp]
		F[CheckpointOp] = hostCallFunctions[CheckpointOp]
		F[NewOp] = hostCallFunctions[NewOp]
		F[UpgradeOp] = hostCallFunctions[UpgradeOp]
		F[TransferOp] = hostCallFunctions[TransferOp]
		F[EjectOp] = hostCallFunctions[EjectOp]
		F[QueryOp] = hostCallFunctions[QueryOp]
		F[SolicitOp] = hostCallFunctions[SolicitOp]
		F[ForgetOp] = hostCallFunctions[ForgetOp]
		F[YieldOp] = hostCallFunctions[YieldOp]
		F[26] = accumulateInvocationHostCallException

		addition := HostCallArgs{
			AccumulateArgs: AccumulateArgs{
				ResultContextX: I(partialState, serviceId),
				ResultContextY: I(partialState, serviceId),
			},
		}

		resultM := Psi_M(StandardCodeFormat(storedCode), 5, Gas(gas), Argument(serialized), F, addition)
		partialState, deferredTransfer, result, gas := C(types.Gas(resultM.Gas), resultM.ReasonOrBytes, AccumulateArgs{
			ResultContextX: resultM.Addition.AccumulateArgs.ResultContextX,
			ResultContextY: resultM.Addition.AccumulateArgs.ResultContextY,
		})
		return Psi_A_ReturnType{
			PartialStateSet:   partialState,
			DeferredTransfers: deferredTransfer,
			Result:            result,
			Gas:               gas,
		}
	}
}

func accumulateInvocationHostCallException(input OmegaInput) (output OmegaOutput) {
	input.Registers[7] = WHAT
	return OmegaOutput{
		ExitReason:   PVMExitTuple(CONTINUE, nil),
		NewGas:       input.Gas - 10,
		NewRegisters: input.Registers,
		NewMemory:    input.Memory,
		Addition:     input.Addition,
	}
}

// (B.12) G
func G(o OmegaOutput, serviceAccount types.ServiceAccount) OmegaOutput {
	o.Addition.AccumulateArgs.ResultContextX.PartialState.ServiceAccounts[o.Addition.AccumulateArgs.ResultContextX.ServiceId] = serviceAccount
	return o
}

func wrapWithG(original Omega) Omega {
	return func(input OmegaInput) OmegaOutput {
		output := original(input)
		serviceAccount := input.Addition.AccumulateArgs.ResultContextX.PartialState.ServiceAccounts[input.Addition.AccumulateArgs.ResultContextX.ServiceId]
		return G(output, serviceAccount)
	}
}

// (B.13) C
func C(gas types.Gas, reasonOrBytes any, resultContext AccumulateArgs) (types.PartialStateSet, []types.DeferredTransfer, *types.OpaqueHash, types.Gas) {
	switch reasonOrBytes.(type) {
	case error:
		return resultContext.ResultContextY.PartialState, resultContext.ResultContextY.DeferredTransfers, resultContext.ResultContextY.Exception, gas
	case types.ByteSequence:
		opaqueHash := reasonOrBytes.(*types.OpaqueHash)
		return resultContext.ResultContextX.PartialState, resultContext.ResultContextX.DeferredTransfers, opaqueHash, gas
	default:
		return resultContext.ResultContextX.PartialState, resultContext.ResultContextX.DeferredTransfers, resultContext.ResultContextX.Exception, gas
	}
}

// (B.10)
func I(partialState types.PartialStateSet, serviceId types.ServiceId) ResultContext {
	eta := store.GetInstance().GetPosteriorStates().GetEta()
	ht := store.GetInstance().GetIntermediateHeader().Slot

	serialized := utilities.SerializeFixedLength(types.U32(serviceId), 4)
	for _, eta := range eta {
		serialized = append(serialized, utilities.SerializeByteSequence(eta[:])...)
	}
	serialized = append(serialized, utilities.SerializeFixedLength(types.U64(ht), 8)...)
	hash := hash.Blake2bHash(serialized)
	result, err := utilities.DeserializeFixedLength(hash[:], types.U32(4))
	if err != nil {
		panic(err)
	}
	var modValue types.U32 = (1 << 32) - (1 << 9) // 2^32 - 2^9
	var addValue types.U32 = 1 << 8               // 2^8
	result = (result % modValue) + addValue

	return ResultContext{
		ServiceId:         serviceId,
		PartialState:      partialState,
		ImportServiceId:   check(serviceId, partialState.ServiceAccounts),
		DeferredTransfers: []types.DeferredTransfer{},
		Exception:         nil,
	}
}

// (B.14)
func check(i types.ServiceId, serviceAccountState types.ServiceAccountState) types.ServiceId {
	for k, _ := range serviceAccountState {
		if k == i {
			var modValue uint32 = (1 << 32) - (1 << 9) // 2^32 - 2^9
			var subValue uint32 = 1 << 8               // 2^8
			result := (uint32(i)-subValue+1)%modValue + subValue
			return check(types.ServiceId(result), serviceAccountState)
		}
	}
	return i
}

type Psi_A_ReturnType struct {
	PartialStateSet   types.PartialStateSet
	DeferredTransfers []types.DeferredTransfer
	Result            *types.OpaqueHash
	Gas               types.Gas
}

// (B.7)
type ResultContext struct {
	ServiceId         types.ServiceId
	PartialState      types.PartialStateSet
	ImportServiceId   types.ServiceId
	DeferredTransfers []types.DeferredTransfer
	Exception         *types.OpaqueHash
}
