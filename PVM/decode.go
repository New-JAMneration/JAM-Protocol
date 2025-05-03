package PolkaVM

import (
	"fmt"
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
)

func getRegModIndex(instructionCode []byte, pc ProgramCounter) uint8 {
	return min(12, (instructionCode[pc+1])%16)
}

func getRegFloorIndex(instructionCode []byte, pc ProgramCounter) uint8 {
	return min(12, (instructionCode[pc+1])>>4)
}

// A.5.2
func decodeOneImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, error) {
	lX := min(4, skipLength)
	immediateData := instructionCode[pc+1 : pc+lX]
	immediate, _, err := ReadUintFixed(immediateData, len(immediateData))
	if err != nil {
		return 0, err
	}
	return int(immediate), nil
}

// A.5.3
func decodeOneRegisterAndOneExtendedWidthImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, uint64, error) {
	panic("not implemented")
}

// A.5.4
func decodeTwoImmediates(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (uint64, uint64, error) {
	lX := ProgramCounter(min(4, uint8(instructionCode[pc+1])))

	decodedVX, err := utils.DeserializeFixedLength(instructionCode[pc+2:pc+2+lX], types.U64(lX))
	if err != nil {
		return 0, 0, fmt.Errorf("opcode %s(%d) at pc=%d deserialize vx raise error : %w", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
	}

	vX, err := SignExtend(int(lX), uint64(decodedVX))
	if err != nil {
		return 0, 0, fmt.Errorf("opcosde %s(%d) at pc=%d signExtend lx raise error : %w", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
	}
	lY := min(4, max(0, skipLength-lX-1))
	decodedVy, err := utils.DeserializeFixedLength(instructionCode[pc+2+lX:pc+2+lX+lY], types.U64(lY))
	if err != nil {
		return 0, 0, fmt.Errorf("opcosde %s(%d) at pc=%d deserialization vy raise error : %w", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
	}
	vY, err := SignExtend(int(lY), uint64(decodedVy))
	if err != nil {
		return 0, 0, fmt.Errorf("opcosde %s(%d) at pc=%d signExtend lx raise error : %w", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
	}

	return vX, vY, nil
}

// A.5.5
// returns vX
func decodeOneOffset(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (ProgramCounter, error) {
	lX := min(4, skipLength)
	offsetData := instructionCode[pc+1 : pc+1+lX]
	offset, _, err := ReadIntFixed(offsetData, len(offsetData))
	if err != nil {
		return 0, err
	}

	return pc + ProgramCounter(offset), nil
}

// A.5.6
// returns rA, vX
func decodeOneRegisterAndOneImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (uint8, uint64, error) {
	rA := min(12, instructionCode[pc+1]%16)
	lX := min(4, max(0, skipLength-1))

	immediateData := instructionCode[pc+2 : pc+2+lX]
	immediate, _, err := ReadUintSignExtended(immediateData, len(immediateData))
	if err != nil {
		log.Printf("opcode %s at instruction %d deserialize vy raise error : %s", zeta[opcode(instructionCode[pc])], pc, err)
		return 0, 0, err
	}

	return rA, immediate, nil
}

// A.5.7
func decodeOneRegisterAndTwoImmediates(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int8, uint64, uint64, error) {
	rA := int8(min(12, instructionCode[pc+1]%16))
	lX := min(4, ProgramCounter(uint8((instructionCode[pc+1] >> 4))))
	pcMargin := pc + 2 + lX
	decodedVX, err := utils.DeserializeFixedLength(instructionCode[pc+2:pcMargin], types.U64(lX))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("opcode %s(%d) at pc=%d deserialize vx raise error : %w", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
	}
	vX, err := SignExtend(int(lX), uint64(decodedVX))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("opcode %s(%d) at pc=%d signExtend vx raise error : %w", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
	}

	lY := min(4, max(0, skipLength-lX-1))
	decodedVY, err := utils.DeserializeFixedLength(instructionCode[pcMargin:pcMargin+lY], types.U64(lY))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("opcode %s(%d) at pc=%d deserialize vy raise error : %w", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
	}
	vY, err := SignExtend(int(lY), uint64(decodedVY))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("opcode %s(%d) at pc=%d signExtend vy raise error : %w", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
	}

	return rA, vX, vY, nil
}

// A.5.8
// returns rA, vX, vY
func decodeOneRegisterOneImmediateAndOneOffset(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (uint8, uint64, ProgramCounter, error) {
	rA := min(12, instructionCode[pc+1]%16)
	lX := ProgramCounter(min(4, (instructionCode[pc+1]>>4)%8))
	lY := min(4, max(0, skipLength-lX-1))

	immediateData := instructionCode[pc+2 : pc+2+lX]
	immediate, _, err := ReadUintSignExtended(immediateData, len(immediateData))
	if err != nil {
		return 0, 0, 0, err
	}

	offsetData := instructionCode[pc+2+lX : pc+2+lX+lY]
	offset, _, err := ReadIntFixed(offsetData, len(offsetData))
	if err != nil {
		return 0, 0, 0, err
	}

	return rA, immediate, pc + ProgramCounter(offset), nil
}

// A.5.9
func decodeTwoRegisters(instructionCode []byte, pc ProgramCounter) (rD uint8, rA uint8, err error) {
	if int(pc+1) >= len(instructionCode) {
		return 0, 0, fmt.Errorf("pc out of bound")
	}
	rD = getRegModIndex(instructionCode, pc)
	rA = getRegFloorIndex(instructionCode, pc)
	return rD, rA, nil
}

func decodeTwoRegistersAndOneImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (uint8, uint8, uint64, error) {
	rA := min(12, instructionCode[pc+1]&15)
	rB := min(12, instructionCode[pc+1]>>4)
	lX := min(4, max(0, skipLength-1))
	decodedVX, err := utils.DeserializeFixedLength(instructionCode[pc+2:pc+2+lX], types.U64(lX))
	if err != nil {
		return 0, 0, 0, PVMExitTuple(PANIC, nil)
	}
	vX, err := SignExtend(int(lX), uint64(decodedVX))
	if err != nil {
		return 0, 0, 0, PVMExitTuple(PANIC, nil)
	}

	return rA, rB, vX, nil
}

// A.5.11
// returns rA, rB, vX
func decodeTwoRegistersAndOneOffset(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (uint8, uint8, ProgramCounter, error) {
	rA := min(12, instructionCode[pc+1]%16)
	rB := min(12, instructionCode[pc+1]>>4)
	lX := min(4, max(0, skipLength-1))

	offsetData := instructionCode[pc+2 : pc+2+lX]
	offset, _, err := ReadIntFixed(offsetData, len(offsetData))
	if err != nil {
		return 0, 0, 0, err
	}

	return rA, rB, pc + ProgramCounter(offset), nil
}

// A.5.12
// returns rA, rB, vX, vY
func decodeTwoRegistersAndTwoImmediates(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (uint8, uint8, uint64, uint64, error) {
	rA := min(12, instructionCode[pc+1]%16)
	rB := min(12, instructionCode[pc+1]>>4)
	lX := ProgramCounter(min(4, instructionCode[pc+2]%8))
	lY := min(4, max(0, skipLength-lX-2))

	vXData := instructionCode[pc+3 : pc+3+lX]
	vX, _, err := ReadUintFixed(vXData, len(vXData))
	if err != nil {
		return 0, 0, 0, 0, err
	}

	vYData := instructionCode[pc+3+lX : pc+3+lX+lY]
	vY, _, err := ReadUintFixed(vYData, len(vYData))
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return rA, rB, vX, vY, nil
}

// A.5.13
func decodeThreeRegisters(instructionCode []byte, pc ProgramCounter) (rA uint8, rB uint8, rD uint8, err error) {
	if int(pc+2) >= len(instructionCode) {
		return 0, 0, 0, fmt.Errorf("pc out of bound")
	}
	rA = getRegModIndex(instructionCode, pc)
	rB = getRegFloorIndex(instructionCode, pc)
	rD = min(12, instructionCode[pc+2])
	return rA, rB, rD, nil
}

func storeIntoMemory(mem Memory, offset int, memIndex uint32, Immediate uint64) error {
	vX := uint32(memIndex)
	pageNum := vX / ZP
	pageIndex := memIndex % ZP

	vY := utils.SerializeFixedLength(types.U64(Immediate), types.U64(offset))
	if _, pageAllocated := mem.Pages[pageNum]; pageAllocated {
		// try to allocate read-only memory --> page-fault
		if mem.Pages[pageNum].Access != MemoryReadWrite {
			return PVMExitTuple(PAGE_FAULT, memIndex)
		}
		if pageIndex+uint32(offset) <= ZP { // data allocated do not exceed maxSize
			copy(mem.Pages[pageNum].Value[pageIndex:], vY)
		} else { // data allocated exceed maxSize --> cross page
			// check next page access
			if _, pageAllocated := mem.Pages[pageNum+1]; !pageAllocated {
				return PVMExitTuple(PAGE_FAULT, memIndex)
			}

			currentPageLength := ZP - pageIndex
			currentPageData := vY[:currentPageLength]
			nextPageData := vY[currentPageLength:]

			copy(mem.Pages[pageNum].Value[pageIndex:], currentPageData)
			copy(mem.Pages[pageNum].Value[:len(nextPageData)], nextPageData)
		}

	} else { // page not allocated
		return PVMExitTuple(PAGE_FAULT, memIndex)
	}
	return PVMExitTuple(CONTINUE, nil)
}

func loadFromMemory(mem Memory, offset uint32, vx uint32) (uint64, error) {
	vX := uint32(vx)

	pageNum := vX / ZP
	pageIndex := vX % ZP
	// load memory : the page must be exist and at least readable
	// we allocated memory at least read-only
	// => if the memory is not allocated -> it's Inaccessible
	if _, pageAllocated := mem.Pages[pageNum]; !pageAllocated {
		return 0, PVMExitTuple(PAGE_FAULT, vx)
	}
	memBytes := make([]byte, offset)

	if pageIndex+offset <= ZP {
		memBytes = mem.Pages[pageNum].Value[pageIndex : pageIndex+offset]
	} else { // cross page memory memory loading
		if mem.Pages[pageNum+1] == nil {
			return 0, PVMExitTuple(PAGE_FAULT, vx)
		}

		remainBytes := mem.Pages[pageNum+1].Value[:offset-(ZP-pageIndex)]
		copy(memBytes, mem.Pages[pageNum].Value[pageIndex:]) // copy current page
		copy(memBytes[:ZP-pageIndex], remainBytes)           // copy next page
	}
	memVal, err := utils.DeserializeFixedLength(memBytes, types.U64(offset))
	if err != nil {
		return 0, PVMExitTuple(PANIC, nil)
	}
	return uint64(memVal), nil
}
