//go:build linux && amd64

package recompiler

import (
	"fmt"
	"time"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler/asm"
)

// opcodeHandler is the unified signature for all opcode emit functions.
type opcodeHandler func(c *Compiler, a *asm.Assembler, instr *PVM.InstrMeta) error

// opcodeHandlers is the dispatch table indexed by PVM opcode (0–230).
// nil entries mean the opcode is unimplemented; compileInstruction returns an error.
var opcodeHandlers [231]opcodeHandler

func init() {
	// 4.3 No-argument
	opcodeHandlers[0] = (*Compiler).emitTrap
	opcodeHandlers[1] = (*Compiler).emitFallthrough

	// 4.4 Immediate load
	opcodeHandlers[10] = (*Compiler).emitEcalli
	opcodeHandlers[20] = (*Compiler).emitLoadImm64
	opcodeHandlers[51] = (*Compiler).emitLoadImm

	// 4.5 Memory store (two immediates)
	opcodeHandlers[30] = makeStoreImm(1)
	opcodeHandlers[31] = makeStoreImm(2)
	opcodeHandlers[32] = makeStoreImm(4)
	opcodeHandlers[33] = makeStoreImm(8)

	// 4.5 One offset (jump)
	opcodeHandlers[40] = (*Compiler).emitJump

	// 4.5 One reg + one imm
	opcodeHandlers[50] = (*Compiler).emitJumpInd
	opcodeHandlers[52] = makeLoad(1, false)
	opcodeHandlers[53] = makeLoad(1, true)
	opcodeHandlers[54] = makeLoad(2, false)
	opcodeHandlers[55] = makeLoad(2, true)
	opcodeHandlers[56] = makeLoad(4, false)
	opcodeHandlers[57] = makeLoad(4, true)
	opcodeHandlers[58] = makeLoad(8, false)
	opcodeHandlers[59] = makeStore(1)
	opcodeHandlers[60] = makeStore(2)
	opcodeHandlers[61] = makeStore(4)
	opcodeHandlers[62] = makeStore(8)

	// 4.5 One reg + two imm (store_imm_ind)
	opcodeHandlers[70] = makeStoreImmInd(1)
	opcodeHandlers[71] = makeStoreImmInd(2)
	opcodeHandlers[72] = makeStoreImmInd(4)
	opcodeHandlers[73] = makeStoreImmInd(8)

	// 4.9 Branch (one reg + imm + offset)
	opcodeHandlers[80] = (*Compiler).emitLoadImmJump
	opcodeHandlers[81] = makeBranchImm(asm.CondEQ)
	opcodeHandlers[82] = makeBranchImm(asm.CondNE)
	opcodeHandlers[83] = makeBranchImm(asm.CondB)
	opcodeHandlers[84] = makeBranchImm(asm.CondBE)
	opcodeHandlers[85] = makeBranchImm(asm.CondAE)
	opcodeHandlers[86] = makeBranchImm(asm.CondA)
	opcodeHandlers[87] = makeBranchImm(asm.CondLT)
	opcodeHandlers[88] = makeBranchImm(asm.CondLE)
	opcodeHandlers[89] = makeBranchImm(asm.CondGE)
	opcodeHandlers[90] = makeBranchImm(asm.CondGT)

	// 4.8 Two-register
	opcodeHandlers[100] = (*Compiler).emitMoveReg
	opcodeHandlers[101] = (*Compiler).emitSbrk
	opcodeHandlers[102] = (*Compiler).emitCountSetBits64
	opcodeHandlers[103] = (*Compiler).emitCountSetBits32
	opcodeHandlers[104] = (*Compiler).emitLeadingZeroBits64
	opcodeHandlers[105] = (*Compiler).emitLeadingZeroBits32
	opcodeHandlers[106] = (*Compiler).emitTrailingZeroBits64
	opcodeHandlers[107] = (*Compiler).emitTrailingZeroBits32
	opcodeHandlers[108] = (*Compiler).emitSignExtend8
	opcodeHandlers[109] = (*Compiler).emitSignExtend16
	opcodeHandlers[110] = (*Compiler).emitZeroExtend16
	opcodeHandlers[111] = (*Compiler).emitReverseBytes

	// 4.5 Two reg + one imm (store_ind, load_ind)
	opcodeHandlers[120] = makeStoreInd(1)
	opcodeHandlers[121] = makeStoreInd(2)
	opcodeHandlers[122] = makeStoreInd(4)
	opcodeHandlers[123] = makeStoreInd(8)
	opcodeHandlers[124] = makeLoadInd(1, false)
	opcodeHandlers[125] = makeLoadInd(1, true)
	opcodeHandlers[126] = makeLoadInd(2, false)
	opcodeHandlers[127] = makeLoadInd(2, true)
	opcodeHandlers[128] = makeLoadInd(4, false)
	opcodeHandlers[129] = makeLoadInd(4, true)
	opcodeHandlers[130] = makeLoadInd(8, false)

	// 4.6 Two reg + one imm (arithmetic)
	opcodeHandlers[131] = (*Compiler).emitAddImm32
	opcodeHandlers[132] = (*Compiler).emitAndImm
	opcodeHandlers[133] = (*Compiler).emitXorImm
	opcodeHandlers[134] = (*Compiler).emitOrImm
	opcodeHandlers[135] = (*Compiler).emitMulImm32
	opcodeHandlers[136] = (*Compiler).emitSetLtUImm
	opcodeHandlers[137] = (*Compiler).emitSetLtSImm
	opcodeHandlers[138] = (*Compiler).emitShloLImm32
	opcodeHandlers[139] = (*Compiler).emitShloRImm32
	opcodeHandlers[140] = (*Compiler).emitSharRImm32
	opcodeHandlers[141] = (*Compiler).emitNegAddImm32
	opcodeHandlers[142] = (*Compiler).emitSetGtUImm
	opcodeHandlers[143] = (*Compiler).emitSetGtSImm
	opcodeHandlers[144] = (*Compiler).emitShloLImmAlt32
	opcodeHandlers[145] = (*Compiler).emitShloRImmAlt32
	opcodeHandlers[146] = (*Compiler).emitSharRImmAlt32
	opcodeHandlers[147] = (*Compiler).emitCmovIzImm
	opcodeHandlers[148] = (*Compiler).emitCmovNzImm
	opcodeHandlers[149] = (*Compiler).emitAddImm64
	opcodeHandlers[150] = (*Compiler).emitMulImm64
	opcodeHandlers[151] = (*Compiler).emitShloLImm64
	opcodeHandlers[152] = (*Compiler).emitShloRImm64
	opcodeHandlers[153] = (*Compiler).emitSharRImm64
	opcodeHandlers[154] = (*Compiler).emitNegAddImm64
	opcodeHandlers[155] = (*Compiler).emitShloLImmAlt64
	opcodeHandlers[156] = (*Compiler).emitShloRImmAlt64
	opcodeHandlers[157] = (*Compiler).emitSharRImmAlt64
	opcodeHandlers[158] = (*Compiler).emitRotRImm64
	opcodeHandlers[159] = (*Compiler).emitRotRImmAlt64
	opcodeHandlers[160] = (*Compiler).emitRotRImm32
	opcodeHandlers[161] = (*Compiler).emitRotRImmAlt32

	// 4.9 Two reg + one offset (branch)
	opcodeHandlers[170] = makeBranch(asm.CondEQ)
	opcodeHandlers[171] = makeBranch(asm.CondNE)
	opcodeHandlers[172] = makeBranch(asm.CondB)
	opcodeHandlers[173] = makeBranch(asm.CondLT)
	opcodeHandlers[174] = makeBranch(asm.CondAE)
	opcodeHandlers[175] = makeBranch(asm.CondGE)

	// 4.9 Two reg + two imm
	opcodeHandlers[180] = (*Compiler).emitLoadImmJumpInd

	// 4.7 Three-register (32-bit)
	opcodeHandlers[190] = (*Compiler).emitAdd32
	opcodeHandlers[191] = (*Compiler).emitSub32
	opcodeHandlers[192] = (*Compiler).emitMul32
	opcodeHandlers[193] = (*Compiler).emitDivU32
	opcodeHandlers[194] = (*Compiler).emitDivS32
	opcodeHandlers[195] = (*Compiler).emitRemU32
	opcodeHandlers[196] = (*Compiler).emitRemS32
	opcodeHandlers[197] = (*Compiler).emitShloL32
	opcodeHandlers[198] = (*Compiler).emitShloR32
	opcodeHandlers[199] = (*Compiler).emitSharR32

	// 4.7 Three-register (64-bit)
	opcodeHandlers[200] = (*Compiler).emitAdd64
	opcodeHandlers[201] = (*Compiler).emitSub64
	opcodeHandlers[202] = (*Compiler).emitMul64
	opcodeHandlers[203] = (*Compiler).emitDivU64
	opcodeHandlers[204] = (*Compiler).emitDivS64
	opcodeHandlers[205] = (*Compiler).emitRemU64
	opcodeHandlers[206] = (*Compiler).emitRemS64
	opcodeHandlers[207] = (*Compiler).emitShloL64
	opcodeHandlers[208] = (*Compiler).emitShloR64
	opcodeHandlers[209] = (*Compiler).emitSharR64

	// 4.7 Bitwise / mul-upper / cmp / cmov / rotate
	opcodeHandlers[210] = (*Compiler).emitAnd
	opcodeHandlers[211] = (*Compiler).emitXor
	opcodeHandlers[212] = (*Compiler).emitOr
	opcodeHandlers[213] = (*Compiler).emitMulUpperSS
	opcodeHandlers[214] = (*Compiler).emitMulUpperUU
	opcodeHandlers[215] = (*Compiler).emitMulUpperSU
	opcodeHandlers[216] = (*Compiler).emitSetLtU
	opcodeHandlers[217] = (*Compiler).emitSetLtS
	opcodeHandlers[218] = (*Compiler).emitCmovIz
	opcodeHandlers[219] = (*Compiler).emitCmovNz
	opcodeHandlers[220] = (*Compiler).emitRotL64
	opcodeHandlers[221] = (*Compiler).emitRotL32
	opcodeHandlers[222] = (*Compiler).emitRotR64
	opcodeHandlers[223] = (*Compiler).emitRotR32
	opcodeHandlers[224] = (*Compiler).emitAndInv
	opcodeHandlers[225] = (*Compiler).emitOrInv
	opcodeHandlers[226] = (*Compiler).emitXnor
	opcodeHandlers[227] = (*Compiler).emitMax
	opcodeHandlers[228] = (*Compiler).emitMaxU
	opcodeHandlers[229] = (*Compiler).emitMin
	opcodeHandlers[230] = (*Compiler).emitMinU
}

