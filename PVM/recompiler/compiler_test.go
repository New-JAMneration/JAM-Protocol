//go:build linux && amd64

package recompiler

import (
	"encoding/binary"
	"testing"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"golang.org/x/sys/unix"
)

// ---------------------------------------------------------------------------
// Blob construction helpers
// ---------------------------------------------------------------------------

// buildBlobExact constructs an A.2-format PVM blob from raw instruction bytes
// with explicit instruction boundary positions (byte indices into instBytes).
func buildBlobExact(instBytes []byte, instrBoundaries []int) []byte {
	instSize := len(instBytes)

	bitmaskByteCount := instSize / 8
	if instSize%8 > 0 {
		bitmaskByteCount++
	}
	bitmaskData := make([]byte, bitmaskByteCount)

	for _, idx := range instrBoundaries {
		bitmaskData[idx/8] |= 1 << (idx % 8)
	}

	var blob []byte
	blob = append(blob, encodeUintVariable(0)...) // jumpTableSize = 0
	blob = append(blob, 0)                        // jumpTableLength = 0
	blob = append(blob, encodeUintVariable(uint64(instSize))...)
	blob = append(blob, instBytes...)
	blob = append(blob, bitmaskData...)
	return blob
}

// encodeUintVariable encodes val in PVM E_V format.
func encodeUintVariable(val uint64) []byte {
	if val < 0x80 {
		return []byte{byte(val)}
	}
	buf := make([]byte, 9)
	buf[0] = 0xFF
	for i := range 8 {
		buf[1+i] = byte(val >> (8 * i))
	}
	return buf
}

// packRegs encodes two PVM register indices into one byte: lo=rA, hi=rB<<4.
func packRegs(rA, rB uint8) byte { return (rB << 4) | (rA & 0x0F) }

// imm32LE returns a 4-byte little-endian encoding of a uint32.
func imm32LE(v uint32) []byte {
	return []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)}
}

// ---------------------------------------------------------------------------
// Layer 1: TestCompileAllOpcodes
// ---------------------------------------------------------------------------

func TestCompileAllOpcodes(t *testing.T) {
	for opIdx := range opcodeHandlers {
		handler := opcodeHandlers[opIdx]
		if handler == nil {
			continue
		}

		opByte := byte(opIdx)
		info := PVM.GetOpcodeInfo(opByte)
		name := info.Name
		if name == "" {
			name = "unknown"
		}

		t.Run(name, func(t *testing.T) {
			instBytes, boundaries := buildMinimalInstr(opByte, info)
			blob := buildBlobExact(instBytes, boundaries)

			prog, exitReason := PVM.DeBlobProgramCode(blob)
			if exitReason != PVM.ExitContinue {
				t.Fatalf("DeBlobProgramCode failed: %v", exitReason)
			}

			ctx, err := NewJITContext()
			if err != nil {
				t.Fatalf("NewJITContext: %v", err)
			}
			defer ctx.Close()

			em, err := NewExecutableMemory(0)
			if err != nil {
				t.Fatalf("NewExecutableMemory: %v", err)
			}
			defer em.Close()
			ctx.SetExecutableMemory(em)

			cache := NewCodeCache()
			compiler := NewCompiler(&prog, ctx, cache)
			block, err := compiler.CompileBasicBlock(0)
			if err != nil {
				t.Fatalf("CompileBasicBlock: %v", err)
			}
			if block.NativeSize == 0 {
				t.Fatal("produced zero-length native code")
			}
		})
	}
}

