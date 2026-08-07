package PVM

const (
	iotaNone = ^ProgramCounter(0)

	robNone = 0
	robDEC  = 1
	robWAIT = 2
	robEXE  = 3
	robFIN  = 4
)

type robEntry struct {
	state      uint8
	cyclesLeft int
	deps       []int
	regs       regSet
	units      ExecUnits
}

func (u ExecUnits) fits(avail ExecUnits) bool {
	return u.A <= avail.A && u.L <= avail.L && u.S <= avail.S &&
		u.M <= avail.M && u.D <= avail.D
}

func (u ExecUnits) subFrom(avail *ExecUnits) {
	avail.A -= u.A
	avail.L -= u.L
	avail.S -= u.S
	avail.M -= u.M
	avail.D -= u.D
}

func (u ExecUnits) addTo(avail *ExecUnits) {
	avail.A += u.A
	avail.L += u.L
	avail.S += u.S
	avail.M += u.M
	avail.D += u.D
}

func (b *BlockState) robActiveCount() int {
	n := 0
	for _, e := range b.ROB {
		if e.state != robNone {
			n++
		}
	}
	return n
}

func (b *BlockState) instrAt(p *Program) *InstrMeta {
	if b.Iota == iotaNone {
		return nil
	}
	// GP eq:instructions: code is zero-padded; a fetch past the end is trap.
	if int(b.Iota) >= len(p.InstructionData) {
		return &InstrMeta{
			PC:      b.Iota,
			Opcode:  0, // trap
			SkipLen: 0,
			Dst:     0xFF,
			Src:     [2]uint8{0xFF, 0xFF},
		}
	}
	if int(b.Iota) >= len(p.InstrIdxAt) {
		return nil
	}
	idx := p.InstrIdxAt[b.Iota]
	if idx < 0 {
		return nil
	}
	return &p.Instrs[idx]
}

func (b *BlockState) canDecode(p *Program) bool {
	if b.Iota == iotaNone {
		return false
	}
	if b.robActiveCount() >= MaxROB {
		return false
	}
	instr := b.instrAt(p)
	if instr == nil {
		return false
	}
	cost := InstructionCost(p, instr)
	return cost.Decode <= b.DecodeSlots
}

func (b *BlockState) decodeMoveReg(p *Program, instr *InstrMeta) {
	srcRegs := instSrcRegs(instr)
	dstRegs := instDstRegs(instr)
	dstSet := regSetFrom(dstRegs)

	for j := range b.ROB {
		if b.ROB[j].state == robNone {
			continue
		}
		if b.ROB[j].regs.intersects(regSetFrom(srcRegs)) {
			b.ROB[j].regs = b.ROB[j].regs.union(dstSet)
		} else {
			b.ROB[j].regs.subtract(dstSet)
		}
	}

	b.Iota = b.nextIota(instr)
	b.DecodeSlots--
}

func (b *BlockState) nextIota(instr *InstrMeta) ProgramCounter {
	if IsBlockTerminator(instr.Opcode) {
		return iotaNone
	}
	return instr.PC + ProgramCounter(1+instr.SkipLen)
}

func appendROBDep(deps []int, j int) []int {
	for _, d := range deps {
		if d == j {
			return deps
		}
	}
	return append(deps, j)
}

func (b *BlockState) decodeToROB(p *Program, instr *InstrMeta) {
	dstRegs := instDstRegs(instr)
	srcRegs := instSrcRegs(instr)
	dstSet := regSetFrom(dstRegs)
	srcSet := regSetFrom(srcRegs)
	cost := InstructionCost(p, instr)

	// RAW deps against current ROB register tracking (before dst rename).
	// When dst∩src ≠ ∅ this also captures the prior writer of dst.
	var deps []int
	for j, e := range b.ROB {
		if e.state != robNone && e.regs.intersects(srcSet) {
			deps = appendROBDep(deps, j)
		}
	}
	// Rename: drop dst from prior ROB entries.
	for j := range b.ROB {
		if b.ROB[j].state != robNone {
			b.ROB[j].regs.subtract(dstSet)
		}
	}

	b.ROB = append(b.ROB, robEntry{
		state:      robDEC,
		cyclesLeft: cost.Cycles,
		deps:       deps,
		regs:       dstSet,
		units:      cost.Units,
	})

	b.Iota = b.nextIota(instr)
	b.DecodeSlots -= cost.Decode
}

func (b *BlockState) decode(p *Program) {
	instr := b.instrAt(p)
	if instr == nil {
		b.Iota = iotaNone
		return
	}
	if instr.Opcode == 100 {
		b.decodeMoveReg(p, instr)
	} else {
		b.decodeToROB(p, instr)
	}
}

