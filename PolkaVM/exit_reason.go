package PolkaVM

import (
	"fmt"
	"math"
<<<<<<< HEAD
	"sort"
=======
>>>>>>> main
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
	HostCall  *uint64         // if Type = "HOST_CALL", store Host-Call identifier
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
			return fmt.Sprintf("%s: %d", msg, *e.HostCall)
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
		if call, ok := meta.(uint64); ok { // string types may change in the future
			return &PVMExitReason{Reason: HOST_CALL, HostCall: &call}
		}
	}
	return &PVMExitReason{Reason: reason}
}

<<<<<<< HEAD
// (A.9) ParseMemoryAccessError parses the memory access error based on the given
// invalid addresses.
func ParseMemoryAccessError(invalidAddresses []uint64) (ExitReasonTypes, error) {
	for i := range invalidAddresses {
		invalidAddresses[i] = invalidAddresses[i] % (1 << 32)
	}
	// Iterate over read addresses and check for errors.0
	if len(invalidAddresses) == 0 {
		return CONTINUE, nil
	}

	minAddress := uint32(math.MaxUint32)
	for _, addr := range invalidAddresses {
		if addr < ZZ {
			return PANIC, nil
		}
		if uint32(addr) < minAddress {
			minAddress = uint32(addr)
		}
	}
	return PAGE_FAULT, PVMExitTuple(PAGE_FAULT, minAddress/ZP)
}

// (A.8) get invalid address // TODO design/align with 4.26 4.27
func GetInvalidAddress(readAddresses []uint64, writeAddresses []uint64, readableAddresses map[int]bool, writeableAddresses map[int]bool) []uint64 {
	var invalidAddresses []uint64
	for _, addr := range readAddresses {
		if !readableAddresses[int(addr)/ZP] {
			invalidAddresses = append(invalidAddresses, addr)
		}
	}
	for _, addr := range writeAddresses {
		if !writeableAddresses[int(addr)/ZP] {
			invalidAddresses = append(invalidAddresses, addr)
		}
	}
	// sort + unique
	sort.Slice(invalidAddresses, func(i, j int) bool {
		return invalidAddresses[i] < invalidAddresses[j]
	})
	uniqueAddresses := []uint64{}
	for i, addr := range invalidAddresses {
		if i > 0 && invalidAddresses[i] == invalidAddresses[i-1] {
			continue
		}
		uniqueAddresses = append(uniqueAddresses, addr)
	}

	return uniqueAddresses
=======
// Branch implements the branch function (A.17)
func Branch(pc ProgramCounter, offset uint32, condition bool, basicBlocks []uint32) (ExitReasonTypes, ProgramCounter) {
	// instructions table will define different offset
	target := pc + ProgramCounter(offset)
	if !condition {
		return CONTINUE, pc // Condition is false, continue execution at next instruction.
	}
	if !IsBasicBlock(uint32(target), basicBlocks) {
		return PANIC, pc // Target is not basic block, panic.
	}
	return CONTINUE, target // Otherwise, jump to the target.
}

// djump implements the dynamic jump function (A.18)
func Djump(target uint32, jumpTable []uint32, basicBlocks []uint32) (ExitReasonTypes, uint32) {
	if target == math.MaxUint32-ZZ+1 {
		return HALT, target // Special case: return the address as is. // TODO design for returning iota
	}
	if target == 0 || target > uint32(len(jumpTable))*ZA || target%ZA != 0 || !IsBasicBlock(jumpTable[target/ZA], basicBlocks) {
		return PANIC, target // If the target is invalid, panic. // TODO design for returning iota
	}
	return CONTINUE, jumpTable[target/ZA] // Otherwise, jump to the target.
}

// contains checks if a slice contains a specific value.
func IsBasicBlock(target uint32, basicBlocks []uint32) bool {
	for _, basicBlock := range basicBlocks {
		if target == basicBlock {
			return true
		}
	}
	return false
>>>>>>> main
}
