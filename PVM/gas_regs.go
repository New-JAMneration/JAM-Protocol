package PVM

// instSrcRegs returns source registers for gas modelling (A.9 / A.10).
// Values follow the instruction tables' read set (inst_srcregs), which may differ
// from decode-time InstrMeta.Dst/Src packing used by the interpreter.
func instSrcRegs(instr *InstrMeta) []Reg {
	switch instr.Opcode {
	case 10: // ecalli
		return nil
	case 20, 51: // load_imm_64, load_imm — dst only
		return nil
	case 30, 31, 32, 33: // store_imm_* — immediates only
		return nil
	case 40, 80: // jump, load_imm_jump — no register reads in the gas model
		return nil
	case 50: // jump_ind — base (decoder packs rA into Dst)
		return regOrNil(instr.Dst)
	case 52, 53, 54, 55, 56, 57, 58: // absolute loads — address is immediate
		return nil
	case 59, 60, 61, 62: // store_u* — value (decoder packs rA into Dst)
		return regOrNil(instr.Dst)
	case 70, 71, 72, 73: // store_imm_ind_* — base (decoder packs rA into Dst)
		return regOrNil(instr.Dst)
	case 81, 82, 83, 84, 85, 86, 87, 88, 89, 90: // branch_*_imm — compare reg (decoder packs rA into Dst)
		return regOrNil(instr.Dst)
	case 120, 121, 122, 123: // store_ind_* — value (Dst field) + base (Src[0])
		var regs []Reg
		if instr.Dst != 0xFF {
			regs = append(regs, Reg(instr.Dst))
		}
		if instr.Src[0] != 0xFF {
			regs = append(regs, Reg(instr.Src[0]))
		}
		return regs
	case 180: // load_imm_jump_ind — base only in the gas model
		return regOrNil(instr.Src[0])
	default:
		var regs []Reg
		for _, s := range instr.Src {
			if s != 0xFF {
				regs = append(regs, Reg(s))
			}
		}
		return regs
	}
}

// instDstRegs returns destination registers for gas modelling (A.9 / A.10).
// Stores and branches have an empty write set (memory / control-flow only).
func instDstRegs(instr *InstrMeta) []Reg {
	switch instr.Opcode {
	case 10: // ecalli
		return nil
	case 30, 31, 32, 33: // store_imm_*
		return nil
	case 40, 50, 80, 180: // jump / jump_ind / load_imm_jump*
		return nil
	case 59, 60, 61, 62: // store_u*
		return nil
	case 70, 71, 72, 73: // store_imm_ind_*
		return nil
	case 81, 82, 83, 84, 85, 86, 87, 88, 89, 90: // branch_*_imm
		return nil
	case 120, 121, 122, 123: // store_ind_*
		return nil
	case 170, 171, 172, 173, 174, 175: // branch_*
		return nil
	default:
		return regOrNil(instr.Dst)
	}
}

func regOrNil(r uint8) []Reg {
	if r == 0xFF {
		return nil
	}
	return []Reg{Reg(r)}
}

type regSet map[Reg]struct{}

func regSetFrom(regs []Reg) regSet {
	if len(regs) == 0 {
		return nil
	}
	s := make(regSet, len(regs))
	for _, r := range regs {
		s[r] = struct{}{}
	}
	return s
}

func (s regSet) intersects(other regSet) bool {
	if len(s) == 0 || len(other) == 0 {
		return false
	}
	for r := range s {
		if _, ok := other[r]; ok {
			return true
		}
	}
	return false
}

func (s regSet) union(other regSet) regSet {
	if len(other) == 0 {
		return s
	}
	if s == nil {
		s = make(regSet, len(other))
	}
	for r := range other {
		s[r] = struct{}{}
	}
	return s
}

func (s regSet) subtract(other regSet) {
	for r := range other {
		delete(s, r)
	}
}

func regsOverlap(dst, src []Reg) bool {
	return regSetFrom(dst).intersects(regSetFrom(src))
}
