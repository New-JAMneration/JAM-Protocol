package PolkaVM

// SingleStepInvoke is A.1 (v0.6.2)
func SingleStepInvoke(programBlob []byte, programCounter uint32,
	gas Gas, registers Registers, memory Memory) (
	error, uint32, Gas, Registers, Memory,
) {
	// deblob programCodeBlob (c, k, j)  A.2
	programCodeBlob, err := DeBlobProgramCode(programBlob)
	if err == PANIC {
		return PVMExitTuple(PANIC, nil), programCounter, gas, registers, memory
	}

	exitReason, programCounterPrime, gasPrime, registersPrime, memoryPrime := SingleStepStateTransition(
		programBlob, programCodeBlob.Bitmasks,
		programCodeBlob.JumpTables, programCounter, gas,
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

// SingleStepStateTransition is A.6, A.7
func SingleStepStateTransition(instructionCode []byte, bitmask []byte, jumpTable []uint64,
	programCounter uint32, gas Gas, registers Registers, memory Memory) (
	error, uint32, Gas, Registers, Memory,
) {
	// TODO : execute instructions and return PVM state with prime
	var err error
	programCounter += uint32(1) + skip(int(programCounter), bitmask)

	return err, programCounter, gas, registers, memory
}
