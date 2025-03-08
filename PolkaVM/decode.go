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

func decodeOneImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, error) {
	lX := min(4, skipLength)
	immediateData := instructionCode[pc+1 : pc+lX]
	immediate, _, err := ReadUintFixed(immediateData, len(immediateData))
	if err != nil {
		return 0, err
	}
	return int(immediate), nil
}

func decodeOneRegisterAndOneExtendedWidthImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, uint64, error) {
	panic("not implemented")
}

func decodeTwoImmediates(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, int, error) {
	panic("not implemented")
}

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

// returns rA, vX
func decodeOneRegisterAndOneImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (uint8, uint64, error) {
	rA := min(12, instructionCode[pc+1]%16)
	lX := min(4, max(0, skipLength-1))

	immediateData := instructionCode[pc+2 : pc+2+lX]
	immediate, _, err := ReadUintSignExtended(immediateData, len(immediateData))
	if err != nil {
		return 0, 0, err
	}

	return rA, immediate, nil
}

func decodeOneRegisterAndTwoImmediates(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, int, int, error) {
	panic("not implemented")
}

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

func decodeTwoRegisters(instructionCode []byte, pc ProgramCounter) (rD uint8, rA uint8, err error) {
	if int(pc+1) >= len(instructionCode) {
		return 0, 0, fmt.Errorf("pc out of bound")
	}
	rD = getRegModIndex(instructionCode, pc)
	rA = getRegFloorIndex(instructionCode, pc)
	return rD, rA, nil
}

func decodeTwoRegistersAndOneImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (uint8, uint8, uint64) {
	rA := min(12, instructionCode[pc+1]&15)
	rB := min(12, instructionCode[pc+1]>>4)
	lX := min(4, max(0, skipLength-1))
	decodedVX, err := utils.DeserializeFixedLength(instructionCode[pc+2:pc+2+lX], types.U64(lX))
	if err != nil {
		return 0, 0, 0
	}
	vX, err := SignExtend(int(lX), uint64(decodedVX))
	return rA, rB, vX
}

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
	if mem.Pages[pageNum] != nil { // page allocated
		// try to allocate read-only memory --> page-fault
		if mem.Pages[pageNum].Access != MemoryReadWrite {
			return PVMExitTuple(PAGE_FAULT, memIndex)
		}

		if pageIndex+uint32(offset) < 4096 { // data allocated do not exceed maxSize
			tempPage := make([]byte, pageIndex+uint32(offset))
			copy(tempPage, mem.Pages[pageNum].Value)
			mem.Pages[pageNum].Value = tempPage

			for i := range offset {
				mem.Pages[pageNum].Value[i+int(pageIndex)] = vY[i]
			}
		} else { // data allocated exceed maxSize
			tempPage := make([]byte, ZP)
			copy(tempPage, mem.Pages[pageNum].Value)
			remainDataLength := ZP - int(pageIndex)
			for i := range remainDataLength {
				mem.Pages[pageNum].Value[i+int(pageIndex)] = vY[i]
			}
			vY = vY[remainDataLength:]
			Immediate = Immediate % (1<<remainDataLength + 1) // filter allocated
			remainDataLength = len(vY) - remainDataLength

			err := storeIntoMemory(mem, remainDataLength, memIndex+uint32(remainDataLength), Immediate)
			if err != PVMExitTuple(CONTINUE, nil) {
				return PVMExitTuple(PAGE_FAULT, memIndex)
			}
		}

	} else { // page not allocated, allocate the page
		mem.Pages[pageNum] = &Page{
			Access: MemoryReadWrite,
			Value:  make([]byte, len(vY)),
		}
	}
	return PVMExitTuple(CONTINUE, nil)
}

func loadFromMemory(mem Memory, offset uint32, vx uint32) (uint64, error) {
	vX := uint32(vx)

	pageNum := vX / ZP
	pageIndex := vX % ZP
	// load memory : the page must be exist and is not Inaccessible
	// Inaccessible should be appreciated ? memory : read-only, read-write, unallocated = Inaccessible

	if mem.Pages[pageNum] == nil || mem.Pages[pageNum].Access == MemoryInaccessible {
		return 0, PVMExitTuple(PAGE_FAULT, vx)
	}
	var memBytes []byte
	if pageIndex+offset < 4096 {
		memBytes = mem.Pages[pageNum].Value[pageIndex : pageIndex+offset]
	} else {
		if mem.Pages[pageNum+1] == nil || mem.Pages[pageNum+1].Access == MemoryInaccessible {
			return 0, PVMExitTuple(PAGE_FAULT, vx)
		}
		memBytes = mem.Pages[pageNum].Value[pageIndex:]
		remainBytes := mem.Pages[pageNum+1].Value[:offset-(ZP-pageIndex)]
		memBytes = append(memBytes, remainBytes...)
	}
	memVal, err := utils.DeserializeFixedLength(memBytes, types.U64(offset))
	if err != nil {
		log.Printf("loadMemory deserialize raise error memoryPage %d at index %d : %s ", pageNum, vx, err)
		return 0, err
	}
	return uint64(memVal), PVMExitTuple(CONTINUE, nil)
}
