package PVM

// InstrCategory classifies PVM opcodes by their operand encoding format (GP A.5).
type InstrCategory uint8

const (
	InstrCatInvalid      InstrCategory = iota // not a valid opcode
	InstrCatNoArg                             // 0, 1
	InstrCatOneImm                            // 10
	InstrCatOneRegExtImm                      // 20
	InstrCatTwoImm                            // 30-33
	InstrCatOneOffset                         // 40
	InstrCatOneRegOneImm                      // 50-62
	InstrCatOneRegTwoImm                      // 70-73
	InstrCatOneRegImmOff                      // 80-90
	InstrCatTwoReg                            // 100-111
	InstrCatTwoRegOneImm                      // 120-161
	InstrCatTwoRegOneOff                      // 170-175
	InstrCatTwoRegTwoImm                      // 180
	InstrCatThreeReg                          // 190-230
)

// OpcodeInfo holds the static, immutable properties of a PVM opcode.
type OpcodeInfo struct {
	Name         string
	Category     InstrCategory
	IsTerminator bool // ends a basic block
	IsLoad       bool // guest memory read (μ)
	IsStore      bool // guest memory write (μ)
	// TODO(gas-model): add OpcodeResource when integrating GP v0.8.0 gas cost model.
	// Resource OpcodeResource  // cycles, decode slots, exec units (A.10)
}

