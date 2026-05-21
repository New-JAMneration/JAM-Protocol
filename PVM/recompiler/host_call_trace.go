//go:build linux && amd64 && trace

package recompiler

import (
	"encoding/json"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
)

func wrapGuestMemoryForHostCallTrace(ctx *JITContext) (PVM.GuestMemory, func() json.RawMessage) {
	detailsFn, wrapped := PVM.BeginHostCallMemTrace(ctx.GuestMemory())
	return wrapped, detailsFn
}