func (b *BlockState) findReady() int {
	for j, e := range b.ROB {
		if e.state != robWAIT {
			continue
		}
		if !e.units.fits(b.UnitsAvail) {
			continue
		}
		ready := true
		for _, dep := range e.deps {
			if dep >= len(b.ROB) || b.ROB[dep].state == robNone {
				continue
			}
			// Dependency satisfied when producer finished executing (cyclesLeft==0),
			// not when it has been purged from the ROB.
			if b.ROB[dep].cyclesLeft != 0 {
				ready = false
				break
			}
		}
		if ready {
			return j
		}
	}
	return -1
}

func (b *BlockState) startExec(j int) {
	b.ROB[j].state = robEXE
	b.ROB[j].units.subFrom(&b.UnitsAvail)
	b.ExecutionSlots--
}

func (b *BlockState) advanceCycle() {
	// In-order retire: purge leading FIN entries, then progress EXE/DEC.
	for j := range b.ROB {
		if b.ROB[j].state == robNone {
			continue
		}
		if b.ROB[j].state != robFIN {
			break
		}
		b.ROB[j].state = robNone
	}
	b.trimLeadingNone()

	// GP: return units for EXE entries with cyclesLeft==1 (about to hit 0).
	var returned ExecUnits
	for j := range b.ROB {
		e := &b.ROB[j]
		if e.state == robEXE && e.cyclesLeft == 1 {
			returned.A += e.units.A
			returned.L += e.units.L
			returned.S += e.units.S
			returned.M += e.units.M
			returned.D += e.units.D
		}
	}

	for j := range b.ROB {
		e := &b.ROB[j]
		switch e.state {
		case robDEC:
			e.state = robWAIT
		case robEXE:
			if e.cyclesLeft > 0 {
				// After cyclesLeft reaches 0, remain EXE for one more cycle
				// before transitioning to FIN.
				e.cyclesLeft--
			} else {
				e.state = robFIN
			}
		}
	}

	returned.addTo(&b.UnitsAvail)
	b.Cyc++
	b.DecodeSlots = DecodeWidth
	b.ExecutionSlots = ExecutionWidth
}

// trimLeadingNone drops retired leading ROB slots and remaps dependency indices
// so physical ROB length stays proportional to MaxROB (active bound).
func (b *BlockState) trimLeadingNone() {
	n := 0
	for n < len(b.ROB) && b.ROB[n].state == robNone {
		n++
	}
	if n == 0 {
		return
	}
	b.ROB = b.ROB[n:]
	for i := range b.ROB {
		if len(b.ROB[i].deps) == 0 {
			continue
		}
		deps := make([]int, 0, len(b.ROB[i].deps))
		for _, d := range b.ROB[i].deps {
			if d < n {
				continue
			}
			deps = append(deps, d-n)
		}
		b.ROB[i].deps = deps
	}
}

// simulateBlock runs A.9 gas_sim until the basic block converges.
// A non-converging simulation is a logic bug — never return a partial gas value.
func (b *BlockState) simulateBlock(p *Program) Gas {
	const maxSteps = 100000
	for step := 0; step < maxSteps; step++ {
		if b.canDecode(p) {
			b.decode(p)
			continue
		}
		// Open block (no terminator): iota advanced past the last instruction.
		if b.Iota != iotaNone && b.instrAt(p) == nil {
			b.Iota = iotaNone
			continue
		}
		if j := b.findReady(); j >= 0 && b.ExecutionSlots > 0 {
			b.startExec(j)
			continue
		}
		if b.Iota == iotaNone && b.robActiveCount() == 0 {
			return blockGasFromCycles(b.Cyc)
		}
		b.advanceCycle()
	}
	panic("gas_sim: basic block did not converge within step limit")
}

// blockGasFromCycles implements eq:gascostforblock = max(cycles - 3, 1).
func blockGasFromCycles(cycles Gas) Gas {
	if cycles <= 3 {
		return 1
	}
	return cycles - 3
}

// invalidBlockGas is the A.9 cost of a trap-equivalent invalid instruction
// (trap timeline DeeER → max(5−3, 1) = 2), used for trailing block entries
// that appear after a final terminator but are outside code.
const invalidBlockGas Gas = 2

// GasCostForBlock returns A.9 gascostforblock for the basic block at startPC.
func GasCostForBlock(p *Program, startPC ProgramCounter) Gas {
	if int(startPC) >= len(p.InstrIdxAt) {
		// Past end of code: treat as an invalid/trap block.
		return invalidBlockGas
	}
	if p.InstrIdxAt[startPC] < 0 {
		return 1
	}
	state := InitBlockState(startPC)
	return state.simulateBlock(p)
}

// GasCostFromPC returns A.9 gascostforblock for the suffix from pc through the
// end of its containing basic block (used when resuming mid-block).
func GasCostFromPC(p *Program, pc ProgramCounter) Gas {
	if int(pc) >= len(p.InstrIdxAt) {
		return invalidBlockGas
	}
	if p.InstrIdxAt[pc] < 0 {
		return 1
	}
	state := InitBlockState(pc)
	return state.simulateBlock(p)
}