// buildMinimalInstr creates the shortest valid basic block for the given opcode.
// Returns (instBytes, boundaries) ready for buildBlobExact.
func buildMinimalInstr(op byte, info *PVM.OpcodeInfo) ([]byte, []int) {
	var inst []byte
	switch info.Category {
	case PVM.InstrCatNoArg:
		inst = []byte{op}
	case PVM.InstrCatOneImm:
		inst = []byte{op, 0x00}
	case PVM.InstrCatOneRegExtImm:
		inst = []byte{op, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	case PVM.InstrCatTwoImm:
		inst = []byte{op, 0x00}
	case PVM.InstrCatOneOffset:
		// offset=0 means target=pc+0=0 (self-referencing). Safe for compile-only.
		inst = []byte{op, 0x00}
	case PVM.InstrCatOneRegOneImm:
		inst = []byte{op, 0x00}
	case PVM.InstrCatOneRegTwoImm:
		inst = []byte{op, 0x00}
	case PVM.InstrCatOneRegImmOff:
		inst = []byte{op, 0x00}
	case PVM.InstrCatTwoReg:
		inst = []byte{op, 0x00}
	case PVM.InstrCatTwoRegOneImm:
		inst = []byte{op, 0x00}
	case PVM.InstrCatTwoRegOneOff:
		inst = []byte{op, 0x00}
	case PVM.InstrCatTwoRegTwoImm:
		inst = []byte{op, 0x00, 0x00}
	case PVM.InstrCatThreeReg:
		inst = []byte{op, 0x00, 0x00}
	default:
		inst = []byte{op}
	}

	if info.IsTerminator {
		return inst, []int{0}
	}
	// Append trap (opcode 0) as block terminator.
	boundaries := []int{0, len(inst)}
	inst = append(inst, 0x00) // trap
	return inst, boundaries
}

// ---------------------------------------------------------------------------
// Layer 2: TestExecuteInstructions
// ---------------------------------------------------------------------------

type execTestCase struct {
	name     string
	inst     []byte // instruction bytes (without trailing trap)
	initRegs PVM.Registers
	wantRegs PVM.Registers
	wantExit PVM.ExitReason
	setupMem func(t *testing.T, ctx *JITContext)
}

func TestExecuteInstructions(t *testing.T) {
	t.Run("Arithmetic3Reg", testArith3Reg)
	t.Run("Arithmetic2RegImm", testArith2RegImm)
	t.Run("TwoReg", testTwoReg)
	t.Run("BasicOps", testBasicOps)
	t.Run("LoadInd", testLoadInd)
}

// TestSbrkExpandAmountInT0 verifies expand limit check when amount is in T0 (rA=2).
// A RegScratch clobber bug computed 2*heapPointer instead of heapPointer+amount.
func TestSbrkExpandAmountInT0(t *testing.T) {
	const heapLimit = uint64(0x50000)

	instBytes := []byte{101, packRegs(0, 2), 0}
	blob := buildBlobExact(instBytes, []int{0, 2})

	prog, exitReason := PVM.DeBlobProgramCode(blob)
	if exitReason != PVM.ExitContinue {
		t.Fatalf("DeBlobProgramCode: %v", exitReason)
	}

	ctx, err := NewJITContext()
	if err != nil {
		t.Fatalf("NewJITContext: %v", err)
	}
	defer ctx.Close()

	em, err := NewExecutableMemory(0)
	if err != nil {
		t.Fatalf("NewExecutableMemory: %v", err)
	}
	defer em.Close()
	ctx.SetExecutableMemory(em)
	ctx.WriteHeapPointer(0x30000)
	ctx.heapLimit = heapLimit

	cache := NewCodeCache()
	compiler := NewCompiler(&prog, ctx, cache)
	block, err := compiler.CompileBasicBlock(0)
	if err != nil {
		t.Fatalf("CompileBasicBlock: %v", err)
	}

	ctx.WriteRegisters(regsWithValues(2, 100))
	ctx.WriteGas(1000)

	gotExit := ExecuteBlock(ctx, block)
	if gotExit != PVM.ExitHostCall|PVM.ExitReason(SbrkCallID) {
		t.Fatalf("exit reason: got %v, want sbrk expand exit (bug returns inline fail / panic)", gotExit)
	}
}

func TestSbrkExpandExitPC(t *testing.T) {
	instBytes := []byte{101, packRegs(0, 7), 0}
	blob := buildBlobExact(instBytes, []int{0, 2})
	prog, exitReason := PVM.DeBlobProgramCode(blob)
	if exitReason != PVM.ExitContinue {
		t.Fatalf("DeBlobProgramCode: %v", exitReason)
	}

	ctx, err := NewJITContext()
	if err != nil {
		t.Fatalf("NewJITContext: %v", err)
	}
	defer ctx.Close()

	em, err := NewExecutableMemory(0)
	if err != nil {
		t.Fatalf("NewExecutableMemory: %v", err)
	}
	defer em.Close()
	ctx.SetExecutableMemory(em)
	ctx.WriteHeapPointer(0x30000)
	ctx.heapLimit = GuestMemorySize

	cache := NewCodeCache()
	compiler := NewCompiler(&prog, ctx, cache)
	block, err := compiler.CompileBasicBlock(0)
	if err != nil {
		t.Fatalf("CompileBasicBlock: %v", err)
	}

	ctx.WriteRegisters(regsWithValues(7, 4096))
	ctx.WriteGas(1000)

	gotExit := ExecuteBlock(ctx, block)
	if gotExit != PVM.ExitHostCall|PVM.ExitReason(SbrkCallID) {
		t.Fatalf("exit reason: got %v", gotExit)
	}

	wantPC := PVM.ProgramCounter(2) // PC 0 + skip 1 + 1
	if gotPC := ctx.ReadExitPC(); gotPC != wantPC {
		t.Fatalf("ExitPC: got %d, want fallthrough %d", gotPC, wantPC)
	}
}

// TestSbrkExpandResumeSuffix verifies control-flow B: after sbrk expand exits to Go,
// resume compiles only the suffix from fallthroughPC and does not re-run earlier instructions.
func TestSbrkExpandResumeSuffix(t *testing.T) {
	const addImm64 = 149

	instBytes := []byte{
		addImm64, packRegs(1, 1), 1, // PC 0: r1 += 1
		101, packRegs(0, 7), // PC 3: sbrk expand, fallthrough PC 5
		addImm64, packRegs(1, 1), 1, // PC 5: r1 += 1
		0, // PC 8: trap
	}
	boundaries := []int{0, 3, 5, 8}
	blob := buildBlobExact(instBytes, boundaries)

	prog, exitReason := PVM.DeBlobProgramCode(blob)
	if exitReason != PVM.ExitContinue {
		t.Fatalf("DeBlobProgramCode: %v", exitReason)
	}
	if prog.Instrs[1].PC != 3 || fallthroughPC(&prog.Instrs[1]) != 5 {
		t.Fatalf("unexpected sbrk layout: pc=%d fallthrough=%d", prog.Instrs[1].PC, fallthroughPC(&prog.Instrs[1]))
	}

	ctx, err := NewJITContext()
	if err != nil {
		t.Fatalf("NewJITContext: %v", err)
	}
	defer ctx.Close()

	em, err := NewExecutableMemory(0)
	if err != nil {
		t.Fatalf("NewExecutableMemory: %v", err)
	}
	defer em.Close()
	ctx.SetExecutableMemory(em)
	ctx.WriteHeapPointer(0x30000)
	ctx.heapLimit = GuestMemorySize
	ctx.WriteRegisters(regsWithValues(1, 0, 7, 4096))
	ctx.WriteGas(1000)

	recomp := NewRecompiler(&prog, ctx)
	gotExit, _ := recomp.BlockBasedInvoke(0)
	if gotExit != PVM.ExitPanic {
		t.Fatalf("BlockBasedInvoke exit: got %v, want trap/panic", gotExit)
	}

	regs := ctx.ReadRegisters()
	if regs[1] != 2 {
		t.Fatalf("r1=%d want 2 (re-running block head would leave r1=1)", regs[1])
	}
	if regs[0] != 0x31000 {
		t.Fatalf("r0=%#x want 0x31000 (sbrk result)", regs[0])
	}

	suffixBlock, err := recomp.compiler.CompileBasicBlock(5)
	if err != nil {
		t.Fatalf("CompileBasicBlock suffix: %v", err)
	}
	if suffixBlock.PVMStartPC != 5 {
		t.Fatalf("suffix cache key PVMStartPC=%d want 5", suffixBlock.PVMStartPC)
	}
}

func runExecTest(t *testing.T, tc execTestCase) {
	t.Helper()

	instBytes := make([]byte, len(tc.inst))
	copy(instBytes, tc.inst)

	lastOp := instBytes[0]
	boundaries := []int{0}
	if !PVM.IsBlockTerminator(lastOp) {
		boundaries = append(boundaries, len(instBytes))
		instBytes = append(instBytes, 0) // trap
	}

	blob := buildBlobExact(instBytes, boundaries)
	prog, exitReason := PVM.DeBlobProgramCode(blob)
	if exitReason != PVM.ExitContinue {
		t.Fatalf("DeBlobProgramCode: %v", exitReason)
	}

	ctx, err := NewJITContext()
	if err != nil {
		t.Fatalf("NewJITContext: %v", err)
	}
	defer ctx.Close()

	em, err := NewExecutableMemory(0)
	if err != nil {
		t.Fatalf("NewExecutableMemory: %v", err)
	}
	defer em.Close()
	ctx.SetExecutableMemory(em)

	if tc.setupMem != nil {
		tc.setupMem(t, ctx)
		ctx.heapLimit = GuestMemorySize
	}

	cache := NewCodeCache()
	compiler := NewCompiler(&prog, ctx, cache)

	block, err := compiler.CompileBasicBlock(0)
	if err != nil {
		t.Fatalf("CompileBasicBlock: %v", err)
	}

	ctx.WriteRegisters(tc.initRegs)
	ctx.WriteExitReason(0)
	ctx.WriteGas(1000)

	gotExit := ExecuteBlock(ctx, block)
	gotRegs := ctx.ReadRegisters()

	if gotRegs != tc.wantRegs {
		t.Errorf("registers mismatch:\n  got  = %v\n  want = %v", gotRegs, tc.wantRegs)
	}
	if gotExit != tc.wantExit {
		t.Errorf("exit reason: got %v, want %v", gotExit, tc.wantExit)
	}
}

// ---------------------------------------------------------------------------
// Test suites
// ---------------------------------------------------------------------------

func testArith3Reg(t *testing.T) {
	// Three-register encoding: [opcode] [packRegs(rA,rB)] [rD]
	// Semantics: Reg[rD] = Reg[rA] op Reg[rB]
	// To compute r2 = r3 op r4: packRegs(3, 4), rD=2.
	cases := []execTestCase{
		{
			name:     "add_64: r2 = r3 + r4",
			inst:     []byte{200, packRegs(3, 4), 2},
			initRegs: regsWithValues(3, 100, 4, 200),
			wantRegs: regsWithValues(2, 300, 3, 100, 4, 200),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "sub_64: r2 = r3 - r4",
			inst:     []byte{201, packRegs(3, 4), 2},
			initRegs: regsWithValues(3, 500, 4, 200),
			wantRegs: regsWithValues(2, 300, 3, 500, 4, 200),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "mul_64: r2 = r3 * r4",
			inst:     []byte{202, packRegs(3, 4), 2},
			initRegs: regsWithValues(3, 7, 4, 6),
			wantRegs: regsWithValues(2, 42, 3, 7, 4, 6),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "and: r2 = r3 & r4",
			inst:     []byte{210, packRegs(3, 4), 2},
			initRegs: regsWithValues(3, 0xFF00, 4, 0x0FF0),
			wantRegs: regsWithValues(2, 0x0F00, 3, 0xFF00, 4, 0x0FF0),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "xor: r2 = r3 ^ r4",
			inst:     []byte{211, packRegs(3, 4), 2},
			initRegs: regsWithValues(3, 0xFF, 4, 0x0F),
			wantRegs: regsWithValues(2, 0xF0, 3, 0xFF, 4, 0x0F),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "or: r2 = r3 | r4",
			inst:     []byte{212, packRegs(3, 4), 2},
			initRegs: regsWithValues(3, 0xF0, 4, 0x0F),
			wantRegs: regsWithValues(2, 0xFF, 3, 0xF0, 4, 0x0F),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "add_32: r2 = trunc32(r3 + r4)",
			inst:     []byte{190, packRegs(3, 4), 2},
			initRegs: regsWithValues(3, 0xFFFFFFFF, 4, 2),
			wantRegs: regsWithValues(2, 0x00000001, 3, 0xFFFFFFFF, 4, 2),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "div_u_64: r2 = r3 / r4 (unsigned)",
			inst:     []byte{203, packRegs(3, 4), 2},
			initRegs: regsWithValues(3, 100, 4, 3),
			wantRegs: regsWithValues(2, 33, 3, 100, 4, 3),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "div_u_64_by_zero: r2 = r3 / 0 → 2^64-1",
			inst:     []byte{203, packRegs(3, 4), 2},
			initRegs: regsWithValues(3, 100, 4, 0),
			wantRegs: regsWithValues(2, ^uint64(0), 3, 100, 4, 0),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "shlo_l_64: r2 = r3 << r4",
			inst:     []byte{207, packRegs(3, 4), 2},
			initRegs: regsWithValues(3, 1, 4, 10),
			wantRegs: regsWithValues(2, 1024, 3, 1, 4, 10),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "shlo_r_64: r2 = r3 >> r4",
			inst:     []byte{208, packRegs(3, 4), 2},
			initRegs: regsWithValues(3, 1024, 4, 5),
			wantRegs: regsWithValues(2, 32, 3, 1024, 4, 5),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "set_lt_u: r2 = (r3 < r4) ? 1 : 0 (unsigned)",
			inst:     []byte{216, packRegs(3, 4), 2},
			initRegs: regsWithValues(3, 5, 4, 10),
			wantRegs: regsWithValues(2, 1, 3, 5, 4, 10),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "max: r2 = max(r3, r4) (signed)",
			inst:     []byte{227, packRegs(3, 4), 2},
			initRegs: regsWithValues(3, 100, 4, 200),
			wantRegs: regsWithValues(2, 200, 3, 100, 4, 200),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "min: r2 = min(r3, r4) (signed)",
			inst:     []byte{229, packRegs(3, 4), 2},
			initRegs: regsWithValues(3, 100, 4, 200),
			wantRegs: regsWithValues(2, 100, 3, 100, 4, 200),
			wantExit: PVM.ExitPanic,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runExecTest(t, tc)
		})
	}
}