// Closure factories for handlers that need extra parameters (size, signed, condCode)
// baked in at init time — zero allocations at dispatch time.

func makeStoreImm(size int) opcodeHandler {
	return func(c *Compiler, a *asm.Assembler, instr *PVM.InstrMeta) error {
		return c.emitStoreImm(a, instr, size)
	}
}

func makeStore(size int) opcodeHandler {
	return func(c *Compiler, a *asm.Assembler, instr *PVM.InstrMeta) error {
		return c.emitStore(a, instr, size)
	}
}

func makeStoreImmInd(size int) opcodeHandler {
	return func(c *Compiler, a *asm.Assembler, instr *PVM.InstrMeta) error {
		return c.emitStoreImmInd(a, instr, size)
	}
}

func makeStoreInd(size int) opcodeHandler {
	return func(c *Compiler, a *asm.Assembler, instr *PVM.InstrMeta) error {
		return c.emitStoreInd(a, instr, size)
	}
}

func makeLoad(size int, signed bool) opcodeHandler {
	return func(c *Compiler, a *asm.Assembler, instr *PVM.InstrMeta) error {
		return c.emitLoad(a, instr, size, signed)
	}
}

func makeLoadInd(size int, signed bool) opcodeHandler {
	return func(c *Compiler, a *asm.Assembler, instr *PVM.InstrMeta) error {
		return c.emitLoadInd(a, instr, size, signed)
	}
}

