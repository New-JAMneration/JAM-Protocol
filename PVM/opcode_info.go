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
// All fields are determined at init-time and never change.
type OpcodeInfo struct {
	Name         string
	Category     InstrCategory
	IsTerminator bool // ends a basic block
	// TODO(gas-model): add OpcodeResource when integrating GP v0.8.0 gas cost model.
	// Resource OpcodeResource  // cycles, decode slots, exec units (A.10)
}

// opcodeInfoTable is indexed by the raw opcode byte (0–255).
// Invalid opcodes have zero-value entries (Category == InstrCatInvalid).
var opcodeInfoTable = [256]OpcodeInfo{
	// A.5.1 No-argument (terminators)
	0: {"trap", InstrCatNoArg, true},
	1: {"fallthrough", InstrCatNoArg, true},

	// A.5.2 One immediate
	10: {"ecalli", InstrCatOneImm, false},

	// A.5.3 One reg + extended-width immediate
	20: {"load_imm_64", InstrCatOneRegExtImm, false},

	// A.5.4 Two immediates (store_imm)
	30: {"store_imm_u8", InstrCatTwoImm, false},
	31: {"store_imm_u16", InstrCatTwoImm, false},
	32: {"store_imm_u32", InstrCatTwoImm, false},
	33: {"store_imm_u64", InstrCatTwoImm, false},

	// A.5.5 One offset (terminators)
	40: {"jump", InstrCatOneOffset, true},

	// A.5.6 One reg + one imm
	50: {"jump_ind", InstrCatOneRegOneImm, true},
	51: {"load_imm", InstrCatOneRegOneImm, false},
	52: {"load_u8", InstrCatOneRegOneImm, false},
	53: {"load_i8", InstrCatOneRegOneImm, false},
	54: {"load_u16", InstrCatOneRegOneImm, false},
	55: {"load_i16", InstrCatOneRegOneImm, false},
	56: {"load_u32", InstrCatOneRegOneImm, false},
	57: {"load_i32", InstrCatOneRegOneImm, false},
	58: {"load_u64", InstrCatOneRegOneImm, false},
	59: {"store_u8", InstrCatOneRegOneImm, false},
	60: {"store_u16", InstrCatOneRegOneImm, false},
	61: {"store_u32", InstrCatOneRegOneImm, false},
	62: {"store_u64", InstrCatOneRegOneImm, false},

	// A.5.7 One reg + two imm (store_imm_ind)
	70: {"store_imm_ind_u8", InstrCatOneRegTwoImm, false},
	71: {"store_imm_ind_u16", InstrCatOneRegTwoImm, false},
	72: {"store_imm_ind_u32", InstrCatOneRegTwoImm, false},
	73: {"store_imm_ind_u64", InstrCatOneRegTwoImm, false},

	// A.5.8 One reg + imm + offset (terminators)
	80: {"load_imm_jump", InstrCatOneRegImmOff, true},
	81: {"branch_eq_imm", InstrCatOneRegImmOff, true},
	82: {"branch_ne_imm", InstrCatOneRegImmOff, true},
	83: {"branch_lt_u_imm", InstrCatOneRegImmOff, true},
	84: {"branch_le_u_imm", InstrCatOneRegImmOff, true},
	85: {"branch_ge_u_imm", InstrCatOneRegImmOff, true},
	86: {"branch_gt_u_imm", InstrCatOneRegImmOff, true},
	87: {"branch_lt_s_imm", InstrCatOneRegImmOff, true},
	88: {"branch_le_s_imm", InstrCatOneRegImmOff, true},
	89: {"branch_ge_s_imm", InstrCatOneRegImmOff, true},
	90: {"branch_gt_s_imm", InstrCatOneRegImmOff, true},

	// A.5.9 Two registers
	100: {"move_reg", InstrCatTwoReg, false},
	101: {"sbrk", InstrCatTwoReg, false},
	102: {"count_set_bits_64", InstrCatTwoReg, false},
	103: {"count_set_bits_32", InstrCatTwoReg, false},
	104: {"leading_zero_bits_64", InstrCatTwoReg, false},
	105: {"leading_zero_bits_32", InstrCatTwoReg, false},
	106: {"trailing_zero_bits_64", InstrCatTwoReg, false},
	107: {"trailing_zero_bits_32", InstrCatTwoReg, false},
	108: {"sign_extend_8", InstrCatTwoReg, false},
	109: {"sign_extend_16", InstrCatTwoReg, false},
	110: {"zero_extend_16", InstrCatTwoReg, false},
	111: {"reverse_bytes", InstrCatTwoReg, false},

	// A.5.10 Two reg + one imm (store_ind, load_ind, arithmetic)
	120: {"store_ind_u8", InstrCatTwoRegOneImm, false},
	121: {"store_ind_u16", InstrCatTwoRegOneImm, false},
	122: {"store_ind_u32", InstrCatTwoRegOneImm, false},
	123: {"store_ind_u64", InstrCatTwoRegOneImm, false},
	124: {"load_ind_u8", InstrCatTwoRegOneImm, false},
	125: {"load_ind_i8", InstrCatTwoRegOneImm, false},
	126: {"load_ind_u16", InstrCatTwoRegOneImm, false},
	127: {"load_ind_i16", InstrCatTwoRegOneImm, false},
	128: {"load_ind_u32", InstrCatTwoRegOneImm, false},
	129: {"load_ind_i32", InstrCatTwoRegOneImm, false},
	130: {"load_ind_u64", InstrCatTwoRegOneImm, false},
	131: {"add_imm_32", InstrCatTwoRegOneImm, false},
	132: {"and_imm", InstrCatTwoRegOneImm, false},
	133: {"xor_imm", InstrCatTwoRegOneImm, false},
	134: {"or_imm", InstrCatTwoRegOneImm, false},
	135: {"mul_imm_32", InstrCatTwoRegOneImm, false},
	136: {"set_lt_u_imm", InstrCatTwoRegOneImm, false},
	137: {"set_lt_s_imm", InstrCatTwoRegOneImm, false},
	138: {"shlo_l_imm_32", InstrCatTwoRegOneImm, false},
	139: {"shlo_r_imm_32", InstrCatTwoRegOneImm, false},
	140: {"shar_r_imm_32", InstrCatTwoRegOneImm, false},
	141: {"neg_add_imm_32", InstrCatTwoRegOneImm, false},
	142: {"set_gt_u_imm", InstrCatTwoRegOneImm, false},
	143: {"set_gt_s_imm", InstrCatTwoRegOneImm, false},
	144: {"shlo_l_imm_alt_32", InstrCatTwoRegOneImm, false},
	145: {"shlo_r_imm_alt_32", InstrCatTwoRegOneImm, false},
	146: {"shar_r_imm_alt_32", InstrCatTwoRegOneImm, false},
	147: {"cmov_iz_imm", InstrCatTwoRegOneImm, false},
	148: {"cmov_nz_imm", InstrCatTwoRegOneImm, false},
	149: {"add_imm_64", InstrCatTwoRegOneImm, false},
	150: {"mul_imm_64", InstrCatTwoRegOneImm, false},
	151: {"shlo_l_imm_64", InstrCatTwoRegOneImm, false},
	152: {"shlo_r_imm_64", InstrCatTwoRegOneImm, false},
	153: {"shar_r_imm_64", InstrCatTwoRegOneImm, false},
	154: {"neg_add_imm_64", InstrCatTwoRegOneImm, false},
	155: {"shlo_l_imm_alt_64", InstrCatTwoRegOneImm, false},
	156: {"shlo_r_imm_alt_64", InstrCatTwoRegOneImm, false},
	157: {"shar_r_imm_alt_64", InstrCatTwoRegOneImm, false},
	158: {"rot_r_64_imm", InstrCatTwoRegOneImm, false},
	159: {"rot_r_64_imm_alt", InstrCatTwoRegOneImm, false},
	160: {"rot_r_32_imm", InstrCatTwoRegOneImm, false},
	161: {"rot_r_32_imm_alt", InstrCatTwoRegOneImm, false},

	// A.5.11 Two reg + one offset (terminators)
	170: {"branch_eq", InstrCatTwoRegOneOff, true},
	171: {"branch_ne", InstrCatTwoRegOneOff, true},
	172: {"branch_lt_u", InstrCatTwoRegOneOff, true},
	173: {"branch_lt_s", InstrCatTwoRegOneOff, true},
	174: {"branch_ge_u", InstrCatTwoRegOneOff, true},
	175: {"branch_ge_s", InstrCatTwoRegOneOff, true},

	// A.5.12 Two reg + two imm (terminator)
	180: {"load_imm_jump_ind", InstrCatTwoRegTwoImm, true},

	// A.5.13 Three registers
	190: {"add_32", InstrCatThreeReg, false},
	191: {"sub_32", InstrCatThreeReg, false},
	192: {"mul_32", InstrCatThreeReg, false},
	193: {"div_u_32", InstrCatThreeReg, false},
	194: {"div_s_32", InstrCatThreeReg, false},
	195: {"rem_u_32", InstrCatThreeReg, false},
	196: {"rem_s_32", InstrCatThreeReg, false},
	197: {"shlo_l_32", InstrCatThreeReg, false},
	198: {"shlo_r_32", InstrCatThreeReg, false},
	199: {"shar_r_32", InstrCatThreeReg, false},
	200: {"add_64", InstrCatThreeReg, false},
	201: {"sub_64", InstrCatThreeReg, false},
	202: {"mul_64", InstrCatThreeReg, false},
	203: {"div_u_64", InstrCatThreeReg, false},
	204: {"div_s_64", InstrCatThreeReg, false},
	205: {"rem_u_64", InstrCatThreeReg, false},
	206: {"rem_s_64", InstrCatThreeReg, false},
	207: {"shlo_l_64", InstrCatThreeReg, false},
	208: {"shlo_r_64", InstrCatThreeReg, false},
	209: {"shar_r_64", InstrCatThreeReg, false},
	210: {"and", InstrCatThreeReg, false},
	211: {"xor", InstrCatThreeReg, false},
	212: {"or", InstrCatThreeReg, false},
	213: {"mul_upper_s_s", InstrCatThreeReg, false},
	214: {"mul_upper_u_u", InstrCatThreeReg, false},
	215: {"mul_upper_s_u", InstrCatThreeReg, false},
	216: {"set_lt_u", InstrCatThreeReg, false},
	217: {"set_lt_s", InstrCatThreeReg, false},
	218: {"cmov_iz", InstrCatThreeReg, false},
	219: {"cmov_nz", InstrCatThreeReg, false},
	220: {"rot_l_64", InstrCatThreeReg, false},
	221: {"rot_l_32", InstrCatThreeReg, false},
	222: {"rot_r_64", InstrCatThreeReg, false},
	223: {"rot_r_32", InstrCatThreeReg, false},
	224: {"and_inv", InstrCatThreeReg, false},
	225: {"or_inv", InstrCatThreeReg, false},
	226: {"xnor", InstrCatThreeReg, false},
	227: {"max", InstrCatThreeReg, false},
	228: {"max_u", InstrCatThreeReg, false},
	229: {"min", InstrCatThreeReg, false},
	230: {"min_u", InstrCatThreeReg, false},
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
