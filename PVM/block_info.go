package PVM

import "encoding/binary"

// InstrMeta holds pre-decoded metadata for a single PVM instruction.
// Populated once at deblob time; never mutated afterwards.
type InstrMeta struct {
	PC      ProgramCounter // 4B
	Opcode  byte           // 1B
	SkipLen uint8          // 1B, max 24
	Dst     uint8          // 1B, destination reg index (0xFF = none)
	Src     [2]uint8       // 2B, source reg indices (0xFF = unused)
	Exec    instrMetaFn    // pre-resolved handler; set at deblob time
	Imm     [2]uint64      // 16B, immediates / branch target PC
}

// BlockMeta holds pre-decoded metadata for a single PVM basic block.
// Populated once at deblob time; never mutated afterwards.
type BlockMeta struct {
	StartPC    ProgramCounter
	EndPC      ProgramCounter // PC of the terminating instruction (inclusive)
	InstrStart int            // index into Program.Instrs[]
	InstrEnd   int            // exclusive upper bound into Program.Instrs[]
	GasCost    Gas            // v0.7.2: = InstrCount; TODO(gas-model): = simulatePipeline()
}

// InstrCount returns the number of instructions in this block.
func (b *BlockMeta) InstrCount() int {
	return b.InstrEnd - b.InstrStart
}

// decodeOperands populates the Dst, Src, Imm fields of an InstrMeta
// by dispatching on the instruction's category and calling the existing
// decode functions from decode.go. Called once per instruction at deblob time.
func decodeOperands(instr *InstrMeta, idata ProgramCode, bitmask Bitmask) {
	pc := instr.PC
	skipLen := ProgramCounter(instr.SkipLen)
	instr.Dst = 0xFF
	instr.Src = [2]uint8{0xFF, 0xFF}

	instrCategory := opcodeInfoTable[instr.Opcode].Category
	if instrCategory != InstrCatNoArg && int(pc)+1 >= len(idata) {
		return
	}

	switch instrCategory {
	case InstrCatNoArg:
		// 0, 1: no operands

	case InstrCatOneImm:
		// 10 (ecalli): Imm[0] = callID
		if skipLen < 1 {
			return
		}
		if callID, err := decodeOneImmediate(idata, pc, skipLen); err == nil {
			instr.Imm[0] = uint64(callID)
		}

	case InstrCatOneRegExtImm:
		// 20 (load_imm_64): Dst = rA, Imm[0] = imm64
		instr.Dst = min(12, idata[pc+1]%16)
		if int(pc+10) <= len(idata) {
			instr.Imm[0] = binary.LittleEndian.Uint64(idata[pc+2 : pc+10])
		}

	case InstrCatTwoImm:
		// 30-33: Imm[0] = addr (vX), Imm[1] = val (vY)
		if vx, vy, err := decodeTwoImmediates(idata, pc, skipLen); err == nil {
			instr.Imm[0] = vx
			instr.Imm[1] = vy
		}

	case InstrCatOneOffset:
		// 40 (jump): Imm[0] = target PC
		if target, err := decodeOneOffset(idata, pc, skipLen); err == nil {
			instr.Imm[0] = uint64(target)
		}

	case InstrCatOneRegOneImm:
		// 50-62: Dst = rA, Src[0] = rA, Imm[0] = vX
		if rA, vX, err := decodeOneRegisterAndOneImmediate(idata, pc, skipLen); err == nil {
			instr.Dst = rA
			instr.Src[0] = rA
			instr.Imm[0] = vX
		}

	case InstrCatOneRegTwoImm:
		// 70-73: Src[0] = rA, Imm[0] = vX, Imm[1] = vY
		if rA, vX, vY, err := decodeOneRegisterAndTwoImmediates(idata, pc, skipLen); err == nil {
			instr.Dst = uint8(rA)
			instr.Src[0] = uint8(rA)
			instr.Imm[0] = vX
			instr.Imm[1] = vY
		}

	case InstrCatOneRegImmOff:
		// 80-90: Dst = rA, Src[0] = rA, Imm[0] = vX, Imm[1] = target PC
		if rA, vX, target, err := decodeOneRegisterOneImmediateAndOneOffset(idata, pc, skipLen); err == nil {
			instr.Dst = rA
			instr.Src[0] = rA
			instr.Imm[0] = vX
			instr.Imm[1] = uint64(target)
		}

	case InstrCatTwoReg:
		// 100-111: Dst = rD, Src[0] = rA
		if rD, rA, err := decodeTwoRegisters(idata, pc); err == nil {
			instr.Dst = rD
			instr.Src[0] = rA
		}

	case InstrCatTwoRegOneImm:
		// 120-161: Dst = rA, Src[0] = rB, Imm[0] = vX
		if rA, rB, vX, err := decodeTwoRegistersAndOneImmediate(idata, pc, skipLen); err == nil {
			instr.Dst = rA
			instr.Src[0] = rB
			instr.Imm[0] = vX
		}

	case InstrCatTwoRegOneOff:
		// 170-175: Src[0] = rA, Src[1] = rB, Imm[0] = target PC
		if rA, rB, target, err := decodeTwoRegistersAndOneOffset(idata, pc, skipLen); err == nil {
			instr.Src[0] = rA
			instr.Src[1] = rB
			instr.Imm[0] = uint64(target)
		}

	case InstrCatTwoRegTwoImm:
		// 180: Dst = rA, Src[0] = rB, Imm[0] = vX, Imm[1] = vY
		if rA, rB, vX, vY, err := decodeTwoRegistersAndTwoImmediates(idata, pc, skipLen); err == nil {
			instr.Dst = rA
			instr.Src[0] = rB
			instr.Imm[0] = vX
			instr.Imm[1] = vY
		}

	case InstrCatThreeReg:
		// 190-230: Dst = rD, Src[0] = rA, Src[1] = rB
		//   Note: decodeThreeRegisters returns (rA, rB, rD, err)
		if rA, rB, rD, err := decodeThreeRegisters(idata, pc); err == nil {
			instr.Dst = rD
			instr.Src[0] = rA
			instr.Src[1] = rB
		}
	}
}

