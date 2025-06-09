package PVM

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func Psi_I(p types.WorkPackage, c types.CoreIndex, authorizerCode *types.ByteSequence) Psi_I_ReturnType {
	if authorizerCode == nil {
		return Psi_I_ReturnType{
			WorkExecResult: types.WorkExecResultBadCode,
			Gas:            0,
		}
	} else if len(*authorizerCode) > types.MaxIsAuthorizedCodeSize {
		return Psi_I_ReturnType{
			WorkExecResult: types.WorkExecResultCodeOversize,
			Gas:            0,
		}
	}

	encoder := types.NewEncoder()
	encoded, err := encoder.Encode(&c)
	if err != nil {
		panic(err)
	}

	F := Omegas{}
	F[GasOp] = HostCallFunctions[GasOp]
	F[FetchOp] = HostCallFunctions[FetchOp] // added 0.6.6
	F[OperationType(len(HostCallFunctions)-1)] = accumulateInvocationHostCallException

	addition := HostCallArgs{}

	resultM := Psi_M(StandardCodeFormat(*authorizerCode), 0, types.IsAuthorizedGas, Argument(encoded), F, addition)
	if resultM.ReasonOrBytes == PANIC {
		return Psi_I_ReturnType{
			WorkExecResult: types.WorkExecResultPanic,
			Gas:            types.Gas(resultM.Gas),
		}
	}
	if resultM.ReasonOrBytes == OUT_OF_GAS {
		return Psi_I_ReturnType{
			WorkExecResult: types.WorkExecResultOutOfGas,
			Gas:            types.Gas(resultM.Gas),
		}
	}

	workOutput, ok := resultM.ReasonOrBytes.([]byte)
	if !ok {
		return Psi_I_ReturnType{
			WorkExecResult: types.WorkExecResultPanic,
			Gas:            types.Gas(resultM.Gas),
		}
	}

	return Psi_I_ReturnType{
		WorkExecResult: types.WorkExecResultOk,
		WorkOutput:     workOutput,
		Gas:            types.Gas(resultM.Gas),
	}
}

type Psi_I_ReturnType struct {
	WorkExecResult types.WorkExecResultType
	WorkOutput     []byte
	Gas            types.Gas
}
