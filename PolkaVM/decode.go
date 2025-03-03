package PolkaVM

func getRegModIndex(instructionCode []byte, pc ProgramCounter) int {
	return min(12, (int(instructionCode[pc+1]) % 16))
}
func getRegFloorIndex(instructionCode []byte, pc ProgramCounter) int {
	return min(12, (int(instructionCode[pc+1]) >> 4))
}

func decodeOneImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, error) {
	panic("not implemented")
}

func decodeOneRegisterAndOneExtendedWidthImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, uint64, error) {
	panic("not implemented")
}

func decodeTwoImmediates(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, int, error) {
	panic("not implemented")
}

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

func decodeOneRegisterAndOneImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, int, error) {
	rA := min(12, instructionCode[pc+1]%16)
	lX := min(4, max(0, skipLength-1))
	immediateData := instructionCode[pc+2 : pc+lX]
	immediate, _, err := ReadUintFixed(immediateData, len(immediateData))
	if err != nil {
		return 0, 0, err
	}

	return int(rA), int(immediate), nil
}

func decodeOneRegisterAndTwoImmediates(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, int, int, error) {
	panic("not implemented")
}

func decodeOneRegisterOneImmediateAndOneOffset(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, int, int, error) {
	panic("not implemented")
}

func decodeTwoRegisters(instructionCode []byte, pc ProgramCounter) (rD int, rA int) {
	rD = getRegModIndex(instructionCode, pc)
	rA = getRegFloorIndex(instructionCode, pc)

	return rD, rA
}

func decodeTwoRegistersAndOneImmediate(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, int, int, error) {
	panic("not implemented")
}

func decodeTwoRegistersAndOneOffset(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, int, int, error) {
	panic("not implemented")
}

func decodeTwoRegistersAndTwoImmediates(instructionCode []byte, pc ProgramCounter, skipLength ProgramCounter) (int, int, int, int, error) {
	panic("not implemented")
}

func decodeThreeRegisters(instructionCode []byte, pc ProgramCounter) (rA int, rB int, rD int) {
	rA = getRegModIndex(instructionCode, pc)
	rB = getRegFloorIndex(instructionCode, pc)
	rD = min(12, int(instructionCode[pc+2]))

	return rA, rB, rD
}
