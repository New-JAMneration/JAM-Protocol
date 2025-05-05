package PolkaVM

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// (B.8) Ψ_A
func Psi_A(
	partialState types.PartialStateSet,
	time types.TimeSlot,
	serviceId types.ServiceId,
	gas types.Gas,
	operands []types.Operand,
	eta types.Entropy,
) (
	psi_result Psi_A_ReturnType,
) {
	c := partialState.ServiceAccounts[serviceId].ServiceInfo.CodeHash
	if storedCode, ok := partialState.ServiceAccounts[serviceId].PreimageLookup[c]; !ok {
		return Psi_A_ReturnType{
			PartialStateSet:   partialState,
			DeferredTransfers: []types.DeferredTransfer{},
			Result:            nil,
			Gas:               0,
			ServiceBlobs:      []ServiceBlob{},
		}
	} else {

		serialized := []byte{}
		encoder := types.NewEncoder()

		// Encode the t
		encoded, err := encoder.Encode(&time)
		if err != nil {
			panic(err)
		}
		serialized = append(serialized, encoded...)

		// Encode the s
		encoded, err = encoder.Encode(&serviceId)
		if err != nil {
			panic(err)
		}
		serialized = append(serialized, encoded...)

		// Encode the o
		// Encode the length of the operands
		encoded, err = encoder.EncodeUint(uint64(len(operands)))
		if err != nil {
			panic(err)
		}
		serialized = append(serialized, encoded...)

		for _, operand := range operands {
			encoded, err = encoder.Encode(&operand)
			if err != nil {
				panic(err)
			}
			serialized = append(serialized, encoded...)
		}

		F := Omegas{}
		F[ReadOp] = wrapWithG(HostCallFunctions[ReadOp])
		F[WriteOp] = wrapWithG(HostCallFunctions[WriteOp])
		F[LookupOp] = wrapWithG(HostCallFunctions[LookupOp])
		F[GasOp] = HostCallFunctions[GasOp]
		F[InfoOp] = wrapWithG(HostCallFunctions[InfoOp])
		F[BlessOp] = HostCallFunctions[BlessOp]
		F[AssignOp] = HostCallFunctions[AssignOp]
		F[DesignateOp] = HostCallFunctions[DesignateOp]
		F[CheckpointOp] = HostCallFunctions[CheckpointOp]
		F[NewOp] = HostCallFunctions[NewOp]
		F[UpgradeOp] = HostCallFunctions[UpgradeOp]
		F[TransferOp] = HostCallFunctions[TransferOp]
		F[EjectOp] = HostCallFunctions[EjectOp]
		F[QueryOp] = HostCallFunctions[QueryOp]
		F[SolicitOp] = HostCallFunctions[SolicitOp]
		F[ForgetOp] = HostCallFunctions[ForgetOp]
		F[YieldOp] = HostCallFunctions[YieldOp]
		F[ProvideOp] = HostCallFunctions[ProvideOp]
		F[OperationType(len(HostCallFunctions)-1)] = accumulateInvocationHostCallException

		addition := HostCallArgs{
			GeneralArgs: GeneralArgs{
				ServiceAccount:      partialState.ServiceAccounts[serviceId],
				ServiceId:           serviceId,
				ServiceAccountState: partialState.ServiceAccounts,
			},
			AccumulateArgs: AccumulateArgs{
				ResultContextX: I(partialState, serviceId, time, eta),
				ResultContextY: I(partialState, serviceId, time, eta),
			},
		}

		resultM := Psi_M(StandardCodeFormat(storedCode), 5, types.Gas(gas), Argument(serialized), F, addition)
		partialState, deferredTransfer, result, gas, serviceBlobs := C(types.Gas(resultM.Gas), resultM.ReasonOrBytes, AccumulateArgs{
			ResultContextX: resultM.Addition.AccumulateArgs.ResultContextX,
			ResultContextY: resultM.Addition.AccumulateArgs.ResultContextY,
		})
		return Psi_A_ReturnType{
			PartialStateSet:   partialState,
			DeferredTransfers: deferredTransfer,
			Result:            result,
			Gas:               gas,
			ServiceBlobs:      serviceBlobs,
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
func C(gas types.Gas, reasonOrBytes any, resultContext AccumulateArgs) (types.PartialStateSet, []types.DeferredTransfer, *types.OpaqueHash, types.Gas, ServiceBlobs) {
	serviceBlobs := make(ServiceBlobs, 0)
	switch reasonOrBytes.(type) {
	case error:
		for _, v := range resultContext.ResultContextY.ServiceBlobs {
			serviceBlobs = append(serviceBlobs, v)
		}
		return resultContext.ResultContextY.PartialState, resultContext.ResultContextY.DeferredTransfers, &resultContext.ResultContextY.Exception, gas, serviceBlobs
	case types.ByteSequence:
		for _, v := range resultContext.ResultContextX.ServiceBlobs {
			serviceBlobs = append(serviceBlobs, v)
		}
		opaqueHash := reasonOrBytes.(*types.OpaqueHash)
		return resultContext.ResultContextX.PartialState, resultContext.ResultContextX.DeferredTransfers, opaqueHash, gas, serviceBlobs
	default:
		for _, v := range resultContext.ResultContextX.ServiceBlobs {
			serviceBlobs = append(serviceBlobs, v)
		}
		return resultContext.ResultContextX.PartialState, resultContext.ResultContextX.DeferredTransfers, &resultContext.ResultContextX.Exception, gas, serviceBlobs
	}
}

// (B.10)
func I(partialState types.PartialStateSet, serviceId types.ServiceId, ht types.TimeSlot, eta types.Entropy) ResultContext {
	serialized := []byte{}
	encoder := types.NewEncoder()

	encoded, err := encoder.Encode(&serviceId)
	if err != nil {
		panic(err)
	}
	serialized = append(serialized, encoded...)
	encoded, err = encoder.Encode(&eta)
	if err != nil {
		panic(err)
	}
	serialized = append(serialized, encoded...)
	encoded, err = encoder.Encode(&ht)
	if err != nil {
		panic(err)
	}
	serialized = append(serialized, encoded...)

	hash := hash.Blake2bHash(serialized)

	var result types.U32
	decoder := types.NewDecoder()
	err = decoder.Decode(hash[:], &result)
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
		Exception:         types.OpaqueHash{},
	}
}

type Psi_A_ReturnType struct {
	PartialStateSet   types.PartialStateSet
	DeferredTransfers []types.DeferredTransfer
	Result            *types.OpaqueHash
	Gas               types.Gas
	ServiceBlobs      []ServiceBlob
}

// (B.7)
type ResultContext struct {
	ServiceId         types.ServiceId                  // s
	PartialState      types.PartialStateSet            // u
	ImportServiceId   types.ServiceId                  // i
	DeferredTransfers []types.DeferredTransfer         // t
	Exception         types.OpaqueHash                 // y
	ServiceBlobs      map[types.OpaqueHash]ServiceBlob // p   v0.6.5
}

// v0.6.5
type ServiceBlob struct {
	ServiceID types.ServiceId
	Blob      []byte
}
type ServiceBlobs []ServiceBlob
