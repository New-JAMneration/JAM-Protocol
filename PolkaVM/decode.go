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
func decodeTwoImmediates(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (uint64, uint64) {
	lX := ProgramCounter(min(4, uint8(instructionCode[pc+1])))

	decodedVX, err := utils.DeserializeFixedLength(instructionCode[pc+2:pc+2+lX], types.U64(lX))
	if err != nil {
		log.Printf("opcode %s(%d) at instruction %d deserialize vx raise error : %s", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
		return 0, 0
	}

	vX, err := SignExtend(int(lX), uint64(decodedVX))
	if err != nil {
		log.Printf("opcode %s(%d) at instruction %d signExtend raise error : %s", zeta[opcode(instructionCode[pc])], opcode(instructionCode[pc]), pc, err)
		return 0, 0
	}
	lY := min(4, max(0, skipLength-lX-1))
	decodedVy, err := utils.DeserializeFixedLength(instructionCode[pc+2+lX:pc+2+lX+lY], types.U64(lY))
	if err != nil {
		log.Printf("opcode %s at instruction %d deserialize vy raise error : %s", zeta[opcode(instructionCode[pc])], pc, err)
	}
	vY, err := SignExtend(int(lY), uint64(decodedVy))
	fmt.Println("signExtented vY : ", vY)
	if err != nil {
		return 0, 0
	}

	return vX, vY
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
		log.Printf("opcode %s at instruction %d deserialize vy raise error : %s", zeta[opcode(instructionCode[pc])], pc, err)
		return 0, 0, err
	}

	return rA, immediate, nil
}

// A.5.7
func decodeOneRegisterAndTwoImmediates(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int8, uint64, uint64, error) {
	rA := int8(min(12, instructionCode[pc+1]%16))
	lX := ProgramCounter(min(4, uint8((instructionCode[pc+1] >> 4))))
	pcMargin := pc + 2 + lX
	decodedVX, err := utils.DeserializeFixedLength(instructionCode[pc+2:pcMargin], types.U64(lX))
	if err != nil {
		fmt.Printf("opcode %s at instruction %d deserialize vy raise error : %s", zeta[opcode(instructionCode[pc])], pc, err)
		return 0, 0, 0, err
	}
	vX, err := SignExtend(int(lX), uint64(decodedVX))

	lY := min(4, max(0, skipLength-lX-1))
	decodedVY, err := utils.DeserializeFixedLength(instructionCode[pcMargin:pcMargin+lY], types.U64(lY))
	vY, err := SignExtend(int(lY), uint64(decodedVY))

	return rA, vX, vY, err
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

// A.5.9
func decodeTwoRegisters(instructionCode []byte, pc ProgramCounter) (rD uint8, rA uint8, err error) {
	if int(pc+1) >= len(instructionCode) {
		return 0, 0, fmt.Errorf("pc out of bound")
	}
	rD = getRegModIndex(instructionCode, pc)
	rA = getRegFloorIndex(instructionCode, pc)
	return rD, rA, nil
}

// A.5.10
func decodeTwoRegistersAndOneImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, int, int, error) {
	panic("not implemented")
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

	vY := utils.SerializeFixedLength(types.U64(Immediate), types.U64(offset))
	if mem.Pages[pageNum] != nil { // page allocated
		// try to allocated read-only memory --> page-fault
		if mem.Pages[pageNum].Access != MemoryReadWrite {
			return PVMExitTuple(PAGE_FAULT, memIndex)
		}
		originLength := len(mem.Pages[pageNum].Value)
		fmt.Println("originLength : ", originLength)
		if originLength+offset < 4096 { // data allocated do not exceed maxSize
			tempPage := make([]byte, originLength+offset)
			copy(tempPage, mem.Pages[pageNum].Value)
			mem.Pages[pageNum].Value = tempPage
			for i := range offset {
				mem.Pages[pageNum].Value[i+originLength] = vY[i]
			}
		} else { // data allocated exceed maxSize
			tempPage := make([]byte, ZP)
			copy(tempPage, mem.Pages[pageNum].Value)
			remainDataLength := ZP - originLength
			for i := range remainDataLength {
				mem.Pages[pageNum].Value[i+originLength] = vY[i]
			}
			vY = vY[remainDataLength:]
			Immediate = Immediate % (1<<remainDataLength + 1) // filter allocated
			remainDataLength = len(vY) - remainDataLength

			err := storeIntoMemory(mem, remainDataLength, memIndex+ZP, Immediate)
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

	memIndex %= ZP

	for i := range len(vY) {
		mem.Pages[pageNum].Value[memIndex] = vY[i]
	}
	return PVMExitTuple(CONTINUE, nil)
}

func loadFromMemory(mem Memory, offset uint32, vx uint32) (uint64, error) {
	vX := uint32(vx)

	pageNum := vX / ZP
	// load memory : the page must be exist and is not Inaccessible
	if mem.Pages[pageNum] != nil || mem.Pages[pageNum].Access == MemoryInaccessible {
		return 0, PVMExitTuple(PAGE_FAULT, vx)
	}

	memBytes := mem.Pages[pageNum].Value[vx : vx+offset]
	memVal, err := utils.DeserializeFixedLength(memBytes, types.U64(offset))
	if err != nil {
		log.Printf("loadMemory deserialize raise error memoryPage %d at index %d : %s ", pageNum, vx, err)
		return 0, err
	}

	return uint64(memVal), PVMExitTuple(CONTINUE, nil)
}