// opcodeInfoTable is indexed by the raw opcode byte (0–255).
// Invalid opcodes have zero-value entries (Category == InstrCatInvalid).
var opcodeInfoTable = [256]OpcodeInfo{
	// A.5.1 No-argument (terminators)
	0: {Name: "trap", Category: InstrCatNoArg, IsTerminator: true},
	1: {Name: "fallthrough", Category: InstrCatNoArg, IsTerminator: true},

	// A.5.2 One immediate
	10: {Name: "ecalli", Category: InstrCatOneImm, IsTerminator: false},

	// A.5.3 One reg + extended-width immediate
	20: {Name: "load_imm_64", Category: InstrCatOneRegExtImm, IsTerminator: false},

	// A.5.4 Two immediates (store_imm)
	30: {Name: "store_imm_u8", Category: InstrCatTwoImm, IsTerminator: false, IsStore: true},
	31: {Name: "store_imm_u16", Category: InstrCatTwoImm, IsTerminator: false, IsStore: true},
	32: {Name: "store_imm_u32", Category: InstrCatTwoImm, IsTerminator: false, IsStore: true},
	33: {Name: "store_imm_u64", Category: InstrCatTwoImm, IsTerminator: false, IsStore: true},

	// A.5.5 One offset (terminators)
	40: {Name: "jump", Category: InstrCatOneOffset, IsTerminator: true},

	// A.5.6 One reg + one imm
	50: {Name: "jump_ind", Category: InstrCatOneRegOneImm, IsTerminator: true},
	51: {Name: "load_imm", Category: InstrCatOneRegOneImm, IsTerminator: false},
	52: {Name: "load_u8", Category: InstrCatOneRegOneImm, IsTerminator: false, IsLoad: true},
	53: {Name: "load_i8", Category: InstrCatOneRegOneImm, IsTerminator: false, IsLoad: true},
	54: {Name: "load_u16", Category: InstrCatOneRegOneImm, IsTerminator: false, IsLoad: true},
	55: {Name: "load_i16", Category: InstrCatOneRegOneImm, IsTerminator: false, IsLoad: true},
	56: {Name: "load_u32", Category: InstrCatOneRegOneImm, IsTerminator: false, IsLoad: true},
	57: {Name: "load_i32", Category: InstrCatOneRegOneImm, IsTerminator: false, IsLoad: true},
	58: {Name: "load_u64", Category: InstrCatOneRegOneImm, IsTerminator: false, IsLoad: true},
	59: {Name: "store_u8", Category: InstrCatOneRegOneImm, IsTerminator: false, IsStore: true},
	60: {Name: "store_u16", Category: InstrCatOneRegOneImm, IsTerminator: false, IsStore: true},
	61: {Name: "store_u32", Category: InstrCatOneRegOneImm, IsTerminator: false, IsStore: true},
	62: {Name: "store_u64", Category: InstrCatOneRegOneImm, IsTerminator: false, IsStore: true},

	// A.5.7 One reg + two imm (store_imm_ind)
	70: {Name: "store_imm_ind_u8", Category: InstrCatOneRegTwoImm, IsTerminator: false, IsStore: true},
	71: {Name: "store_imm_ind_u16", Category: InstrCatOneRegTwoImm, IsTerminator: false, IsStore: true},
	72: {Name: "store_imm_ind_u32", Category: InstrCatOneRegTwoImm, IsTerminator: false, IsStore: true},
	73: {Name: "store_imm_ind_u64", Category: InstrCatOneRegTwoImm, IsTerminator: false, IsStore: true},

	// A.5.8 One reg + imm + offset (terminators)
	80: {Name: "load_imm_jump", Category: InstrCatOneRegImmOff, IsTerminator: true},
	81: {Name: "branch_eq_imm", Category: InstrCatOneRegImmOff, IsTerminator: true},
	82: {Name: "branch_ne_imm", Category: InstrCatOneRegImmOff, IsTerminator: true},
	83: {Name: "branch_lt_u_imm", Category: InstrCatOneRegImmOff, IsTerminator: true},
	84: {Name: "branch_le_u_imm", Category: InstrCatOneRegImmOff, IsTerminator: true},
	85: {Name: "branch_ge_u_imm", Category: InstrCatOneRegImmOff, IsTerminator: true},
	86: {Name: "branch_gt_u_imm", Category: InstrCatOneRegImmOff, IsTerminator: true},
	87: {Name: "branch_lt_s_imm", Category: InstrCatOneRegImmOff, IsTerminator: true},
	88: {Name: "branch_le_s_imm", Category: InstrCatOneRegImmOff, IsTerminator: true},
	89: {Name: "branch_ge_s_imm", Category: InstrCatOneRegImmOff, IsTerminator: true},
	90: {Name: "branch_gt_s_imm", Category: InstrCatOneRegImmOff, IsTerminator: true},

	// A.5.9 Two registers
	100: {Name: "move_reg", Category: InstrCatTwoReg, IsTerminator: false},
	101: {Name: "sbrk", Category: InstrCatTwoReg, IsTerminator: false},
	102: {Name: "count_set_bits_64", Category: InstrCatTwoReg, IsTerminator: false},
	103: {Name: "count_set_bits_32", Category: InstrCatTwoReg, IsTerminator: false},
	104: {Name: "leading_zero_bits_64", Category: InstrCatTwoReg, IsTerminator: false},
	105: {Name: "leading_zero_bits_32", Category: InstrCatTwoReg, IsTerminator: false},
	106: {Name: "trailing_zero_bits_64", Category: InstrCatTwoReg, IsTerminator: false},
	107: {Name: "trailing_zero_bits_32", Category: InstrCatTwoReg, IsTerminator: false},
	108: {Name: "sign_extend_8", Category: InstrCatTwoReg, IsTerminator: false},
	109: {Name: "sign_extend_16", Category: InstrCatTwoReg, IsTerminator: false},
	110: {Name: "zero_extend_16", Category: InstrCatTwoReg, IsTerminator: false},
	111: {Name: "reverse_bytes", Category: InstrCatTwoReg, IsTerminator: false},

	// A.5.10 Two reg + one imm (store_ind, load_ind, arithmetic)
	120: {Name: "store_ind_u8", Category: InstrCatTwoRegOneImm, IsTerminator: false, IsStore: true},
	121: {Name: "store_ind_u16", Category: InstrCatTwoRegOneImm, IsTerminator: false, IsStore: true},
	122: {Name: "store_ind_u32", Category: InstrCatTwoRegOneImm, IsTerminator: false, IsStore: true},
	123: {Name: "store_ind_u64", Category: InstrCatTwoRegOneImm, IsTerminator: false, IsStore: true},
	124: {Name: "load_ind_u8", Category: InstrCatTwoRegOneImm, IsTerminator: false, IsLoad: true},
	125: {Name: "load_ind_i8", Category: InstrCatTwoRegOneImm, IsTerminator: false, IsLoad: true},
	126: {Name: "load_ind_u16", Category: InstrCatTwoRegOneImm, IsTerminator: false, IsLoad: true},
	127: {Name: "load_ind_i16", Category: InstrCatTwoRegOneImm, IsTerminator: false, IsLoad: true},
	128: {Name: "load_ind_u32", Category: InstrCatTwoRegOneImm, IsTerminator: false, IsLoad: true},
	129: {Name: "load_ind_i32", Category: InstrCatTwoRegOneImm, IsTerminator: false, IsLoad: true},
	130: {Name: "load_ind_u64", Category: InstrCatTwoRegOneImm, IsTerminator: false, IsLoad: true},
	131: {Name: "add_imm_32", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	132: {Name: "and_imm", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	133: {Name: "xor_imm", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	134: {Name: "or_imm", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	135: {Name: "mul_imm_32", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	136: {Name: "set_lt_u_imm", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	137: {Name: "set_lt_s_imm", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	138: {Name: "shlo_l_imm_32", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	139: {Name: "shlo_r_imm_32", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	140: {Name: "shar_r_imm_32", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	141: {Name: "neg_add_imm_32", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	142: {Name: "set_gt_u_imm", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	143: {Name: "set_gt_s_imm", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	144: {Name: "shlo_l_imm_alt_32", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	145: {Name: "shlo_r_imm_alt_32", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	146: {Name: "shar_r_imm_alt_32", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	147: {Name: "cmov_iz_imm", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	148: {Name: "cmov_nz_imm", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	149: {Name: "add_imm_64", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	150: {Name: "mul_imm_64", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	151: {Name: "shlo_l_imm_64", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	152: {Name: "shlo_r_imm_64", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	153: {Name: "shar_r_imm_64", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	154: {Name: "neg_add_imm_64", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	155: {Name: "shlo_l_imm_alt_64", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	156: {Name: "shlo_r_imm_alt_64", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	157: {Name: "shar_r_imm_alt_64", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	158: {Name: "rot_r_64_imm", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	159: {Name: "rot_r_64_imm_alt", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	160: {Name: "rot_r_32_imm", Category: InstrCatTwoRegOneImm, IsTerminator: false},
	161: {Name: "rot_r_32_imm_alt", Category: InstrCatTwoRegOneImm, IsTerminator: false},

	// A.5.11 Two reg + one offset (terminators)
	170: {Name: "branch_eq", Category: InstrCatTwoRegOneOff, IsTerminator: true},
	171: {Name: "branch_ne", Category: InstrCatTwoRegOneOff, IsTerminator: true},
	172: {Name: "branch_lt_u", Category: InstrCatTwoRegOneOff, IsTerminator: true},
	173: {Name: "branch_lt_s", Category: InstrCatTwoRegOneOff, IsTerminator: true},
	174: {Name: "branch_ge_u", Category: InstrCatTwoRegOneOff, IsTerminator: true},
	175: {Name: "branch_ge_s", Category: InstrCatTwoRegOneOff, IsTerminator: true},

	// A.5.12 Two reg + two imm (terminator)
	180: {Name: "load_imm_jump_ind", Category: InstrCatTwoRegTwoImm, IsTerminator: true},

	// A.5.13 Three registers
	190: {Name: "add_32", Category: InstrCatThreeReg, IsTerminator: false},
	191: {Name: "sub_32", Category: InstrCatThreeReg, IsTerminator: false},
	192: {Name: "mul_32", Category: InstrCatThreeReg, IsTerminator: false},
	193: {Name: "div_u_32", Category: InstrCatThreeReg, IsTerminator: false},
	194: {Name: "div_s_32", Category: InstrCatThreeReg, IsTerminator: false},
	195: {Name: "rem_u_32", Category: InstrCatThreeReg, IsTerminator: false},
	196: {Name: "rem_s_32", Category: InstrCatThreeReg, IsTerminator: false},
	197: {Name: "shlo_l_32", Category: InstrCatThreeReg, IsTerminator: false},
	198: {Name: "shlo_r_32", Category: InstrCatThreeReg, IsTerminator: false},
	199: {Name: "shar_r_32", Category: InstrCatThreeReg, IsTerminator: false},
	200: {Name: "add_64", Category: InstrCatThreeReg, IsTerminator: false},
	201: {Name: "sub_64", Category: InstrCatThreeReg, IsTerminator: false},
	202: {Name: "mul_64", Category: InstrCatThreeReg, IsTerminator: false},
	203: {Name: "div_u_64", Category: InstrCatThreeReg, IsTerminator: false},
	204: {Name: "div_s_64", Category: InstrCatThreeReg, IsTerminator: false},
	205: {Name: "rem_u_64", Category: InstrCatThreeReg, IsTerminator: false},
	206: {Name: "rem_s_64", Category: InstrCatThreeReg, IsTerminator: false},
	207: {Name: "shlo_l_64", Category: InstrCatThreeReg, IsTerminator: false},
	208: {Name: "shlo_r_64", Category: InstrCatThreeReg, IsTerminator: false},
	209: {Name: "shar_r_64", Category: InstrCatThreeReg, IsTerminator: false},
	210: {Name: "and", Category: InstrCatThreeReg, IsTerminator: false},
	211: {Name: "xor", Category: InstrCatThreeReg, IsTerminator: false},
	212: {Name: "or", Category: InstrCatThreeReg, IsTerminator: false},
	213: {Name: "mul_upper_s_s", Category: InstrCatThreeReg, IsTerminator: false},
	214: {Name: "mul_upper_u_u", Category: InstrCatThreeReg, IsTerminator: false},
	215: {Name: "mul_upper_s_u", Category: InstrCatThreeReg, IsTerminator: false},
	216: {Name: "set_lt_u", Category: InstrCatThreeReg, IsTerminator: false},
	217: {Name: "set_lt_s", Category: InstrCatThreeReg, IsTerminator: false},
	218: {Name: "cmov_iz", Category: InstrCatThreeReg, IsTerminator: false},
	219: {Name: "cmov_nz", Category: InstrCatThreeReg, IsTerminator: false},
	220: {Name: "rot_l_64", Category: InstrCatThreeReg, IsTerminator: false},
	221: {Name: "rot_l_32", Category: InstrCatThreeReg, IsTerminator: false},
	222: {Name: "rot_r_64", Category: InstrCatThreeReg, IsTerminator: false},
	223: {Name: "rot_r_32", Category: InstrCatThreeReg, IsTerminator: false},
	224: {Name: "and_inv", Category: InstrCatThreeReg, IsTerminator: false},
	225: {Name: "or_inv", Category: InstrCatThreeReg, IsTerminator: false},
	226: {Name: "xnor", Category: InstrCatThreeReg, IsTerminator: false},
	227: {Name: "max", Category: InstrCatThreeReg, IsTerminator: false},
	228: {Name: "max_u", Category: InstrCatThreeReg, IsTerminator: false},
	229: {Name: "min", Category: InstrCatThreeReg, IsTerminator: false},
	230: {Name: "min_u", Category: InstrCatThreeReg, IsTerminator: false},
}

func IsValidOpcode(op byte) bool {
	return opcodeInfoTable[op].Category != InstrCatInvalid
}

// IsBlockTerminator reports whether op ends a basic block.
func IsBlockTerminator(op byte) bool {
	return opcodeInfoTable[op].IsTerminator
}

// GetOpcodeInfo returns the static metadata for a given opcode.
func GetOpcodeInfo(op byte) *OpcodeInfo {
	return &opcodeInfoTable[op]
}

// OpcodeName returns the mnemonic for debugging/logging.
// Returns "" for invalid opcodes.
func OpcodeName(op byte) string {
	return opcodeInfoTable[op].Name
}

// IsLoadOpcode reports whether the opcode performs a guest memory read.
func IsLoadOpcode(op byte) bool {
	return opcodeInfoTable[op].IsLoad
}

// IsStoreOpcode reports whether the opcode performs a guest memory write.
func IsStoreOpcode(op byte) bool {
	return opcodeInfoTable[op].IsStore
}

// MemAccessWidth returns the guest memory access width in bytes for load/store
// opcodes. The second return is false for non-memory opcodes.
func MemAccessWidth(op byte) (int, bool) {
	info := opcodeInfoTable[op]
	if !info.IsLoad && !info.IsStore {
		return 0, false
	}
	switch op {
	case 30, 52, 59, 70, 120, 124, 125:
		return 1, true
	case 31, 53, 55, 60, 71, 121, 126, 127:
		return 2, true
	case 32, 54, 57, 61, 72, 122, 128, 129:
		return 4, true
	case 33, 58, 62, 73, 123, 130:
		return 8, true
	default:
		return 0, false
	}
}
