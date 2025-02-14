package PolkaVM

// SingleStepInvoke is A.1 (v0.6.2)
func SingleStepInvoke(programBlob []byte, programCounter uint32,
	gas Gas, registers Registers, memory Memory) (
	ExitReasonTypes, uint32, Gas, Registers, Memory,
) {
	// deblob programCodeBlob (c, k, j)  A.2
	programCodeBlob, err := DeBlobProgramCode(programBlob)
	if err == PANIC {
		return err, programCounter, gas, registers, memory
	}

	var exitReason ExitReasonTypes
	exitReason, programCounter, gas, registers, memory = SingleStepStateTransition(
		programBlob, programCodeBlob.Bitmasks,
		programCodeBlob.JumpTables, programCounter, gas,
		registers, memory)

	if gas < 0 {
		return OUT_OF_GAS, programCounter, gas, registers, memory
	}

	switch exitReason {
	case CONTINUE:
		SingleStepInvoke(programBlob, programCounter, gas, registers, memory)
	case HALT:
	case PANIC:
		return exitReason, 0, gas, registers, memory
	default:
		return exitReason, programCounter, gas, registers, memory
	}

	return exitReason, programCounter, gas, registers, memory
}

// SingleStepStateTransition is A.6, A.7
func SingleStepStateTransition(instructionCode []byte, bitmask []byte, jumpTable []uint64,
	programCounter uint32, gas Gas, registers Registers, memory Memory) (
	ExitReasonTypes, uint32, Gas, Registers, Memory,
) {
	// TODO : execute instructions and output exit-reason, registers, memory
	var exitReason ExitReasonTypes
	programCounter += uint32(1) + skip(int(programCounter), bitmask)

	return exitReason, programCounter, gas, registers, memory
}
