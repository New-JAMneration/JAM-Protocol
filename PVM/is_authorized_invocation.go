package PVM

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func Psi_I(p types.WorkPackage, c types.CoreIndex, code types.ByteSequence) Psi_I_ReturnType {
	serialized := []byte{}
	encoder := types.NewEncoder()

	encoded, err := encoder.Encode(&p)
	if err != nil {
		panic(err)
	}
	serialized = append(serialized, encoded...)

	encoded, err = encoder.Encode(&c)
	if err != nil {
		panic(err)
	}
	serialized = append(serialized, encoded...)

	F := Omegas{}
	F[GasOp] = HostCallFunctions[GasOp]
	F[27] = accumulateInvocationHostCallException

	addition := HostCallArgs{}

	resultM := Psi_M(StandardCodeFormat(code), 0, types.IsAuthorizedGas, Argument(serialized), F, addition)
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
