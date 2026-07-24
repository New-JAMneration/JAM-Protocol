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

// ResolveDynamicJump resolves a jump-table address to a program PC.
// It does not require the destination to be a basic-block start; the interpreter
// enforces that separately in djump.
func ResolveDynamicJump(a uint32, jumpTable JumpTable) (ExitReason, ProgramCounter) {
	switch {
	case a == 0xffff0000:
		return ExitHalt, 0
	case a == 0 || a > jumpTable.Size*ZA || a%ZA != 0:
		return ExitPanic, 0
	}
	index := a/ZA - 1
	dest, _, err := ReadUintFixed(jumpTable.Data[index*jumpTable.Length:], int(jumpTable.Length))
	if err != nil {
		panic(err.Error())
	}
	return ExitContinue, ProgramCounter(dest)
}

func djump(pc ProgramCounter, a uint32, jumpTable JumpTable, bitmask Bitmask) (ExitReason, ProgramCounter) {
	return DjumpResolve(pc, a, jumpTable, bitmask)
}

// DjumpResolve performs the full graypaper §4.4.4 dynamic-jump resolution and validation:
// HALT for the sentinel address, panic on misaligned / out-of-range / non-basic-block targets,
// otherwise returns the resolved program counter.
// The pc parameter is the PC reported on panic (typically the jump_ind instruction PC).
// Exported for the JIT recompiler to call from its dispatcher; the interpreter
// continues to use the lowercase djump alias above.
func DjumpResolve(pc ProgramCounter, a uint32, jumpTable JumpTable, bitmask Bitmask) (ExitReason, ProgramCounter) {
	if a == 0xffff0000 {
		return ExitHalt, pc
	}
	reason, newPC := ResolveDynamicJump(a, jumpTable)
	if reason != ExitContinue {
		if reason.GetReasonType() == PANIC {
			return ExitPanic, pc
		}
		return reason, newPC
	}
	if !bitmask.IsStartOfBasicBlock(newPC) {
		return ExitPanic, pc
	}
	return ExitContinue, newPC
}
