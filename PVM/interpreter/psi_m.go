package interpreter

import (
	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// init registers the interpreter backend so PVM.Psi_M can dispatch to it.
// Every binary that runs the PVM must blank-import this package.
func init() {
	PVM.Psi_M_interpreterHook = psiMInterpreter
}

func psiMInterpreter(
	code PVM.StandardCodeFormat,
	counter PVM.ProgramCounter,
	gas types.Gas,
	argument PVM.Argument,
	omegas PVM.Omegas,
	addition PVM.HostCallArgs,
) (
	psi_result PVM.Psi_M_ReturnType,
) {
	programCode, registers, memory, exitReason := PVM.SingleInitializer(code, argument)
	// Y(p) = nil
	if exitReason != PVM.ExitContinue {
		return PVM.Psi_M_ReturnType{
			Gas:           0,
			ReasonOrBytes: PVM.ExitPanic,
			Addition:      addition,
		}
	}

	program, exitReason := PVM.DeBlobProgramCode(programCode)
	if exitReason != PVM.ExitContinue {
		return PVM.Psi_M_ReturnType{
			Gas:           0,
			ReasonOrBytes: PVM.ExitPanic,
			Addition:      addition,
		}
	}

	addition.Program = &program

	host := NewHost(&program, registers, &memory, PVM.Gas(gas), addition, omegas)
	closeTrace := wireInterpreterAccumulateTrace(host, programCode, counter, gas, registers, addition)
	psiHResult := host.HostCall(counter, 0)
	closeTrace(psiHResult)

	g, v, a := PVM.R(gas, psiHResult)
	return PVM.Psi_M_ReturnType{
		Gas:           types.Gas(g),
		ReasonOrBytes: v,
		Addition:      a,
	}
}
