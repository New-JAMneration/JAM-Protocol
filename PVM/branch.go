package PVM

func branch(pc ProgramCounter, b ProgramCounter, C bool, bitmask Bitmask, instruction ProgramCode) (ExitReason, ProgramCounter) {
	switch {
	case !C:
		return ExitContinue, pc
	case !bitmask.IsStartOfBasicBlock(b) && instruction.isOpcodeValid(b):
		return ExitPanic, pc
	default:
		return ExitContinue, b
	}
}

func djump(pc ProgramCounter, a uint32, jumpTable JumpTable, bitmask Bitmask) (ExitReason, ProgramCounter) {
	switch {
	case a == 0xffff0000:
		return ExitHalt, pc
		// jumpTable Size : |j|  , jumpTable Length E_1(z)
	case a == 0 || a > jumpTable.Size*ZA || a%ZA != 0:
		// case a == 0 || a > jumpTable.Size*jumpTable.Length || a%jumpTable.Length != 0:
		return ExitPanic, pc
	}
	index := a/ZA - 1 // GP,  if  ZA > 1, index = ZA*index
	dest, _, err := ReadUintFixed(jumpTable.Data[index*jumpTable.Length:], int(jumpTable.Length))
	if err != nil {
		// memory corruption?
		panic(err.Error())
	}

	newPC := ProgramCounter(dest)

	if !bitmask.IsStartOfBasicBlock(newPC) {
		return ExitPanic, pc
	}

	return ExitContinue, newPC
}
