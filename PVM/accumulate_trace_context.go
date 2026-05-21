package PVM

import (
	"os"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// AccumulateTraceContext holds metadata for optional PVM accumulate tracing (Psi_A only).
type AccumulateTraceContext struct {
	ServiceID types.ServiceID
	CodeHash  types.OpaqueHash
	Timeslot  types.TimeSlot
}

// NewAccumulateTraceContextIfEnabled returns trace metadata when JAM_PVM_TRACE_DIR is set.
func NewAccumulateTraceContextIfEnabled(serviceID types.ServiceID, codeHash types.OpaqueHash, timeslot types.TimeSlot) *AccumulateTraceContext {
	if os.Getenv("JAM_PVM_TRACE_DIR") == "" {
		return nil
	}
	return &AccumulateTraceContext{
		ServiceID: serviceID,
		CodeHash:  codeHash,
		Timeslot:  timeslot,
	}
}