func testArith2RegImm(t *testing.T) {
	// Two-reg + imm encoding: [opcode] [packRegs(rA,rB)] [imm bytes...]
	// Semantics: Reg[rA] = Reg[rB] op sign_extend(imm)
	// skipLen=2 → lX = min(4,max(0,2-1))=1 → 1 imm byte
	cases := []execTestCase{
		{
			name:     "add_imm_32: r0 = trunc32(r1 + 5)",
			inst:     []byte{131, packRegs(0, 1), 5}, // add_imm_32 rA=0, rB=1, imm=5
			initRegs: regsWithValues(1, 10),
			wantRegs: regsWithValues(0, 15, 1, 10),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "add_imm_32: r0 = sext_4(trunc32(r1 + imm))",
			inst:     []byte{131, packRegs(0, 1), 0xFF},
			initRegs: regsWithValues(1, 0),
			wantRegs: regsWithValues(0, 0xFFFFFFFFFFFFFFFF, 1, 0),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "add_imm_64: r0 = r1 + 10",
			inst:     []byte{149, packRegs(0, 1), 10},
			initRegs: regsWithValues(1, 20),
			wantRegs: regsWithValues(0, 30, 1, 20),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "and_imm: r0 = r1 & 0x0F",
			inst:     []byte{132, packRegs(0, 1), 0x0F},
			initRegs: regsWithValues(1, 0xFF),
			wantRegs: regsWithValues(0, 0x0F, 1, 0xFF),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "xor_imm: r0 = r1 ^ 0xFF",
			inst:     []byte{133, packRegs(0, 1), 0xFF},
			initRegs: regsWithValues(1, 0xF0),
			// imm 0xFF sign-extends to 0xFFFFFFFFFFFFFFFF
			wantRegs: regsWithValues(0, 0xF0^0xFFFFFFFFFFFFFFFF, 1, 0xF0),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "or_imm: r0 = r1 | 0x0F",
			inst:     []byte{134, packRegs(0, 1), 0x0F},
			initRegs: regsWithValues(1, 0xF0),
			wantRegs: regsWithValues(0, 0xFF, 1, 0xF0),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "mul_imm_32: r0 = trunc32(r1 * 3)",
			inst:     []byte{135, packRegs(0, 1), 3},
			initRegs: regsWithValues(1, 7),
			wantRegs: regsWithValues(0, 21, 1, 7),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "mul_imm_64: r0 = r1 * 4",
			inst:     []byte{150, packRegs(0, 1), 4},
			initRegs: regsWithValues(1, 8),
			wantRegs: regsWithValues(0, 32, 1, 8),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "shlo_l_imm_64: r0 = r1 << 3",
			inst:     []byte{151, packRegs(0, 1), 3},
			initRegs: regsWithValues(1, 1),
			wantRegs: regsWithValues(0, 8, 1, 1),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "shlo_r_imm_64: r0 = r1 >> 2",
			inst:     []byte{152, packRegs(0, 1), 2},
			initRegs: regsWithValues(1, 100),
			wantRegs: regsWithValues(0, 25, 1, 100),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "neg_add_imm_64: r0 = imm - r1 = 10 - 3 = 7",
			inst:     []byte{154, packRegs(0, 1), 10},
			initRegs: regsWithValues(1, 3),
			wantRegs: regsWithValues(0, 7, 1, 3),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "set_lt_u_imm: r0 = (r1 <u 4) ? 1 : 0",
			inst:     []byte{136, packRegs(0, 1), 4},
			initRegs: regsWithValues(1, 3),
			wantRegs: regsWithValues(0, 1, 1, 3),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "set_lt_u_imm: r0 = (r1 <u 3) ? 1 : 0 (equal)",
			inst:     []byte{136, packRegs(0, 1), 3},
			initRegs: regsWithValues(1, 3),
			wantRegs: regsWithValues(0, 0, 1, 3),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "set_lt_u_imm same reg: r0 = (r0 <u 4) ? 1 : 0",
			inst:     []byte{136, packRegs(0, 0), 4},
			initRegs: regsWithValues(0, 3),
			wantRegs: regsWithValues(0, 1),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "set_lt_u_imm 2-byte imm: r0 = (r1 <u 0x100) ? 1 : 0",
			inst:     []byte{136, packRegs(0, 1), 0x00, 0x01},
			initRegs: regsWithValues(1, 3),
			wantRegs: regsWithValues(0, 1, 1, 3),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "set_lt_u_imm ext reg: r5 = (r6 <u 4) ? 1 : 0",
			inst:     []byte{136, packRegs(5, 6), 4},
			initRegs: regsWithValues(6, 3),
			wantRegs: regsWithValues(5, 1, 6, 3),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "set_lt_u_imm ext reg: r9 = (r10 <u 4) ? 1 : 0",
			inst:     []byte{136, packRegs(9, 10), 4},
			initRegs: regsWithValues(10, 3),
			wantRegs: regsWithValues(9, 1, 10, 3),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "fuzz 1766241814 step 11498: r12 = (r7 <u 10) ? 1 : 0",
			inst:     []byte{136, packRegs(12, 7), 0x0a},
			initRegs: regsWithValues(7, 3),
			wantRegs: regsWithValues(12, 1, 7, 3),
			wantExit: PVM.ExitPanic,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runExecTest(t, tc)
		})
	}
}

