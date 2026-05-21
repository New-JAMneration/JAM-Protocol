// Package PVMtrace provides per-instruction and host-call boundary tracing
// for PVM interpreter and recompiler backends. It produces gzip-compressed
// fixed-width binary streams (pc, opcode, gas, dst_val, src1_val, src2_val,
// loads, stores) plus NDJSON host-call records (host_calls.jsonl.gz).
//
// Controlled by //go:build trace. Without the tag, all methods are no-op stubs.
package PVMtrace

const (
	FormatVersion   = 1
	GraypaperVersion = "0.7.2"

	BackendInterpreter = "interpreter"
	BackendRecompiler  = "recompiler"

	TraceModeNormal          = "normal"
	TraceModeDebugSingleStep = "debug-single-step"
)

// InitialState captures VM state at the start of an invocation.
type InitialState struct {
	PC   uint64
	Gas  int64
	Regs [13]uint64
}

// FinalState captures VM state at the end of an invocation.
type FinalState struct {
	PC         uint64
	Gas        int64
	ExitReason string
	Regs       [13]uint64
}
