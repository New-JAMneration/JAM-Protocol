//go:build linux && amd64 && !trace

package recompiler

import (
	"encoding/json"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
)

func wrapGuestMemoryForHostCallTrace(ctx *JITContext) (PVM.GuestMemory, func() json.RawMessage) {
	return ctx.GuestMemory(), func() json.RawMessage { return nil }
}
