//go:build !trace

package PVM

import (
	"encoding/json"
)

// BeginHostCallMemTrace is a no-op when built without -tags=trace.
func BeginHostCallMemTrace(inner GuestMemory) (detailsFn func() json.RawMessage, wrapped GuestMemory) {
	return func() json.RawMessage { return nil }, inner
}