func testTwoReg(t *testing.T) {
	// Two-register encoding: [opcode] [packRegs(rD, rA)]
	// Semantics: Reg[rD] = f(Reg[rA])
	cases := []execTestCase{
		{
			name:     "move_reg: r0 = r1",
			inst:     []byte{100, packRegs(0, 1)},
			initRegs: regsWithValues(1, 42),
			wantRegs: regsWithValues(0, 42, 1, 42),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "sign_extend_8: r0 = sext8(r1)",
			inst:     []byte{108, packRegs(0, 1)},
			initRegs: regsWithValues(1, 0x80), // -128 in i8
			wantRegs: regsWithValues(0, 0xFFFFFFFFFFFFFF80, 1, 0x80),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "sign_extend_16: r0 = sext16(r1)",
			inst:     []byte{109, packRegs(0, 1)},
			initRegs: regsWithValues(1, 0x8000), // -32768 in i16
			wantRegs: regsWithValues(0, 0xFFFFFFFFFFFF8000, 1, 0x8000),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "zero_extend_16: r0 = zext16(r1)",
			inst:     []byte{110, packRegs(0, 1)},
			initRegs: regsWithValues(1, 0xDEADBEEF12340000|0xABCD),
			wantRegs: regsWithValues(0, 0xABCD, 1, 0xDEADBEEF12340000|0xABCD),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "reverse_bytes: r0 = bswap64(r1)",
			inst:     []byte{111, packRegs(0, 1)},
			initRegs: regsWithValues(1, 0x0102030405060708),
			wantRegs: regsWithValues(0, 0x0807060504030201, 1, 0x0102030405060708),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "count_set_bits_64: r0 = popcnt64(r1)",
			inst:     []byte{102, packRegs(0, 1)},
			initRegs: regsWithValues(1, 0xFF),
			wantRegs: regsWithValues(0, 8, 1, 0xFF),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "leading_zero_bits_64: r0 = lzcnt64(r1)",
			inst:     []byte{104, packRegs(0, 1)},
			initRegs: regsWithValues(1, 1), // 63 leading zeros
			wantRegs: regsWithValues(0, 63, 1, 1),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "trailing_zero_bits_64: r0 = tzcnt64(r1)",
			inst:     []byte{106, packRegs(0, 1)},
			initRegs: regsWithValues(1, 0x100), // 8 trailing zeros
			wantRegs: regsWithValues(0, 8, 1, 0x100),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "sbrk query: r0 = heap pointer when r7==0",
			inst:     []byte{101, packRegs(0, 7)},
			initRegs: regsWithValues(7, 0),
			wantRegs: regsWithValues(0, 0x33000, 7, 0),
			wantExit: PVM.ExitPanic,
			setupMem: func(_ *testing.T, ctx *JITContext) {
				ctx.WriteHeapPointer(0x33000)
			},
		},
		{
			name:     "sbrk expand: exits to Go when r7!=0",
			inst:     []byte{101, packRegs(0, 7)},
			initRegs: regsWithValues(7, 4096),
			wantRegs: regsWithValues(0, 0, 7, 4096),
			wantExit: PVM.ExitHostCall | PVM.ExitReason(SbrkCallID),
			setupMem: func(_ *testing.T, ctx *JITContext) {
				ctx.WriteHeapPointer(0x30000)
				ctx.heapLimit = GuestMemorySize
			},
		},
		{
			name:     "sbrk overflow: r0=0 inline when newHP wraps",
			inst:     []byte{101, packRegs(0, 7)},
			initRegs: regsWithValues(7, 1),
			wantRegs: regsWithValues(0, 0, 7, 1),
			wantExit: PVM.ExitPanic,
			setupMem: func(_ *testing.T, ctx *JITContext) {
				ctx.WriteHeapPointer(^uint64(0))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runExecTest(t, tc)
		})
	}
}