// preDecodeBlocks performs a single-pass scan of the entire program blob,
// populating Program.Instrs, Program.BlockAt, and Program.InstrIdxAt.
// Called once at the end of DeBlobProgramCode.
func (p *Program) preDecodeBlocks() ExitReason {
	idata := p.InstructionData
	bitmask := p.Bitmasks
	n := len(idata)

	p.Instrs = make([]InstrMeta, 0, n/4)
	p.BlockAt = make([]*BlockMeta, n)
	p.InstrIdxAt = make([]int32, n)
	for i := range p.InstrIdxAt {
		p.InstrIdxAt[i] = -1
	}

	pc := ProgramCounter(0)
	for pc < ProgramCounter(n) {
		if !bitmask.IsStartOfBasicBlock(pc) {
			pc++
			continue
		}

		block := &BlockMeta{
			StartPC:    pc,
			InstrStart: len(p.Instrs),
		}

		for {
			if pc >= ProgramCounter(n) {
				return ExitPanic
			}
			op := idata[pc]
			if !IsValidOpcode(op) {
				return ExitPanic
			}

			skipLen := skip(int(pc), bitmask)

			idx := len(p.Instrs)
			p.Instrs = append(p.Instrs, InstrMeta{
				PC:      pc,
				Opcode:  op,
				SkipLen: uint8(skipLen),
				Exec:    instrMetaExecForOpcode(op),
			})
			p.InstrIdxAt[pc] = int32(idx)

			decodeOperands(&p.Instrs[idx], idata, bitmask)

			if IsBlockTerminator(op) {
				block.EndPC = pc
				block.InstrEnd = len(p.Instrs)
				block.GasCost = Gas(block.InstrEnd - block.InstrStart)
				p.BlockAt[block.StartPC] = block
				pc += ProgramCounter(skipLen) + 1
				break
			}

			pc += ProgramCounter(skipLen) + 1
		}
	}

	return ExitContinue
}

// LookupBlock returns the pre-decoded BlockMeta for a basic block starting at pc.
// Returns nil if pc is not the start of a known basic block.
func (p *Program) LookupBlock(pc ProgramCounter) *BlockMeta {
	if int(pc) >= len(p.BlockAt) {
		return nil
	}
	return p.BlockAt[pc]
}
