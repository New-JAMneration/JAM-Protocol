package PVM

import (
	"fmt"
)

// SetExecutionBackend selects the PVM execution backend for subsequent Psi_M
// calls. Unlike assigning ExecutionBackend directly, this fails if the backend
// is unknown or not linked (no silent fallback to interpreter).
func SetExecutionBackend(backend string) error {
	switch backend {
	case BackendInterpreter:
		if Psi_M_interpreterHook == nil {
			return fmt.Errorf("pvm backend %q is not linked", backend)
		}
		ExecutionBackend = BackendInterpreter
		return nil
	case BackendRecompiler:
		if Psi_M_recompilerHook == nil {
			return fmt.Errorf("pvm backend %q is not available (requires linux/amd64 with cgo and recompiler linked)", backend)
		}
		ExecutionBackend = BackendRecompiler
		return nil
	default:
		return fmt.Errorf("pvm backend must be %q or %q, got %q",
			BackendInterpreter, BackendRecompiler, backend)
	}
}

// WithExecutionBackend runs fn with the given backend selected, then restores
// the previous ExecutionBackend. Useful for tests that must exercise both
// backends without leaking global state.
func WithExecutionBackend(backend string, fn func()) error {
	prev := ExecutionBackend
	if err := SetExecutionBackend(backend); err != nil {
		return err
	}
	defer func() { ExecutionBackend = prev }()
	fn()
	return nil
}
