package PVM

import "testing"

func TestSetExecutionBackendInterpreter(t *testing.T) {
	// Interpreter hook is registered by packages that blank-import it; this
	// package alone may not. Accept either success or a clear "not linked" error.
	prev := ExecutionBackend
	defer func() { ExecutionBackend = prev }()

	err := SetExecutionBackend(BackendInterpreter)
	if err != nil && Psi_M_interpreterHook != nil {
		t.Fatalf("interpreter linked but SetExecutionBackend failed: %v", err)
	}
	if err == nil && ExecutionBackend != BackendInterpreter {
		t.Fatalf("ExecutionBackend=%q", ExecutionBackend)
	}

	if err := SetExecutionBackend("nope"); err == nil {
		t.Fatal("expected error for unknown backend")
	}
}
