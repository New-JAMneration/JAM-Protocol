package PVM

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/logger"
)

// SingleStepInvoke is A.1 (v0.6.2)
func SingleStepInvoke(programBlob []byte, programCounter ProgramCounter,
	gas Gas, registers Registers, memory Memory, gasConsumed Gas) (
	error, ProgramCounter, Gas, Registers, Memory,
) {
	// deblob programCodeBlob (c, k, j)  A.2
	programCodeBlob, err := DeBlobProgramCode(programBlob)
	if err == PVMExitTuple(PANIC, nil) {
		return err, programCounter, gas, registers, memory
	}
	var singleGasConsumed Gas
	gasPrime := gas

	var exitReason error
	exitReason, programCounterPrime, singleGasConsumed, registersPrime, memoryPrime := SingleStepStateTransition(
		programCodeBlob.InstructionData, programCodeBlob.Bitmasks,
		programCodeBlob.JumpTable, programCounter, gas,
		registers, memory)

	// accumulate how many gas will be consumed
	gasConsumed += singleGasConsumed

	if exitReason == ErrNotImplemented {
		return exitReason, programCounterPrime, gas - gasConsumed, registersPrime, memoryPrime
	}

	if isBasicBlockTerminationInstruction(programCodeBlob.InstructionData[programCounter]) {
		gasPrime -= gasConsumed
		gasConsumed = 0
		logger.Debugf("       gas : %d -> %d", gas, gasPrime)
	}

	switch exitReason.(*PVMExitReason).Reason {
	case CONTINUE:
		if gas < 0 {
			return PVMExitTuple(OUT_OF_GAS, nil), programCounter, gasPrime, registers, memory
		}

		return SingleStepInvoke(programBlob, programCounterPrime, gasPrime, registersPrime, memoryPrime, gasConsumed)
	case HALT:
		logger.Debugf("instr: fallthrough")
		fallthrough
	case PANIC:
		return exitReason, programCounterPrime, gasPrime - gasConsumed, registersPrime, memoryPrime
	default: // host-call
		return exitReason, programCounterPrime, gasPrime, registersPrime, memoryPrime
	}
}

var ErrNotImplemented = fmt.Errorf("instruction not implemented")

// (v.6.2 A.6, A.7) SingleStepStateTransition
func SingleStepStateTransition(instructionCode ProgramCode, bitmask Bitmask, jumpTable JumpTable,
	programCounter ProgramCounter, gas Gas, registers Registers, memory Memory) (
	error, ProgramCounter, Gas, Registers, Memory,
) {
	if int(programCounter) >= len(instructionCode) {
		return PVMExitTuple(PANIC, nil), programCounter, gas, registers, memory
	}
	var exitReason error
	// (v.6.2 A.19) l = skip(iota)
	skipLength := ProgramCounter(skip(int(programCounter), bitmask))

	// opcodeData := instructionCode[programCounter.isOpocode()]
	opcodeData := instructionCode.isOpcode(programCounter)
	target := execInstructions[opcodeData]
	if target == nil {
		return ErrNotImplemented, programCounter, gas, registers, memory
	}

	exitReason, newProgramCounter, gasDelta, registers, memory := execInstructions[opcodeData](instructionCode, programCounter, skipLength, registers, memory, jumpTable, bitmask)

	// detailed instruction print
	// log.Printf("instr:%s(%d) pc=%d operand=%v gas=%d registers=%x", zeta[opcode(opcodeData)], opcodeData, programCounter, instructionCode[programCounter:programCounter+skipLength+1], gas, registers)
	// log.Printf("       gas : %d -> %d", gas+gasDelta, gas)

	reason := exitReason.(*PVMExitReason).Reason
	if (reason == CONTINUE || reason == HOST_CALL) && newProgramCounter == programCounter {
		newProgramCounter += skipLength + 1
	}

	return exitReason, newProgramCounter, gasDelta, registers, memory
}
