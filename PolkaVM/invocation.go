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

	if exitReason == ErrNotImplemented {
		return exitReason, programCounterPrime, gasPrime, registersPrime, memoryPrime
	}

	switch exitReason.(*PVMExitReason).Reason {
	case CONTINUE:
		if gasPrime < 0 {
			return PVMExitTuple(OUT_OF_GAS, nil), programCounter, gas, registers, memory
		}

		return SingleStepInvoke(programBlob, programCounterPrime, gasPrime, registersPrime, memoryPrime)
	case HALT:
		fallthrough
	case PANIC:
		return exitReason, programCounterPrime, gasPrime, registersPrime, memoryPrime
	default:
		return exitReason, programCounterPrime, gasPrime, registersPrime, memoryPrime
	}
}

var ErrNotImplemented = fmt.Errorf("instruction not implemented")

// (v.6.2 A.6, A.7) SingleStepStateTransition
func SingleStepStateTransition(instructionCode []byte, bitmask Bitmask, jumpTable JumpTable,
	programCounter ProgramCounter, gas Gas, registers Registers, memory Memory) (
	error, ProgramCounter, Gas, Registers, Memory,
) {
	if int(programCounter) >= len(instructionCode) {
		return PVMExitTuple(PANIC, nil), programCounter, gas, registers, memory
	}

	var exitReason error
	gasDelta := Gas(0)
	// (v.6.2 A.19) l = skip(iota)
	skipLength := ProgramCounter(skip(int(programCounter), bitmask))

	opcode := instructionCode[programCounter]

	target := execInstructions[opcode]
	if target == nil {
		return ErrNotImplemented, programCounter, gas, registers, memory
	}

	exitReason, newProgramCounter, gasDelta, registers, memory := execInstructions[opcode](instructionCode, programCounter, skipLength, registers, memory, jumpTable, bitmask)

	// recently, set all gasDelta = 2 for consistent with testvector
	gas -= gasDelta

	if exitReason.(*PVMExitReason).Reason == CONTINUE && newProgramCounter == programCounter {
		newProgramCounter += skipLength + 1
	}

	return exitReason, newProgramCounter, gas, registers, memory
}
