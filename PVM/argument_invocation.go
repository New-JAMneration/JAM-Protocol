package PVM

import (
	"math/bits"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

const (
	BackendInterpreter = "interpreter"
	BackendRecompiler  = "recompiler"
)

// ExecutionBackend selects PVM execution backend at runtime.
var ExecutionBackend = BackendInterpreter

// PsiMBackend is the contract every PVM execution backend implements: given a
// program plus entry state, run it to completion and return the finalised
// Psi_M result. Both the interpreter and the recompiler register one of these;
// Psi_M dispatches to the selected backend. Declaring it as a function pointer
// keeps core free of any backend import — the hooks break the import cycle.
type PsiMBackend func(
	code StandardCodeFormat,
	counter ProgramCounter,
	gas types.Gas,
	argument Argument,
	omegas Omegas,
	addition HostCallArgs,
) Psi_M_ReturnType

var (
	// Psi_M_interpreterHook is registered by the interpreter backend
	// (PVM/interpreter) in its init(). It is the default / fallback and is
	// expected to always be linked: every binary that runs the PVM must
	// blank-import PVM/interpreter.
	Psi_M_interpreterHook PsiMBackend
	// Psi_M_recompilerHook is registered by PVM/recompiler when linked
	// (linux/amd64 only).
	Psi_M_recompilerHook PsiMBackend
)

// (A.40) Ψ_M
func Psi_M(
	code StandardCodeFormat,
	counter ProgramCounter, // program counter
	gas types.Gas, // gas counter
	argument Argument, // argument
	omegas Omegas, // jump table
	addition HostCallArgs, // host-call context
) (
	psi_result Psi_M_ReturnType,
) {
	// Recompiler when selected and linked; otherwise fall back to interpreter
	// (e.g. recompiler requested on a build that did not link it).
	if ExecutionBackend == BackendRecompiler && Psi_M_recompilerHook != nil {
		return Psi_M_recompilerHook(code, counter, gas, argument, omegas, addition)
	}
	if Psi_M_interpreterHook != nil {
		return Psi_M_interpreterHook(code, counter, gas, argument, omegas, addition)
	}
	// Unreachable in a correctly built binary (the interpreter backend is always
	// linked). If it is reachable, no backend is registered: PVM can't run and
	// any result would be wrong anyway, so fail loud instead of returning junk.
	panic("PVM.Psi_M: no execution backend registered (interpreter backend not linked)")
}

// (A.41) R
func R(priorGas types.Gas, Psi_H_Return Psi_H_ReturnType) (Gas, any, HostCallArgs) {
	u := priorGas - types.Gas(max(*Psi_H_Return.VM.Gas, 0))

	switch Psi_H_Return.ExitReason.GetReasonType() {
	case OUT_OF_GAS:
		return Gas(u), OUT_OF_GAS, Psi_H_Return.Addition
	case HALT:
		start := uint64(Psi_H_Return.VM.Registers[7])
		length := uint64(Psi_H_Return.VM.Registers[8])
		// Read the return slice [r7, r7+r8) through the GuestMemory abstraction;
		// both backends set VM.Mem (interpreter: paged, recompiler: segment-aware
		// flat). Read returns a fresh copy safe to return up the stack.
		mem := Psi_H_Return.VM.Mem
		if mem.IsReadable(start, length) {
			if length == 0 {
				return Gas(u), nil, Psi_H_Return.Addition
			}
			return Gas(u), mem.Read(start, length), Psi_H_Return.Addition
		}
		return Gas(u), []byte{}, Psi_H_Return.Addition
	default:
		return Gas(u), PANIC, Psi_H_Return.Addition
	}
}

// checkOverflow returns a*b and reports whether the multiplication overflowed uint64.
func checkOverflow(a, b uint64) (uint64, bool) {
	hi, lo := bits.Mul64(a, b)
	return lo, hi != 0
}

func isReadable(start, offset uint64, m Memory) bool {
	if offset == 0 {
		return true
	}
	// overflow + PVM RAM (2^32) bound check
	if offset > (1<<32) || start > (1<<32)-offset {
		return false
	}
	startPage := uint32(start / ZP)
	endPage := uint32((start + offset - 1) / ZP)

	for p := startPage; p <= endPage; p++ {
		if m.GetPageAccess(p) == MemoryInaccessible {
			return false
		}
	}
	return true
}

func isWriteable(start, offset uint64, m Memory) bool {
	if offset == 0 {
		return true
	}
	// overflow + PVM RAM (2^32) bound check
	if offset > (1<<32) || start > (1<<32)-offset {
		return false
	}
	startPage := uint32(start / ZP)
	endPage := uint32((start + offset - 1) / ZP)

	for p := startPage; p <= endPage; p++ {
		if m.GetPageAccess(p) != MemoryReadWrite {
			return false
		}
	}
	return true
}

type Psi_M_ReturnType struct {
	Gas           types.Gas
	ReasonOrBytes any
	Addition      HostCallArgs
}
