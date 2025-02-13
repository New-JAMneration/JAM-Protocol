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

// skip computes the distance to the next opcode  A.3
func skip(i int, bitmask []byte) uint32 {
	j := 1
	for ; j < len(bitmask); j++ {
		if bitmask[j] == byte(1) {
			break
		}
	}
	return uint32(min(24, i+1+j))
}

func inBasicBlock(data []byte, bitmask []byte, n int) bool {
	if bitmask[n] != byte(1) {
		return false
	}
	// TODO data[n] is in defined opcodes, need to wait for opcodes defined
	// data[n]

	return true
}
