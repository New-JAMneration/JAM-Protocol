package PVM

// sjump (A.20): unconditional static jump.
// Panics if target b is not a basic block start (b ∉ ϖ).
func sjump(pc ProgramCounter, b ProgramCounter, bitmask Bitmask) (ExitReason, ProgramCounter) {
	if !bitmask.IsStartOfBasicBlock(b) {
		return ExitPanic, pc
	}
	return ExitContinue, b
}

// branch (A.21): conditional jump with dual-target validation.
// Both b and ft must be basic block starts; panics otherwise.
func branch(pc ProgramCounter, b ProgramCounter, C bool, ft ProgramCounter, bitmask Bitmask) (ExitReason, ProgramCounter) {
	if !bitmask.IsStartOfBasicBlock(b) || !bitmask.IsStartOfBasicBlock(ft) {
		return ExitPanic, pc
	}
	if !C {
		return ExitContinue, ft
	}
	return ExitContinue, b
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

// DjumpResolve (A.22): dynamic jump resolution and validation.
// Exported for the JIT recompiler; interpreter uses the lowercase djump alias.
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
