//go:build !trace

package interpreter

import (
	"encoding/json"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func wireInterpreterAccumulateTrace(_ *Host, _ []byte, _ PVM.ProgramCounter, _ types.Gas, _ PVM.Registers, _ PVM.HostCallArgs) func(PVM.Psi_H_ReturnType) {
	return func(PVM.Psi_H_ReturnType) {}
}

func (h *Host) beginAccumulateHostCallTrace() (rin [13]uint64, gIn int64, ok bool) {
	return [13]uint64{}, 0, false
}

func wrapGuestMemoryForHostCallTrace(mem PVM.GuestMemory) (PVM.GuestMemory, func() json.RawMessage) {
	return mem, func() json.RawMessage { return nil }
}

func (h *Host) recordAccumulateHostCallAfterOmega(_ PVM.OperationType, _ [13]uint64, _ int64, _ [13]uint64, _ int64, _ bool, _ uint64, _ PVM.ExitReason, _ json.RawMessage) {
}
