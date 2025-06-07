package PVM

func branch(pc ProgramCounter, b ProgramCounter, C bool, bitmask Bitmask) (ExitReasonTypes, ProgramCounter) {
	switch {
	case !C:
		return CONTINUE, pc
	case !bitmask.IsStartOfBasicBlock(b) && b.isOpcodeValid():
		return PANIC, pc
	default:
		return CONTINUE, b
	}
}

func djump(pc ProgramCounter, a uint32, jumpTable JumpTable, bitmask Bitmask) (ExitReasonTypes, ProgramCounter) {
	switch {
	case a == 0xffff0000:
		return HALT, pc
		// jumpTable Size : |j|  , jumpTable Length E_1(z)
	case a == 0 || a > jumpTable.Size*ZA || a%ZA != 0:
		// case a == 0 || a > jumpTable.Size*jumpTable.Length || a%jumpTable.Length != 0:
		return PANIC, pc
	}
	index := a/ZA - 1 // GP,  if  ZA > 1, index = ZA*index
	dest, _, err := ReadUintFixed(jumpTable.Data[index*jumpTable.Length:], int(jumpTable.Length))
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