func makeBranchImm(cc asm.ConditionCode) opcodeHandler {
	return func(c *Compiler, a *asm.Assembler, instr *PVM.InstrMeta) error {
		return c.emitBranchImm(a, instr, cc)
	}
}

func makeBranch(cc asm.ConditionCode) opcodeHandler {
	return func(c *Compiler, a *asm.Assembler, instr *PVM.InstrMeta) error {
		return c.emitBranch(a, instr, cc)
	}
}

type Compiler struct {
	asm     *asm.Assembler
	program *PVM.Program
	cache   *CodeCache
	ctx     *JITContext
	linking map[PVM.ProgramCounter]bool // in-progress compileForLink (cycle guard)
	djump   *djumpSupport               // jump-table rodata + PC→native dispatch (lazy init)

	// Per-block static link targets (block linking, strategy a). Set just before
	// the instruction loop and read by the terminator emit handlers (emitJump /
	// emitBranch* via emitLinkOrExit). nil ⇒ fall back to the Go dispatcher.
	// compileForLink recurses and clobbers these, so they are (re)assigned for
	// THIS block only after all pre-compilation — see compileBasicBlockAtDepth.
	linkFallthrough *CompiledBlock // not-taken / sequential successor
	linkTaken       *CompiledBlock // static jump / branch-taken target
}

