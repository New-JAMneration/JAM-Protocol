package PVMtrace

import "encoding/json"

// TraceInfo is the metadata written to meta/info.json for each trace invocation.
type TraceInfo struct {
	FormatVersion    int            `json:"format_version"`
	GraypaperVersion string         `json:"graypaper_version,omitempty"`
	Backend          string         `json:"backend"`
	TraceMode        string         `json:"trace_mode"`
	ServiceID        uint32         `json:"service_id"`
	CodeHash         string         `json:"codehash"`
	Timeslot         uint64         `json:"timeslot,omitempty"`
	RunID            string         `json:"run_id,omitempty"`
	InvocationType   string         `json:"invocation_type,omitempty"`
	InitialPC        uint64         `json:"initial_pc"`
	InitialGas       int64          `json:"initial_gas"`
	InitialRegs      [13]uint64     `json:"initial_regs,omitempty"`
	FinalPC          uint64         `json:"final_pc"`
	FinalGas         int64          `json:"final_gas"`
	FinalExitReason  string         `json:"final_exit_reason"`
	FinalRegs        [13]uint64     `json:"final_regs,omitempty"`
	TotalSteps       int64          `json:"total_steps"`
	Truncated        bool           `json:"truncated"`
	Streams          []string       `json:"streams,omitempty"`
}

// HostCallRecord is a single host-call boundary record written to host_calls.jsonl.gz.
type HostCallRecord struct {
	Step       int            `json:"step"`
	PC         uint64         `json:"pc"`
	Op         uint64         `json:"op"`
	OpName     string         `json:"op_name"`
	RegsIn     [13]uint64     `json:"regs_in"`
	GasIn      int64          `json:"gas_in"`
	RegsOut    [13]uint64     `json:"regs_out"`
	GasOut     int64          `json:"gas_out"`
	ExitReason string         `json:"exit_reason"`
	Details    json.RawMessage `json:"details,omitempty"`
}

// HostCallDetails contains common fields for all host-call detail records.
type HostCallDetails struct {
	MemReads  []MemAccess       `json:"memreads,omitempty"`
	MemWrites []MemAccess       `json:"memwrites,omitempty"`
	SetRegs   map[string]string `json:"setregs,omitempty"`
	SetGas    *uint64           `json:"setgas,omitempty"`
}

// MemAccess represents a memory read or write during a host-call.
type MemAccess struct {
	Addr string `json:"addr"`
	Len  int    `json:"len"`
	Ok   bool   `json:"ok"`
	Data string `json:"data,omitempty"`
}
