package PolkaVM

func branch(pc ProgramCounter, b ProgramCounter, C bool, bitmask Bitmask) (ExitReasonTypes, ProgramCounter) {
	switch {
	case !C:
		return CONTINUE, pc
	case !bitmask.IsStartOfBasicBlock(b):
		return PANIC, pc
	default:
		return CONTINUE, b
	}
}

func djump(pc ProgramCounter, a uint32, jumpTable JumpTable, bitmask Bitmask) (ExitReasonTypes, ProgramCounter) {
	switch {
	case a == 0xffff0000:
		return HALT, pc
	case a == 0 || a > jumpTable.Length*ZA || a%jumpTable.Length != 0:
		return PANIC, pc
	}

	index := a/ZA - 1

	dest, _, err := ReadUintFixed(jumpTable.Data[index:], int(jumpTable.Size))
	if err != nil {
		// memory corruption?
		panic(err.Error())
	}

	newPC := ProgramCounter(dest)

	if !bitmask.IsStartOfBasicBlock(newPC) {
		return PANIC, pc
	}

	return CONTINUE, newPC
}
