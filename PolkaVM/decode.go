package PolkaVM

import (
	"fmt"
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

func decodeThreeRegisters(instructionCode []byte, pc ProgramCounter) (rA uint8, rB uint8, rD uint8, err error) {
	if int(pc+2) >= len(instructionCode) {
		return 0, 0, 0, fmt.Errorf("pc out of bound")
	}
	rA = getRegModIndex(instructionCode, pc)
	rB = getRegFloorIndex(instructionCode, pc)
	rD = min(12, instructionCode[pc+2])
	return rA, rB, rD, nil
}
