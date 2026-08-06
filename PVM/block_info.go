package PVM

import "encoding/binary"

// InstrMeta holds pre-decoded metadata for a single PVM instruction.
// Populated once at deblob time; never mutated afterwards.
type InstrMeta struct {
	PC         ProgramCounter // 4B
	Opcode     byte           // 1B
	SkipLen    uint8          // 1B, max 24
	Dst        uint8          // 1B, destination reg index (0xFF = none)
	Src        [2]uint8       // 2B, source reg indices (0xFF = unused)
	BlockStart ProgramCounter // 4B, 𝔏(PC) (A.6): start of the enclosing basic block
	Exec       instrMetaFn    // pre-resolved handler; set at deblob time
	Imm        [2]uint64      // 16B, immediates / branch target PC
}

// BlockMeta holds pre-decoded metadata for a single PVM basic block.
// Populated once at deblob time; never mutated afterwards.
type BlockMeta struct {
	StartPC    ProgramCounter
	EndPC      ProgramCounter // PC of the terminating instruction (inclusive)
	InstrStart int            // index into Program.Instrs[]
	InstrEnd   int            // exclusive upper bound into Program.Instrs[]
	GasCost    Gas            // A.9 gascostforblock = max(cycles − 3, 1)
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

// preDecodeBlocks builds InstrMeta / BlockMeta and caches A.9 gas per block.
// If code ends without a terminator, the prefix is still emitted; bad PCs panic
// at execution. Mid-stream 𝔳_inst failures remain fatal.
func (p *Program) preDecodeBlocks() ExitReason {
	idata := p.InstructionData
	bitmask := p.Bitmasks
	n := len(idata)

	if len(bitmask) != n || n == 0 {
		return ExitPanic
	}

	p.Instrs = make([]InstrMeta, 0, n/4)
	p.BlockAt = make([]*BlockMeta, n)
	p.InstrIdxAt = make([]int32, n)
	for i := range p.InstrIdxAt {
		p.InstrIdxAt[i] = -1
	}

	blockStartPC := 0
	blockInstrStart := 0

	// emitBlock writes BlockMeta + gas for [blockStartPC, endPC].
	// Used on terminators and at EOF when the last op is not a terminator.
	emitBlock := func(endPC int) {
		block := &BlockMeta{
			StartPC:    ProgramCounter(blockStartPC),
			EndPC:      ProgramCounter(endPC),
			InstrStart: blockInstrStart,
			InstrEnd:   len(p.Instrs),
		}
		block.GasCost = GasCostForBlock(p, block.StartPC)
		for i := block.InstrStart; i < block.InstrEnd; i++ {
			p.Instrs[i].BlockStart = block.StartPC
		}
		p.BlockAt[blockStartPC] = block
	}

	for pc := 0; ; {
		if pc >= n {
			// Code ended mid-block (no terminator): keep the prefix block.
			if blockInstrStart < len(p.Instrs) {
				emitBlock(int(p.Instrs[len(p.Instrs)-1].PC))
			}
			return ExitContinue
		}

		skipLen := skip(pc, bitmask)
		next := pc + 1 + int(skipLen)

		// 𝔳_inst: every step of the walk lands on a defined instruction.
		if !validInst(idata, bitmask, uint64(pc)) {
			return ExitPanic
		}

		op := idata[pc]
		idx := len(p.Instrs)
		p.Instrs = append(p.Instrs, InstrMeta{
			PC:      ProgramCounter(pc),
			Opcode:  op,
			SkipLen: uint8(skipLen),
			Exec:    instrMetaExecForOpcode(op),
		})
		p.InstrIdxAt[pc] = int32(idx)

		decodeOperands(&p.Instrs[idx], idata, bitmask)

		if IsBlockTerminator(op) {
			emitBlock(pc)
			// ϖ | A.3: the index following a terminator starts the next block.
			blockStartPC = next
			blockInstrStart = len(p.Instrs)
		}

		pc = next
	}
}

// LookupBlock returns the pre-decoded BlockMeta for a basic block starting at pc.
// Returns nil if pc is not the start of a known basic block.
func (p *Program) LookupBlock(pc ProgramCounter) *BlockMeta {
	if int(pc) >= len(p.BlockAt) {
		return nil
	}
	return p.BlockAt[pc]
}

// StartOfBasicBlock is 𝔏(ι) | A.3: the start of the basic block containing pc.
// Reports false when pc is not an instruction start, for which 𝔏 is undefined.
func (p *Program) StartOfBasicBlock(pc ProgramCounter) (ProgramCounter, bool) {
	if int(pc) >= len(p.InstrIdxAt) {
		return 0, false
	}
	idx := p.InstrIdxAt[pc]
	if idx < 0 {
		return 0, false
	}
	return p.Instrs[idx].BlockStart, true
}

// BlockContaining returns the BlockMeta whose instruction range includes pc.
// LookupBlock only works at block entry PCs; this resolves mid-block resume PCs
// (e.g. after a host-call returns to Go and continues at the next instruction).
func (p *Program) BlockContaining(pc ProgramCounter) *BlockMeta {
	if b := p.LookupBlock(pc); b != nil {
		return b
	}
	start, ok := p.StartOfBasicBlock(pc)
	if !ok {
		return nil
	}
	return p.LookupBlock(start)
}
