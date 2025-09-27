package PVM

import (
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/logger"
)

var (
	instrCount = 0

	// hex, dec
	instrLogFormat = "dec"
)

/*
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
*/

// JIT version of (A.1) ψ_1
func SingleInvoke(program Program, pc ProgramCounter, gas Gas, reg Registers, mem Memory) (error, ProgramCounter, Gas, Registers, Memory) {
	gasPrime := Gas(gas)

	// decode instructions in a block
	pcPrime, instrCount, err := DecodeInstructionBlock(program.InstructionData, pc, program.Bitmasks)
	if err != nil {
		logger.Errorf("DecodeInstructionBlock error : %v", err)
		return err, 0, Gas(gas), reg, mem
	}

	// check gas, currently each instruction gas = 1, so only check instrCount
	if Gas(gas) < Gas(instrCount) {
		logger.Debugf("service out-of-gas: required %d, but only %d", instrCount, gas)
		return PVMExitTuple(OUT_OF_GAS, nil), pc, Gas(gas), reg, mem
	}

	// To avoid duplicate charging
	// ecalli will back to host-call level, then continue to execute remaining instructions
	if program.Bitmasks.IsStartOfBasicBlock(pc) {
		// charge gas
		gasPrime -= Gas(instrCount)
		logger.Debugf("    charge gas : %d -> %d", gas, gasPrime)
	}

	// execute instructions in the block
	pc, regPrime, memPrime, exitReason := ExecuteInstructions(program.InstructionData, program.Bitmasks, program.JumpTable, pc, pcPrime, reg, mem)
	var pvmExit *PVMExitReason
	if !errors.As(exitReason, &pvmExit) {
		return exitReason, 0, 0, Registers{}, Memory{}
	}

	reason := exitReason.(*PVMExitReason).Reason
	switch reason {
	case PANIC, HALT:
		return exitReason, 0, Gas(gasPrime), regPrime, memPrime
	case HOST_CALL:
		return exitReason, pc, Gas(gasPrime), regPrime, memPrime
	}

	// reason == CONTINUE
	return SingleInvoke(program, pc, gasPrime, regPrime, memPrime)
}

func DecodeInstructionBlock(instructionData ProgramCode, pc ProgramCounter, bitmask Bitmask) (ProgramCounter, int64, error) {
	pcPrime := pc
	count := int64(1)

	for {

		// check pc is not out of range and avoid infinit-loop
		if pc > ProgramCounter(len(instructionData)) {
			logger.Debugf("PVM panic: program counter out of range, pcPrime = %d > program-length = %d", pcPrime, len(instructionData))
			return pc, 0, PVMExitTuple(PANIC, nil)
		}

		// check opcode is valid after computing with skip
		if !instructionData.isOpcodeValid(pc) {
			logger.Debugf("PVM panic: decode program failed: opcode invalid")
			return pc, 0, PVMExitTuple(PANIC, nil)
		}

		// reach instruction block end
		if isBasicBlockTerminationInstruction(instructionData[pcPrime]) {
			return pcPrime, count, nil
		}
		count++
		skipLength := skip(int(pcPrime), bitmask)
		pcPrime += ProgramCounter(skipLength) + 1
	}
}

// execute each instruction in block[pc:pcPrime] , pcPrime is computed by DecodeInstructionBlock
func ExecuteInstructions(instructionData ProgramCode, bitmask Bitmask, jumpTable JumpTable, pc ProgramCounter, pcPrime ProgramCounter, registers Registers, memory Memory) (ProgramCounter, Registers, Memory, error) {
	// no need to worry about gas, opcode valid here, it's checked in HostCall and DecodeInstructionBlock respectively
	for pc <= pcPrime {
		opcodeData := instructionData[pc]
		skipLength := ProgramCounter(skip(int(pc), bitmask))

		exitReason, newPC, registersPrime, memoryPrime := execInstructions[opcodeData](instructionData, pc, skipLength, registers, memory, jumpTable, bitmask)

		registers = registersPrime
		memory = memoryPrime
		instrCount++

		var pvmExit *PVMExitReason
		if !errors.As(exitReason, &pvmExit) && exitReason != nil {
			return pc, registers, memory, exitReason
		}

		reason := exitReason.(*PVMExitReason).Reason
		switch reason {
		case PANIC, HALT:
			return 0, registers, memory, exitReason
		case HOST_CALL:
			return pc + skipLength + 1, registers, memory, exitReason
		}

		if pc != newPC {
			// check branch
			return newPC, registers, memory, exitReason
		}

		pc += skipLength + 1
	}

	return pc, registers, memory, PVMExitTuple(CONTINUE, nil)
}

type Int interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 | ~int8 | ~int16 | ~int32 | ~int64 | ~int | ~uint
}

func formatInt[T Int](num T) string {
	if instrLogFormat == "hex" {
		return fmt.Sprintf("0x%x", num)
	}
	return fmt.Sprintf("%d", num)
}
