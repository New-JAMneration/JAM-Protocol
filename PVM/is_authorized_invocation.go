package PVM

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// Ψ_I is the is-authorized entry point.
//
// Backend requirement: this runs the PVM (via Psi_M), which dispatches to a
// registered execution backend — it does not execute anything itself. The
// calling package MUST link a backend by blank-importing one so its init()
// registers the hook:
//
//	import _ ".../PVM/interpreter"  // cross-platform default; always available
//	import _ ".../PVM/recompiler"   // optional, linux/amd64 only
//
// The backend is selected at runtime via PVM.ExecutionBackend (interpreter is
// the default and fallback). With no backend registered, Psi_M panics.
func Psi_I(p types.WorkPackage, c types.CoreIndex, authorizerCode types.ByteSequence) Psi_I_ReturnType {
	if len(authorizerCode) == 0 {
		return Psi_I_ReturnType{
			WorkExecResult: types.WorkExecResultBadCode,
			Gas:            0,
		}
	} else if len(authorizerCode) > types.MaxIsAuthorizedCodeSize {
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

	addition := HostCallArgs{
		GeneralArgs: GeneralArgs{
			ServiceID: nil,
			CoreID:    &c,
		},
	}

	resultM := Psi_M(StandardCodeFormat(authorizerCode), 0, types.IsAuthorizedGas, Argument(encoded), IsAuthorizedOmegas, addition)
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
