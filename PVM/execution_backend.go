package PVM

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// SetExecutionBackend validates and assigns the process-global ExecutionBackend
// used by Psi_M. Prefer assigning once at process start; do not toggle after
// concurrent Psi_M callers exist. Dual-backend tests should use Psi_M_OnBackend
// instead of swapping this global.
func SetExecutionBackend(backend string) error {
	if _, err := psiMHook(backend); err != nil {
		return err
	}
	ExecutionBackend = backend
	return nil
}

// Psi_M_OnBackend runs Ψ_M on an explicit backend without mutating
// ExecutionBackend. Prefer this in dual-backend tests over temporarily
// swapping the process-global selector.
func Psi_M_OnBackend(
	backend string,
	code StandardCodeFormat,
	counter ProgramCounter,
	gas types.Gas,
	argument Argument,
	omegas Omegas,
	addition HostCallArgs,
) (Psi_M_ReturnType, error) {
	hook, err := psiMHook(backend)
	if err != nil {
		return Psi_M_ReturnType{}, err
	}
	return hook(code, counter, gas, argument, omegas, addition), nil
}

func psiMHook(backend string) (PsiMBackend, error) {
	switch backend {
	case BackendInterpreter:
		if Psi_M_interpreterHook == nil {
			return nil, fmt.Errorf("pvm backend %q is not linked", backend)
		}
		return Psi_M_interpreterHook, nil
	case BackendRecompiler:
		if Psi_M_recompilerHook == nil {
			return nil, fmt.Errorf("pvm backend %q is not available (requires linux/amd64 with cgo and recompiler linked)", backend)
		}
		return Psi_M_recompilerHook, nil
	default:
		return nil, fmt.Errorf("pvm backend must be %q or %q, got %q",
			BackendInterpreter, BackendRecompiler, backend)
	}
}
