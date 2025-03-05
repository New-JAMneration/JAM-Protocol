package PolkaVM

import "fmt"

// SingleStepInvoke is A.1 (v0.6.2)
func SingleStepInvoke(programBlob []byte, programCounter ProgramCounter,
	gas Gas, registers Registers, memory Memory) (
	error, ProgramCounter, Gas, Registers, Memory,
) {
	// deblob programCodeBlob (c, k, j)  A.2
	programCodeBlob, err := DeBlobProgramCode(programBlob)
	if err == PVMExitTuple(PANIC, nil) {
		return err, programCounter, gas, registers, memory
	}

	var exitReason error
	exitReason, programCounterPrime, gasPrime, registersPrime, memoryPrime := SingleStepStateTransition(
		programCodeBlob.InstructionData, programCodeBlob.Bitmasks,
		programCodeBlob.JumpTable, programCounter, gas,
		registers, memory)

	if gas < 0 {
		return PVMExitTuple(OUT_OF_GAS, nil), programCounter, gas, registers, memory
	}

	switch exitReason {
	case PVMExitTuple(CONTINUE, nil):
		SingleStepInvoke(programBlob, programCounterPrime, gasPrime, registersPrime, memoryPrime)
	case PVMExitTuple(HALT, nil):
	case PVMExitTuple(PANIC, nil):
		return exitReason, 0, gasPrime, registersPrime, memoryPrime
	default:
		return exitReason, programCounterPrime, gasPrime, registersPrime, memoryPrime
	}

	return exitReason, programCounterPrime, gasPrime, registersPrime, memoryPrime
}

// (v.6.2 A.6, A.7) SingleStepStateTransition
func SingleStepStateTransition(instructionCode []byte, bitmask []bool, jumpTable JumpTable,
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
	exitReason, programCounter, gasDelta, registers, memory = execInstructions[opcode](instructionCode, programCounter, skipLength, registers, memory, jumpTable, bitmask)
	// recently, set all gasDelta = 2 for consistent with testvector
	gas -= gasDelta
	// TODO : execute instructions and output exit-reason, registers, memory
	programCounter += skipLength + 1

	fmt.Println("run opcode", opcode)
	return exitReason, programCounter, gas, registers, memory
}
