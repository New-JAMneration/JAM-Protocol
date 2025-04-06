package PolkaVM

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func Psi_I(p types.WorkPackage, c types.CoreIndex, code types.OpaqueHash) Psi_I_ReturnType {

	serialized := []byte{}
	encoder := types.NewEncoder()

	// todo 改寫法
	encoded_p, err := encoder.Encode(&p)
	if err != nil {
		panic(err)
	}

	encoded_c, err := encoder.Encode(&c)
	if err != nil {
		panic(err)
	}

	serialized = append(serialized, encoded_p...)
	serialized = append(serialized, encoded_c...)

	F := Omegas{}
	F[GasOp] = hostCallFunctions[GasOp]
	F[27] = accumulateInvocationHostCallException

	// 准备HostCallArgs
	// 參數全空
	addition := HostCallArgs{
		// todo: 看灰皮書
		GeneralArgs: GeneralArgs{
			ServiceAccount:      types.ServiceAccount{},
			ServiceId:           types.ServiceId(0),
			ServiceAccountState: types.ServiceAccountState{},
		},
		AccumulateArgs: AccumulateArgs{
			ResultContextX: ResultContext{},
			ResultContextY: ResultContext{},
		},
	}

	// todo: 找 gas
	resultM := Psi_M(code[:], 0, Gas(types.IsAuthorizedGas), Argument(serialized), F, addition)
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

	// 斷言 resultM.ReasonOrBytes 是 []byte
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
	// todo: R 的返回值
	WorkOutput []byte
	Gas        types.Gas
}
