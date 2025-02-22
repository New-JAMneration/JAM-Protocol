package PolkaVM

import "fmt"

// SingleStepInvoke is A.1 (v0.6.2)
func SingleStepInvoke(programBlob []byte, programCounter ProgramCounter,
	gas Gas, registers Registers, memory Memory) (
	error, ProgramCounter, Gas, Registers, Memory,
) {
	// deblob programCodeBlob (c, k, j)  A.2
	programCodeBlob, err := DeBlobProgramCode(programBlob)
	if err == PANIC {
		return PVMExitTuple(PANIC, nil), programCounter, gas, registers, memory
	}

	var exitReason error
	exitReason, programCounter, gas, registers, memory = SingleStepStateTransition(
		programCodeBlob.InstructionData, programCodeBlob.Bitmasks,
		programCodeBlob.JumpTables, programCounter, gas,
		registers, memory)

	if gas < 0 {
		return PVMExitTuple(OUT_OF_GAS, nil), programCounter, gas, registers, memory
	}

	switch exitReason {
	case PVMExitTuple(CONTINUE, nil):
		SingleStepInvoke(programBlob, programCounter, gas, registers, memory)
	case PVMExitTuple(HALT, nil):
	case PVMExitTuple(PANIC, nil):
		return PVMExitTuple(PANIC, nil), 0, gas, registers, memory
	default:
		return exitReason, programCounter, gas, registers, memory
	}

	return exitReason, programCounter, gas, registers, memory
}

// (v.6.2 A.6, A.7) SingleStepStateTransition
func SingleStepStateTransition(instructionCode []byte, bitmask []byte, jumpTable []uint64,
	programCounter ProgramCounter, gas Gas, registers Registers, memory Memory) (
	error, ProgramCounter, Gas, Registers, Memory,
) {
	var exitReason error
	gasDelta := Gas(0)
	// (v.6.2 A.4) append zero to trap
	instructionCode = append(instructionCode, 0, 0)
	// (v.6.2 A.19) l = skip(iota)
	skipLength := ProgramCounter(skip(int(programCounter), bitmask))
	opcode := instructionCode[programCounter]
	exitReason, programCounter, gasDelta, registers, memory = execInstructions[opcode](instructionCode, programCounter, skipLength, registers, memory)
	// recently, set all gasDelta = 2 for consistent with testvector
	gas -= gasDelta
	// TODO : execute instructions and output exit-reason, registers, memory
	programCounter += skipLength

	fmt.Println("run opcode", opcode)
	return exitReason, programCounter, gas, registers, memory
}
