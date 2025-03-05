package PolkaVM

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
)

// A.5.2
func decodeOneImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, error) {
	panic("not implemented")
}

// A.5.3
func decodeOneRegisterAndOneExtendedWidthImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, uint64, error) {
	panic("not implemented")
}

// A.5.4
func decodeTwoImmediates(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (uint64, uint64) {
	lX := min(4, int(instructionCode[1])%8)
	pcMargin := int(pc) + 2 + lX
	decodedVX, err := utils.DeserializeFixedLength(instructionCode[pc+2:pcMargin], types.U64(lX))
	if err != nil {
		fmt.Printf("opcode %s at instruction %d deserialize vx raise error : %s", zeta[opcode(instructionCode[pc])], pc, err)
	}

	vX, err := SignExtend(lX, uint64(decodedVX))
	// uint64 % uint32
	// vX := uint32(vXTemp)

	ly := min(4, max(0, int(skipLength)-lX-1))
	decodedVy, err := utils.DeserializeFixedLength(instructionCode[pcMargin:pcMargin+lX], types.U64(ly))
	if err != nil {
		fmt.Printf("opcode %s at instruction %d deserialize vy raise error : %s", zeta[opcode(instructionCode[pc])], pc, err)
	}
	vY, err := SignExtend(ly, uint64(decodedVy))
	if err != nil {
		return 0, 0
	}
	// vY := utils.WrapU64(types.U64(vyVal)).Serialize()

	return vX, vY
}

// A.5.5
func decodeOneOffset(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, error) {
	// TODO: deal with signed vs. unsigned and its impact on say 16 bit reads
	lX := min(4, skipLength)
	offsetData := instructionCode[pc+1 : pc+lX]
	offset, _, err := ReadUintFixed(offsetData, len(offsetData))
	if err != nil {
		return 0, err
	}

	return int(offset), nil
}

// A.5.6
func decodeOneRegisterAndOneImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, int, error) {
	rA := min(12, instructionCode[pc+1]%16)
	lX := min(4, max(0, skipLength-1))
	immediateData := instructionCode[pc+2 : pc+lX]
	immediate, _, err := ReadUintFixed(immediateData, len(immediateData))
	if err != nil {
		fmt.Printf("opcode %s at instruction %d deserialize vy raise error : %s", zeta[opcode(instructionCode[pc])], pc, err)
		return 0, 0, err
	}

	return int(rA), int(immediate), nil
}

// A.5.7
func decodeOneRegisterAndTwoImmediates(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int8, uint64, uint64, error) {
	rA := int8(min(12, instructionCode[pc+1]%16))
	lX := ProgramCounter(min(4, (instructionCode[pc+1]/16)%8))
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

// A.5.8
func decodeOneRegisterOneImmediateAndOneOffset(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, int, int, error) {
	panic("not implemented")
}

// A.5.9
func decodeTwoRegisters(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, int, error) {
	panic("not implemented")
}

// A.5.10
func decodeTwoRegistersAndOneImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, int, int, error) {
	panic("not implemented")
}

// A.5.11
func decodeTwoRegistersAndOneOffset(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, int, int, error) {
	panic("not implemented")
}

// A.5.12
func decodeTwoRegistersAndTwoImmediates(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, int, int, int, error) {
	panic("not implemented")
}

// A.5.13
func decodeThreeRegisters(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, int, int, error) {
	panic("not implemented")
}

func storeMem(mem Memory, vx uint64, vy uint64) error {
	vX := uint32(vx)
	pageNum := vX / ZP

	if mem.Pages[pageNum] != nil {
		if mem.Pages[pageNum].Access != MemoryReadWrite {
			return PVMExitTuple(PAGE_FAULT, vx)
		}
	} else {
		mem.Pages[pageNum] = &Page{
			Access: MemoryReadWrite,
			Value:  make([]byte, ZP),
		}
	}
	vY := utils.WrapU64(types.U64(vy)).Serialize()
	vx %= ZP
	for i := range len(vY) {
		mem.Pages[pageNum].Value[vx] = vY[i]
	}

	return PVMExitTuple(CONTINUE, nil)
}

func loadMemory(mem Memory, vx uint32) error {
	pageNum := vx / ZP
	// load memory : the page must be exist and is not Inaccessible
	if mem.Pages[pageNum] != nil || mem.Pages[pageNum].Access == MemoryInaccessible {
		return PVMExitTuple(PAGE_FAULT, vx)
	}
	return PVMExitTuple(CONTINUE, nil)
}
