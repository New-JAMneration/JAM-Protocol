package PolkaVM

import (
	"fmt"
)

type ExitReasonTypes int

const (
	CONTINUE ExitReasonTypes = iota
	HALT
	PANIC
	OUT_OF_GAS
	PAGE_FAULT
	HOST_CALL
)

var exitMessages = map[ExitReasonTypes]string{
	CONTINUE:   "Continue (▸)",
	HALT:       "Regular halt (∎)",
	PANIC:      "Panic (☇)",
	OUT_OF_GAS: "Out-Of-Gas (∞)",
	PAGE_FAULT: "Page fault (F)",
	HOST_CALL:  "Host-Call identifier (̵h)",
}

// varepsilon ε
type PVMExitReason struct {
	Reason    ExitReasonTypes // exit types: "HALT", "PANIC", "OUT_OF_GAS", "PAGE_FAULT", "HOST_CALL"
	HostCall  *string         // if Type = "HOST_CALL", store Host-Call identifier
	FaultAddr *uint64         // if Type = "PAGE_FAULT", store wrong RAM address
}

// error interface
func (e *PVMExitReason) Error() string {
	msg, exists := exitMessages[e.Reason]
	if !exists {
		msg = "Unknown exit reason"
	}

	switch e.Reason {
	case PAGE_FAULT:
		if e.FaultAddr != nil {
			return fmt.Sprintf("%s at RAM address: %d", msg, *e.FaultAddr)
		}
	case HOST_CALL:
		if e.HostCall != nil {
			return fmt.Sprintf("%s: %s", msg, *e.HostCall)
		}
	}
	return msg
}

// PVMExitTuple can handle varepsilon type with (reason, meta)
func PVMExitTuple(reason ExitReasonTypes, meta interface{}) error {
	switch reason {
	case PAGE_FAULT:
		if addr, ok := meta.(uint64); ok { // uint64 types may change in the future
			return &PVMExitReason{Reason: PAGE_FAULT, FaultAddr: &addr}
		}
	case HOST_CALL:
		if call, ok := meta.(string); ok { // string types may change in the future
			return &PVMExitReason{Reason: HOST_CALL, HostCall: &call}
		}
	}
	return &PVMExitReason{Reason: reason}
}