func testLoadInd(t *testing.T) {
	const (
		memAddr = uint32(0x20000)
		memVal  = uint64(0x328c0)
		vX      = uint32(0x600)
	)

	cases := []execTestCase{
		{
			name: "load_ind_u64 into T0 (reg 2)",
			inst: append([]byte{130, packRegs(2, 7)}, imm32LE(vX)...),
			setupMem: func(t *testing.T, ctx *JITContext) {
				t.Helper()
				end := memAddr + PVM.ZP
				if err := ctx.mapSegment(memAddr, end, nil, unix.PROT_READ|unix.PROT_WRITE); err != nil {
					t.Fatalf("mapSegment: %v", err)
				}
				binary.LittleEndian.PutUint64(ctx.guestMem[memAddr:], memVal)
			},
			initRegs: regsWithValues(2, 0, 7, uint64(memAddr-vX)),
			wantRegs: regsWithValues(2, memVal, 7, uint64(memAddr-vX)),
			wantExit: PVM.ExitPanic,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runExecTest(t, tc)
		})
	}
}

func testBasicOps(t *testing.T) {
	cases := []execTestCase{
		{
			name:     "trap: causes ExitPanic",
			inst:     []byte{0}, // trap (opcode 0) — self-terminating
			initRegs: PVM.Registers{},
			wantRegs: PVM.Registers{},
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "fallthrough: NOP then exit",
			inst:     []byte{1}, // fallthrough emits NOP; exit trampoline returns ExitContinue
			initRegs: PVM.Registers{},
			wantRegs: PVM.Registers{},
			wantExit: PVM.ExitContinue,
		},
		{
			name:     "load_imm: r0 = 42",
			inst:     []byte{51, 0x00, 42}, // load_imm rA=0, imm=42
			initRegs: PVM.Registers{},
			wantRegs: regsWithValues(0, 42),
			wantExit: PVM.ExitPanic,
		},
		{
			name:     "load_imm_64: r0 = 0x12345678",
			inst:     []byte{20, 0x00, 0x78, 0x56, 0x34, 0x12, 0x00, 0x00, 0x00, 0x00}, // load_imm_64 rA=0, imm=0x12345678
			initRegs: PVM.Registers{},
			wantRegs: regsWithValues(0, 0x12345678),
			wantExit: PVM.ExitPanic,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runExecTest(t, tc)
		})
	}
}

// ---------------------------------------------------------------------------
// Register builder helper
// ---------------------------------------------------------------------------

// regsWithValues builds a Registers array from (index, value) pairs.
// Any register not specified stays zero.
func regsWithValues(kvs ...uint64) PVM.Registers {
	var regs PVM.Registers
	for i := 0; i+1 < len(kvs); i += 2 {
		idx := kvs[i]
		val := kvs[i+1]
		regs[idx] = val
	}
	return regs
}
