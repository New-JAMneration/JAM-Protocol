//go:build !pvmtrace

package PVM

import (
	"encoding/json"
)

// BeginHostCallMemTrace is a no-op when built without -tags=pvmtrace.
func BeginHostCallMemTrace(inner GuestMemory) (detailsFn func() json.RawMessage, wrapped GuestMemory) {
	return func() json.RawMessage { return nil }, inner
}
