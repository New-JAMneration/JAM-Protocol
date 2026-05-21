package accumulation

// Ensure the default PVM execution backend (interpreter) is linked: PVM.Psi_M
// panics if no backend is registered (see PVM/interpreter and PVM.Psi_M). The
// interpreter is cross-platform and always available; the recompiler
// (linux/amd64) is linked separately and selected at runtime via
// PVM.ExecutionBackend. This package calls PVM.Psi_A, so the backend must be
// linked here for both production binaries and this package's own tests.
import _ "github.com/New-JAMneration/JAM-Protocol/PVM/interpreter"