func NewCompiler(program *PVM.Program, ctx *JITContext, cache *CodeCache) *Compiler {
	if cache.em == nil && ctx.executableMem != nil {
		cache.BindExecutableMemory(ctx.executableMem)
	}
	return &Compiler{
		asm:     asm.NewAssembler(),
		program: program,
		cache:   cache,
		ctx:     ctx,
	}
}

func (c *Compiler) CompileBasicBlock(startPC PVM.ProgramCounter) (*CompiledBlock, error) {
	return c.compileBasicBlockAtDepth(startPC, 0)
}

func (c *Compiler) compileBasicBlockAtDepth(startPC PVM.ProgramCounter, linkDepth int) (*CompiledBlock, error) {
	if jitProfile {
		jm.compileCalls.Add(1)
		// Time only the top-level entry; recursive eager-link sub-compiles are
		// already inside this elapsed window, so timing them again double-counts.
		if linkDepth == 0 {
			t := time.Now()
			defer func() { jm.compileNanos.Add(int64(time.Since(t))) }()
		}
	}
	if c.linking == nil {
		c.linking = make(map[PVM.ProgramCounter]bool)
	}
	if c.linking[startPC] {
		return nil, fmt.Errorf("compile cycle at PC=%d", startPC)
	}
	c.linking[startPC] = true
	defer delete(c.linking, startPC)

	jt := c.program.JumpTable
	if jt.Length > 0 && jt.Size > 0 && len(jt.Data) > 0 {
		if err := c.ensureDjumpSupport(); err != nil {
			return nil, err
		}
	}

	blockMeta := c.program.BlockContaining(startPC)
	if blockMeta == nil {
		idx := c.program.InstrIdxAt[startPC]
		if idx < 0 {
			return nil, fmt.Errorf("no pre-decoded block at PC=%d", startPC)
		}
		startIdx := int(idx)
		endIdx := startIdx
		for endIdx < len(c.program.Instrs) {
			if PVM.IsBlockTerminator(c.program.Instrs[endIdx].Opcode) {
				endIdx++
				break
			}
			endIdx++
		}
		if endIdx <= startIdx {
			return nil, fmt.Errorf("empty instruction tail at PC=%d", startPC)
		}
		blockMeta = &PVM.BlockMeta{
			StartPC:    startPC,
			EndPC:      c.program.Instrs[endIdx-1].PC,
			InstrStart: startIdx,
			InstrEnd:   endIdx,
			GasCost:    PVM.Gas(endIdx - startIdx),
		}
	} else if startPC != blockMeta.StartPC {
		// Resume inside a decoded block (e.g. after sbrk). Compile the suffix
		// from startPC only — executing from the block head would re-run earlier
		// instructions and can loop on sbrk until the process SIGSEGVs.
		idx := c.program.InstrIdxAt[startPC]
		if idx < 0 {
			return nil, fmt.Errorf("no instruction at PC=%d", startPC)
		}
		startIdx := int(idx)
		if startIdx < blockMeta.InstrStart || startIdx >= blockMeta.InstrEnd {
			return nil, fmt.Errorf("PC=%d outside block [%d,%d)", startPC, blockMeta.InstrStart, blockMeta.InstrEnd)
		}
		blockMeta = &PVM.BlockMeta{
			StartPC:    startPC,
			EndPC:      c.program.Instrs[blockMeta.InstrEnd-1].PC,
			InstrStart: startIdx,
			InstrEnd:   blockMeta.InstrEnd,
			GasCost:    PVM.Gas(blockMeta.InstrEnd - startIdx),
		}
	}

	instrs := c.program.Instrs[blockMeta.InstrStart:blockMeta.InstrEnd]
	lastInstr := &instrs[len(instrs)-1]
	fallthroughPC := lastInstr.PC + PVM.ProgramCounter(lastInstr.SkipLen) + 1

	// Pre-compile static link targets before touching the shared Assembler
	// (strategy a): the not-taken/sequential fallthrough, and the terminator's
	// static jump/branch-taken target. compileForLink may recurse and clobber
	// c.link*, so c.link* are assigned for THIS block only after all
	// pre-compilation, immediately before the loop. Back-edges (target still in
	// c.linking) return an error → nil link → safe emitExitToPC fallback.
	var linkFallthrough, linkTaken *CompiledBlock
	if c.program.Bitmasks.IsStartOfBasicBlock(fallthroughPC) {
		linkFallthrough, _ = c.compileForLink(fallthroughPC, linkDepth+1)
	}
	if tgt, ok := staticBranchTarget(lastInstr); ok && c.program.Bitmasks.IsStartOfBasicBlock(tgt) {
		linkTaken, _ = c.compileForLink(tgt, linkDepth+1)
	}
	c.linkFallthrough = linkFallthrough
	c.linkTaken = linkTaken

	a := c.asm
	a.Reset()

	// blockBased gas charging (0.8.0 uncommented this):
	// c.emitBlockGasCheck(a, blockMeta.StartPC, int64(blockMeta.GasCost))

	for i := range instrs {
		instr := &instrs[i]

		// per-instruction gas charging (GP v0.7.2, remove this in 0.8.0):
		c.emitGasCheck(a, instr.PC)

		handler := opcodeHandlers[instr.Opcode]
		if handler == nil {
			return nil, fmt.Errorf("unsupported opcode %d at PC=%d", instr.Opcode, instr.PC)
		}
		if err := handler(c, a, instr); err != nil {
			return nil, fmt.Errorf("failed to compile instruction at PC=%d: %w", instr.PC, err)
		}
	}

	// blockBased gas charging (0.8.0 uncommented this):
	// emitBlockOutOfGasExit(a, blockMeta.StartPC)

	blockEpilogue := fmt.Sprintf("block_epilogue_%d", blockMeta.StartPC)
	a.Jmp(blockEpilogue)

	// per-instruction gas charging (GP v0.7.2, remove this in 0.8.0):
	for i := range instrs {
		emitOutOfGasExit(a, instrs[i].PC)
	}

	_ = a.BindLabel(blockEpilogue)
	emitFallthroughEpilogue(a, fallthroughPC, linkFallthrough)
	EmitExitTrampoline(a)

	code, err := a.Finalize()
	if err != nil {
		return nil, fmt.Errorf("finalize assembler: %w", err)
	}

	if c.ctx.executableMem == nil {
		return nil, fmt.Errorf("executable memory not initialized")
	}

	em := c.ctx.executableMem
	offset, err := em.Write(code)
	if err != nil {
		return nil, fmt.Errorf("write native code: %w", err)
	}

	block := &CompiledBlock{
		PVMStartPC:   startPC,
		PVMEndPC:     fallthroughPC,
		NativeAddr:   em.GetPtr(offset),
		NativeOffset: offset,
		NativeSize:   len(code),
		GasCost:      int64(blockMeta.GasCost),
		InstrCount:   blockMeta.InstrCount(),
	}
	c.cache.Put(block)
	c.registerDispatch(block)
	return block, nil
}
