package PVM

import (
	"fmt"
	"math"
	"sort"
)

type ExitReasonType uint8

const (
	CONTINUE ExitReasonType = iota
	HALT
	PANIC
	OUT_OF_GAS
	PAGE_FAULT
	HOST_CALL
)

// use mask to store the status and value
type ExitReason uint64

const (
	// GP do not define CONTINUE
	// if the actual result code defined in B.1 is needed
	// result =  (res >> 56) -1
	ExitContinue  = ExitReason(CONTINUE)         // 0x00
	ExitHalt      = ExitReason(HALT) << 56       // 0x01
	ExitPanic     = ExitReason(PANIC) << 56      // 0x02
	ExitOOG       = ExitReason(OUT_OF_GAS) << 56 // 0x03
	ExitPageFault = ExitReason(PAGE_FAULT) << 56 // 0x04000000XXXXXXXX
	ExitHostCall  = ExitReason(HOST_CALL) << 56  // 0x050000000000XXXX

)

func (e ExitReason) String() string {
	exitType := e.GetReasonType()

	switch exitType {
	case CONTINUE:
		return "Continue"
	case HALT:
		return "Halt"
	case PANIC:
		return "Panic"
	case OUT_OF_GAS:
		return "Out of Gas"
	case PAGE_FAULT:
		return fmt.Sprintf("Page fault at 0x%08x", e.GetPageFaultAddress())
	case HOST_CALL:
		return fmt.Sprintf("Host Call %d", e.GetHostCallID())
	default:
		return fmt.Sprintf("Unknown ExitReasonType(%d)", exitType)
	}
}

func (e ExitReason) GetReasonType() ExitReasonType {
	return ExitReasonType(e >> 56)
}

func (e ExitReason) GetHostCallID() uint8 {
	return uint8(e)
}

func (e ExitReason) GetPageFaultAddress() uint32 {
	return uint32(e)
}

// Branch implements the branch function (A.17)
func Branch(pc ProgramCounter, offset uint32, condition bool, basicBlocks []uint32) (ExitReasonType, ProgramCounter) {
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
func Djump(target uint32, jumpTable []uint32, basicBlocks []uint32) (ExitReasonType, uint32) {
	if target == math.MaxUint32-ZZ+1 {
		return HALT, target // Special case: return the address as is.
	}
	if target == 0 || target > uint32(len(jumpTable))*ZA || target%ZA != 0 || !IsBasicBlock(jumpTable[target/ZA], basicBlocks) {
		return PANIC, target // If the target is invalid, panic.
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
}

// (A.9) ParseMemoryAccessError parses the memory access error based on the given
// invalid addresses.
func ParseMemoryAccessError(invalidAddresses []uint64) ExitReason {
	for i := range invalidAddresses {
		invalidAddresses[i] = invalidAddresses[i] % (1 << 32)
	}
	// Iterate over read addresses and check for errors.0
	if len(invalidAddresses) == 0 {
		return ExitContinue
	}

	minAddress := uint32(math.MaxUint32)
	for _, addr := range invalidAddresses {
		if addr < ZZ {
			return ExitPanic
		}
		if uint32(addr) < minAddress {
			minAddress = uint32(addr)
		}
	}
	return ExitPageFault | ExitReason(minAddress)
}

// (A.8) get invalid address ß
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
}
