//go:build !trace

package PVMtrace

import "encoding/json"

// Trace is a no-op stub when built without -tags=trace.
type Trace struct{}

func NewTrace(serviceID uint32, codeHash [32]byte, timeslot uint64, programCode []byte, init InitialState, invocationType, backend string) *Trace {
	return nil
}

func (t *Trace) Steps() int64                    { return 0 }
func (t *Trace) RecordStep(pc uint32, opcode byte, dst, src0, src1 uint8, dstVal, src1Val, src2Val uint64, gas int64, loadAddr uint32, loadVal uint64, storeAddr uint32, storeVal uint64) {}
func (t *Trace) RecordBlockEntry(pc uint64, gas int64) {}
func (t *Trace) RecordHostCall(_ HostCallRecord)  {}
func (t *Trace) Close(_ FinalState) error         { return nil }

// suppress unused import
var _ json.RawMessage
