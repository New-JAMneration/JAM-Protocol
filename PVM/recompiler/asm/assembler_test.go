package asm

import (
	"bytes"
	"fmt"
	"testing"
)

// expectBytes is a test helper that compares emitted bytes to expected.
func expectBytes(t *testing.T, name string, got, want []byte) {
	t.Helper()
	if !bytes.Equal(got, want) {
		t.Errorf("%s:\n  got  %s\n  want %s", name, hexDump(got), hexDump(want))
	}
}

func hexDump(b []byte) string {
	if len(b) == 0 {
		return "[]"
	}
	s := "["
	for i, v := range b {
		if i > 0 {
			s += " "
		}
		s += fmt.Sprintf("%02X", v)
	}
	return s + "]"
}

// emit runs fn on a fresh assembler and returns the emitted bytes.
func emit(fn func(a *Assembler)) []byte {
	a := NewAssembler()
	fn(a)
	return a.Buffer().Bytes()
}

// ---------------------------------------------------------------------------
// Register helpers
// ---------------------------------------------------------------------------

func TestRegister_IsExtended(t *testing.T) {
	for _, r := range []Register{RAX, RCX, RDX, RBX, RSP, RBP, RSI, RDI} {
		if r.IsExtended() {
			t.Errorf("%s should not be extended", r)
		}
	}
	for _, r := range []Register{R8, R9, R10, R11, R12, R13, R14, R15} {
		if !r.IsExtended() {
			t.Errorf("%s should be extended", r)
		}
	}
}

func TestRegister_Lo3(t *testing.T) {
	if RAX.Lo3() != 0 {
		t.Error("RAX.Lo3() != 0")
	}
	if R8.Lo3() != 0 {
		t.Error("R8.Lo3() != 0")
	}
	if RDI.Lo3() != 7 {
		t.Error("RDI.Lo3() != 7")
	}
	if R15.Lo3() != 7 {
		t.Error("R15.Lo3() != 7")
	}
}

func TestRegister_IsCalleeSaved(t *testing.T) {
	calleeSaved := map[Register]bool{RBX: true, RBP: true, R12: true, R13: true, R14: true, R15: true}
	for r := Register(0); r <= R15; r++ {
		got := r.IsCalleeSaved()
		want := calleeSaved[r]
		if got != want {
			t.Errorf("%s.IsCalleeSaved() = %v, want %v", r, got, want)
		}
	}
}

func TestRegister_String(t *testing.T) {
	if RAX.String() != "rax" {
		t.Error("RAX.String() != rax")
	}
	if R15.String() != "r15" {
		t.Error("R15.String() != r15")
	}
}

// ---------------------------------------------------------------------------
// CodeBuffer basics
// ---------------------------------------------------------------------------

func TestCodeBuffer_Emit(t *testing.T) {
	buf := NewCodeBuffer()
	buf.Emit(0x90, 0xC3)
	if buf.Len() != 2 {
		t.Fatalf("Len = %d, want 2", buf.Len())
	}
	expectBytes(t, "Emit", buf.Bytes(), []byte{0x90, 0xC3})
}

func TestCodeBuffer_EmitUint32LE(t *testing.T) {
	buf := NewCodeBuffer()
	buf.EmitUint32LE(0xDEADBEEF)
	expectBytes(t, "EmitUint32LE", buf.Bytes(), []byte{0xEF, 0xBE, 0xAD, 0xDE})
}

func TestCodeBuffer_EmitUint64LE(t *testing.T) {
	buf := NewCodeBuffer()
	buf.EmitUint64LE(0x0102030405060708)
	expectBytes(t, "EmitUint64LE", buf.Bytes(), []byte{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01})
}

func TestCodeBuffer_Reset(t *testing.T) {
	buf := NewCodeBuffer()
	buf.Emit(0x90, 0x90)
	buf.Reset()
	if buf.Len() != 0 {
		t.Fatalf("Len after Reset = %d", buf.Len())
	}
}

// ---------------------------------------------------------------------------
// Label system
// ---------------------------------------------------------------------------

func TestLabel_BackwardReference(t *testing.T) {
	a := NewAssembler()
	_ = a.BindLabel("loop")
	a.Nop()
	a.Jmp("loop")
	code, err := a.Finalize()
	if err != nil {
		t.Fatal(err)
	}
	// NOP = 1 byte at offset 0
	// JMP rel32 = E9 xx xx xx xx at offset 1, 5 bytes
	// target=0, fixup_offset=2 (after E9), size=4
	// rel = 0 - (2+4) = -6
	expectBytes(t, "backward jmp", code, []byte{
		0x90,                         // NOP
		0xE9, 0xFA, 0xFF, 0xFF, 0xFF, // JMP rel32(-6)
	})
}

func TestLabel_ForwardReference(t *testing.T) {
	a := NewAssembler()
	a.Jmp("end")
	a.Nop()
	a.Nop()
	_ = a.BindLabel("end")
	a.Ret()
	code, err := a.Finalize()
	if err != nil {
		t.Fatal(err)
	}
	// JMP rel32 at offset 0: E9 xx xx xx xx (5 bytes)
	// NOP at offset 5
	// NOP at offset 6
	// label "end" at offset 7
	// rel = 7 - (1+4) = 2
	expectBytes(t, "forward jmp", code, []byte{
		0xE9, 0x02, 0x00, 0x00, 0x00, // JMP +2
		0x90, // NOP
		0x90, // NOP
		0xC3, // RET
	})
}

func TestLabel_MultipleReferences(t *testing.T) {
	a := NewAssembler()
	a.Jmp("target")
	a.Nop()
	a.Jmp("target")
	_ = a.BindLabel("target")
	a.Ret()
	code, err := a.Finalize()
	if err != nil {
		t.Fatal(err)
	}
	// JMP at 0: E9 xx xx xx xx → target=11, rel=11-(1+4)=6
	// NOP at 5
	// JMP at 6: E9 xx xx xx xx → target=11, rel=11-(7+4)=0
	// target at 11: RET
	expectBytes(t, "multi-ref", code, []byte{
		0xE9, 0x06, 0x00, 0x00, 0x00, // JMP +6
		0x90,                         // NOP
		0xE9, 0x00, 0x00, 0x00, 0x00, // JMP +0
		0xC3, // RET
	})
}

func TestLabel_Unresolved(t *testing.T) {
	a := NewAssembler()
	a.Jmp("nowhere")
	_, err := a.Finalize()
	if err == nil {
		t.Fatal("expected error for unresolved label")
	}
}

func TestLabel_Duplicate(t *testing.T) {
	a := NewAssembler()
	if err := a.BindLabel("x"); err != nil {
		t.Fatal(err)
	}
	if err := a.BindLabel("x"); err == nil {
		t.Fatal("expected error for duplicate label")
	}
}

// ---------------------------------------------------------------------------
// Stack operations
// ---------------------------------------------------------------------------

func TestPush(t *testing.T) {
	expectBytes(t, "PUSH RAX", emit(func(a *Assembler) { a.Push(RAX) }), []byte{0x50})
	expectBytes(t, "PUSH RDI", emit(func(a *Assembler) { a.Push(RDI) }), []byte{0x57})
	expectBytes(t, "PUSH R8", emit(func(a *Assembler) { a.Push(R8) }), []byte{0x41, 0x50})
	expectBytes(t, "PUSH R15", emit(func(a *Assembler) { a.Push(R15) }), []byte{0x41, 0x57})
}

func TestPop(t *testing.T) {
	expectBytes(t, "POP RAX", emit(func(a *Assembler) { a.Pop(RAX) }), []byte{0x58})
	expectBytes(t, "POP RDI", emit(func(a *Assembler) { a.Pop(RDI) }), []byte{0x5F})
	expectBytes(t, "POP R8", emit(func(a *Assembler) { a.Pop(R8) }), []byte{0x41, 0x58})
	expectBytes(t, "POP R15", emit(func(a *Assembler) { a.Pop(R15) }), []byte{0x41, 0x5F})
}

// ---------------------------------------------------------------------------
// MOV instructions
// ---------------------------------------------------------------------------

func TestMovFromRegToReg(t *testing.T) {
	// MOV RAX, RCX → REX.W(48) 89 C8(mod=11,src=RCX(001),dst=RAX(000))
	expectBytes(t, "MOV RAX,RCX", emit(func(a *Assembler) { a.MovRegToReg(RAX, RCX) }),
		[]byte{0x48, 0x89, 0xC8})
	// MOV R8, R9 → REX.WRB(4D) 89 C8
	expectBytes(t, "MOV R8,R9", emit(func(a *Assembler) { a.MovRegToReg(R8, R9) }),
		[]byte{0x4D, 0x89, 0xC8})
	// MOV RAX, R15 → REX.WR(4C) 89 F8
	expectBytes(t, "MOV RAX,R15", emit(func(a *Assembler) { a.MovRegToReg(RAX, R15) }),
		[]byte{0x4C, 0x89, 0xF8})
	// MOV R12, RCX → REX.WB(49) 89 CC
	expectBytes(t, "MOV R12,RCX", emit(func(a *Assembler) { a.MovRegToReg(R12, RCX) }),
		[]byte{0x49, 0x89, 0xCC})
}

func TestMovFromImm32ToReg(t *testing.T) {
	// MOV RAX, 42 (sign-extended) → REX.W(48) C7 C0 2A000000
	expectBytes(t, "MOV RAX,42", emit(func(a *Assembler) { a.MovImm32ToReg(RAX, 42) }),
		[]byte{0x48, 0xC7, 0xC0, 0x2A, 0x00, 0x00, 0x00})
	// MOV R10, -1 → REX.WB(49) C7 C2 FFFFFFFF
	expectBytes(t, "MOV R10,-1", emit(func(a *Assembler) { a.MovImm32ToReg(R10, -1) }),
		[]byte{0x49, 0xC7, 0xC2, 0xFF, 0xFF, 0xFF, 0xFF})
}

func TestMovFromImm64ToReg(t *testing.T) {
	// MOV RAX, 0x123456789ABCDEF0 → REX.W(48) B8 F0DEBC9A78563412
	expectBytes(t, "MOVABS RAX", emit(func(a *Assembler) { a.MovImm64ToReg(RAX, 0x123456789ABCDEF0) }),
		[]byte{0x48, 0xB8, 0xF0, 0xDE, 0xBC, 0x9A, 0x78, 0x56, 0x34, 0x12})
	// MOV R11, 0xFF → REX.WB(49) BB FF00000000000000
	expectBytes(t, "MOVABS R11", emit(func(a *Assembler) { a.MovImm64ToReg(R11, 0xFF) }),
		[]byte{0x49, 0xBB, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func TestMovFromMemToReg(t *testing.T) {
	// MOV RAX, [RCX] → REX.W(48) 8B 01
	expectBytes(t, "MOV RAX,[RCX]", emit(func(a *Assembler) { a.MovMemToReg(RAX, RCX, 0) }),
		[]byte{0x48, 0x8B, 0x01})
	// MOV RAX, [RCX+0x10] → REX.W(48) 8B 41 10
	expectBytes(t, "MOV RAX,[RCX+16]", emit(func(a *Assembler) { a.MovMemToReg(RAX, RCX, 0x10) }),
		[]byte{0x48, 0x8B, 0x41, 0x10})
	// MOV RAX, [RCX+0x100] → REX.W(48) 8B 81 00010000
	expectBytes(t, "MOV RAX,[RCX+256]", emit(func(a *Assembler) { a.MovMemToReg(RAX, RCX, 0x100) }),
		[]byte{0x48, 0x8B, 0x81, 0x00, 0x01, 0x00, 0x00})
}

func TestMovFromMemToReg_RSP_RBP(t *testing.T) {
	// RSP as base needs SIB: MOV RAX, [RSP] → 48 8B 04 24
	expectBytes(t, "MOV RAX,[RSP]", emit(func(a *Assembler) { a.MovMemToReg(RAX, RSP, 0) }),
		[]byte{0x48, 0x8B, 0x04, 0x24})
	// RBP with disp=0 needs disp8=0: MOV RAX, [RBP] → 48 8B 45 00
	expectBytes(t, "MOV RAX,[RBP]", emit(func(a *Assembler) { a.MovMemToReg(RAX, RBP, 0) }),
		[]byte{0x48, 0x8B, 0x45, 0x00})
}

func TestMovFromRegToMem(t *testing.T) {
	// MOV [RCX], RAX → REX.W(48) 89 01
	expectBytes(t, "MOV [RCX],RAX", emit(func(a *Assembler) { a.MovRegToMem(RCX, 0, RAX) }),
		[]byte{0x48, 0x89, 0x01})
	// MOV [RCX+8], RAX → 48 89 41 08
	expectBytes(t, "MOV [RCX+8],RAX", emit(func(a *Assembler) { a.MovRegToMem(RCX, 8, RAX) }),
		[]byte{0x48, 0x89, 0x41, 0x08})
}

func TestMovFromByteToRegZx(t *testing.T) {
	// MOVZX RAX, CL → 48 0F B6 C1
	expectBytes(t, "MOVZX RAX,CL", emit(func(a *Assembler) { a.MovFromByteToRegZx(RAX, RCX) }),
		[]byte{0x48, 0x0F, 0xB6, 0xC1})
}

func TestMovFromByteToRegSx(t *testing.T) {
	// MOVSX RAX, CL → 48 0F BE C1
	expectBytes(t, "MOVSX RAX,CL", emit(func(a *Assembler) { a.MovFromByteToRegSx(RAX, RCX) }),
		[]byte{0x48, 0x0F, 0xBE, 0xC1})
}

func TestMovFromWordToRegZx(t *testing.T) {
	// MOVZX RAX, CX → 48 0F B7 C1
	expectBytes(t, "MOVZX RAX,CX", emit(func(a *Assembler) { a.MovFromWordToRegZx(RAX, RCX) }),
		[]byte{0x48, 0x0F, 0xB7, 0xC1})
}

func TestMovFromWordToRegSx(t *testing.T) {
	// MOVSX RAX, CX → 48 0F BF C1
	expectBytes(t, "MOVSX RAX,CX", emit(func(a *Assembler) { a.MovFromWordToRegSx(RAX, RCX) }),
		[]byte{0x48, 0x0F, 0xBF, 0xC1})
}

func TestMovFromDwordToRegSx(t *testing.T) {
	// MOVSXD RAX, ECX → 48 63 C1
	expectBytes(t, "MOVSXD RAX,ECX", emit(func(a *Assembler) { a.MovFromDwordToRegSx(RAX, RCX) }),
		[]byte{0x48, 0x63, 0xC1})
}

// ---------------------------------------------------------------------------
// Arithmetic
// ---------------------------------------------------------------------------

func TestAddRegReg(t *testing.T) {
	// ADD RAX, RCX → 48 01 C8
	expectBytes(t, "ADD RAX,RCX", emit(func(a *Assembler) { a.AddRegReg(RAX, RCX) }),
		[]byte{0x48, 0x01, 0xC8})
	// ADD R8, R9 → 4D 01 C8
	expectBytes(t, "ADD R8,R9", emit(func(a *Assembler) { a.AddRegReg(R8, R9) }),
		[]byte{0x4D, 0x01, 0xC8})
}

func TestAddRegImm32(t *testing.T) {
	// ADD RAX, 1 (imm8) → 48 83 C0 01
	expectBytes(t, "ADD RAX,1", emit(func(a *Assembler) { a.AddRegImm32(RAX, 1) }),
		[]byte{0x48, 0x83, 0xC0, 0x01})
	// ADD RAX, 256 (imm32) → 48 81 C0 00010000
	expectBytes(t, "ADD RAX,256", emit(func(a *Assembler) { a.AddRegImm32(RAX, 256) }),
		[]byte{0x48, 0x81, 0xC0, 0x00, 0x01, 0x00, 0x00})
	// ADD R10, -5 → 49 83 C2 FB
	expectBytes(t, "ADD R10,-5", emit(func(a *Assembler) { a.AddRegImm32(R10, -5) }),
		[]byte{0x49, 0x83, 0xC2, 0xFB})
}

func TestAdd32RegReg(t *testing.T) {
	// ADD EAX, ECX → 01 C8 (no REX needed for low regs)
	expectBytes(t, "ADD EAX,ECX", emit(func(a *Assembler) { a.Add32RegReg(RAX, RCX) }),
		[]byte{0x01, 0xC8})
	// ADD R8D, R9D → 45 01 C8
	expectBytes(t, "ADD R8D,R9D", emit(func(a *Assembler) { a.Add32RegReg(R8, R9) }),
		[]byte{0x45, 0x01, 0xC8})
}

func TestSubRegReg(t *testing.T) {
	// SUB RAX, RCX → 48 29 C8
	expectBytes(t, "SUB RAX,RCX", emit(func(a *Assembler) { a.SubRegReg(RAX, RCX) }),
		[]byte{0x48, 0x29, 0xC8})
}

func TestSubRegImm32(t *testing.T) {
	// SUB RSP, 8 → 48 83 EC 08
	expectBytes(t, "SUB RSP,8", emit(func(a *Assembler) { a.SubRegImm32(RSP, 8) }),
		[]byte{0x48, 0x83, 0xEC, 0x08})
	// SUB RAX, 1000 → 48 81 E8 E8030000
	expectBytes(t, "SUB RAX,1000", emit(func(a *Assembler) { a.SubRegImm32(RAX, 1000) }),
		[]byte{0x48, 0x81, 0xE8, 0xE8, 0x03, 0x00, 0x00})
}

func TestImulRegReg(t *testing.T) {
	// IMUL RAX, RCX → 48 0F AF C1
	expectBytes(t, "IMUL RAX,RCX", emit(func(a *Assembler) { a.ImulRegReg(RAX, RCX) }),
		[]byte{0x48, 0x0F, 0xAF, 0xC1})
}

func TestImulRegImm32(t *testing.T) {
	// IMUL RAX, RCX, 10 (imm8) → 48 6B C1 0A
	expectBytes(t, "IMUL RAX,RCX,10", emit(func(a *Assembler) { a.ImulRegImm32(RAX, RCX, 10) }),
		[]byte{0x48, 0x6B, 0xC1, 0x0A})
	// IMUL RAX, RCX, 300 (imm32) → 48 69 C1 2C010000
	expectBytes(t, "IMUL RAX,RCX,300", emit(func(a *Assembler) { a.ImulRegImm32(RAX, RCX, 300) }),
		[]byte{0x48, 0x69, 0xC1, 0x2C, 0x01, 0x00, 0x00})
}

func TestIdiv(t *testing.T) {
	// IDIV RCX → 48 F7 F9
	expectBytes(t, "IDIV RCX", emit(func(a *Assembler) { a.Idiv(RCX) }),
		[]byte{0x48, 0xF7, 0xF9})
	// IDIV R12 → 49 F7 FC
	expectBytes(t, "IDIV R12", emit(func(a *Assembler) { a.Idiv(R12) }),
		[]byte{0x49, 0xF7, 0xFC})
}

func TestDiv(t *testing.T) {
	// DIV RCX → 48 F7 F1
	expectBytes(t, "DIV RCX", emit(func(a *Assembler) { a.Div(RCX) }),
		[]byte{0x48, 0xF7, 0xF1})
}

func TestNeg(t *testing.T) {
	// NEG RAX → 48 F7 D8
	expectBytes(t, "NEG RAX", emit(func(a *Assembler) { a.Neg(RAX) }),
		[]byte{0x48, 0xF7, 0xD8})
	// NEG R8 → 49 F7 D8
	expectBytes(t, "NEG R8", emit(func(a *Assembler) { a.Neg(R8) }),
		[]byte{0x49, 0xF7, 0xD8})
}

func TestCqo(t *testing.T) {
	// CQO → 48 99
	expectBytes(t, "CQO", emit(func(a *Assembler) { a.Cqo() }),
		[]byte{0x48, 0x99})
}

func TestXor_Zero(t *testing.T) {
	// XOR RDX, RDX → 48 31 D2
	expectBytes(t, "XOR RDX,RDX", emit(func(a *Assembler) { a.Xor(RDX) }),
		[]byte{0x48, 0x31, 0xD2})
}

// ---------------------------------------------------------------------------
// Bitwise operations
// ---------------------------------------------------------------------------

func TestAndRegReg(t *testing.T) {
	// AND RAX, RCX → 48 21 C8
	expectBytes(t, "AND RAX,RCX", emit(func(a *Assembler) { a.AndRegReg(RAX, RCX) }),
		[]byte{0x48, 0x21, 0xC8})
}

func TestAndRegImm32(t *testing.T) {
	// AND RAX, 0xFF → 48 81 E0 FF000000
	expectBytes(t, "AND RAX,0xFF", emit(func(a *Assembler) { a.AndRegImm32(RAX, 0xFF) }),
		[]byte{0x48, 0x81, 0xE0, 0xFF, 0x00, 0x00, 0x00})
	// AND RAX, 7 (imm8) → 48 83 E0 07
	expectBytes(t, "AND RAX,7", emit(func(a *Assembler) { a.AndRegImm32(RAX, 7) }),
		[]byte{0x48, 0x83, 0xE0, 0x07})
}

func TestOrRegReg(t *testing.T) {
	// OR RAX, RCX → 48 09 C8
	expectBytes(t, "OR RAX,RCX", emit(func(a *Assembler) { a.OrRegReg(RAX, RCX) }),
		[]byte{0x48, 0x09, 0xC8})
}

func TestOrRegImm32(t *testing.T) {
	// OR RAX, 3 → 48 83 C8 03
	expectBytes(t, "OR RAX,3", emit(func(a *Assembler) { a.OrRegImm32(RAX, 3) }),
		[]byte{0x48, 0x83, 0xC8, 0x03})
}

func TestXorRegReg(t *testing.T) {
	// XOR RAX, RCX → 48 31 C8
	expectBytes(t, "XOR RAX,RCX", emit(func(a *Assembler) { a.XorRegReg(RAX, RCX) }),
		[]byte{0x48, 0x31, 0xC8})
}

func TestXorRegImm32(t *testing.T) {
	// XOR RAX, -1 → 48 83 F0 FF
	expectBytes(t, "XOR RAX,-1", emit(func(a *Assembler) { a.XorRegImm32(RAX, -1) }),
		[]byte{0x48, 0x83, 0xF0, 0xFF})
}

func TestNot(t *testing.T) {
	// NOT RAX → 48 F7 D0
	expectBytes(t, "NOT RAX", emit(func(a *Assembler) { a.Not(RAX) }),
		[]byte{0x48, 0xF7, 0xD0})
}

func TestShlRegCL(t *testing.T) {
	// SHL RAX, CL → 48 D3 E0
	expectBytes(t, "SHL RAX,CL", emit(func(a *Assembler) { a.ShlRegCL(RAX) }),
		[]byte{0x48, 0xD3, 0xE0})
}

func TestShlRegImm(t *testing.T) {
	// SHL RAX, 1 → 48 D1 E0
	expectBytes(t, "SHL RAX,1", emit(func(a *Assembler) { a.ShlRegImm(RAX, 1) }),
		[]byte{0x48, 0xD1, 0xE0})
	// SHL RAX, 4 → 48 C1 E0 04
	expectBytes(t, "SHL RAX,4", emit(func(a *Assembler) { a.ShlRegImm(RAX, 4) }),
		[]byte{0x48, 0xC1, 0xE0, 0x04})
}

func TestShrRegCL(t *testing.T) {
	// SHR RAX, CL → 48 D3 E8
	expectBytes(t, "SHR RAX,CL", emit(func(a *Assembler) { a.ShrRegCL(RAX) }),
		[]byte{0x48, 0xD3, 0xE8})
}

func TestShrRegImm(t *testing.T) {
	// SHR RAX, 1 → 48 D1 E8
	expectBytes(t, "SHR RAX,1", emit(func(a *Assembler) { a.ShrRegImm(RAX, 1) }),
		[]byte{0x48, 0xD1, 0xE8})
	// SHR RAX, 8 → 48 C1 E8 08
	expectBytes(t, "SHR RAX,8", emit(func(a *Assembler) { a.ShrRegImm(RAX, 8) }),
		[]byte{0x48, 0xC1, 0xE8, 0x08})
}

func TestSarRegCL(t *testing.T) {
	// SAR RAX, CL → 48 D3 F8
	expectBytes(t, "SAR RAX,CL", emit(func(a *Assembler) { a.SarRegCL(RAX) }),
		[]byte{0x48, 0xD3, 0xF8})
}

func TestSarRegImm(t *testing.T) {
	// SAR RAX, 1 → 48 D1 F8
	expectBytes(t, "SAR RAX,1", emit(func(a *Assembler) { a.SarRegImm(RAX, 1) }),
		[]byte{0x48, 0xD1, 0xF8})
	// SAR RAX, 16 → 48 C1 F8 10
	expectBytes(t, "SAR RAX,16", emit(func(a *Assembler) { a.SarRegImm(RAX, 16) }),
		[]byte{0x48, 0xC1, 0xF8, 0x10})
}

func TestRolRegCL(t *testing.T) {
	// ROL RAX, CL → 48 D3 C0
	expectBytes(t, "ROL RAX,CL", emit(func(a *Assembler) { a.RolRegCL(RAX) }),
		[]byte{0x48, 0xD3, 0xC0})
}

func TestRolRegImm(t *testing.T) {
	// ROL RAX, 1 → 48 D1 C0
	expectBytes(t, "ROL RAX,1", emit(func(a *Assembler) { a.RolRegImm(RAX, 1) }),
		[]byte{0x48, 0xD1, 0xC0})
	// ROL RAX, 5 → 48 C1 C0 05
	expectBytes(t, "ROL RAX,5", emit(func(a *Assembler) { a.RolRegImm(RAX, 5) }),
		[]byte{0x48, 0xC1, 0xC0, 0x05})
}

func TestRorRegCL(t *testing.T) {
	// ROR RAX, CL → 48 D3 C8
	expectBytes(t, "ROR RAX,CL", emit(func(a *Assembler) { a.RorRegCL(RAX) }),
		[]byte{0x48, 0xD3, 0xC8})
}

func TestRorRegImm(t *testing.T) {
	// ROR RAX, 1 → 48 D1 C8
	expectBytes(t, "ROR RAX,1", emit(func(a *Assembler) { a.RorRegImm(RAX, 1) }),
		[]byte{0x48, 0xD1, 0xC8})
}

func TestPopcnt(t *testing.T) {
	// POPCNT RAX, RCX → F3 48 0F B8 C1
	expectBytes(t, "POPCNT RAX,RCX", emit(func(a *Assembler) { a.Popcnt(RAX, RCX) }),
		[]byte{0xF3, 0x48, 0x0F, 0xB8, 0xC1})
	// POPCNT R8, R9 → F3 4D 0F B8 C1
	expectBytes(t, "POPCNT R8,R9", emit(func(a *Assembler) { a.Popcnt(R8, R9) }),
		[]byte{0xF3, 0x4D, 0x0F, 0xB8, 0xC1})
}

func TestLzcnt(t *testing.T) {
	// LZCNT RAX, RCX → F3 48 0F BD C1
	expectBytes(t, "LZCNT RAX,RCX", emit(func(a *Assembler) { a.Lzcnt(RAX, RCX) }),
		[]byte{0xF3, 0x48, 0x0F, 0xBD, 0xC1})
}

func TestTzcnt(t *testing.T) {
	// TZCNT RAX, RCX → F3 48 0F BC C1
	expectBytes(t, "TZCNT RAX,RCX", emit(func(a *Assembler) { a.Tzcnt(RAX, RCX) }),
		[]byte{0xF3, 0x48, 0x0F, 0xBC, 0xC1})
}

func TestBswap(t *testing.T) {
	// BSWAP RAX → 48 0F C8
	expectBytes(t, "BSWAP RAX", emit(func(a *Assembler) { a.Bswap(RAX) }),
		[]byte{0x48, 0x0F, 0xC8})
	// BSWAP R10 → 49 0F CA
	expectBytes(t, "BSWAP R10", emit(func(a *Assembler) { a.Bswap(R10) }),
		[]byte{0x49, 0x0F, 0xCA})
}

func TestBswap32(t *testing.T) {
	// BSWAP EAX → 0F C8
	expectBytes(t, "BSWAP EAX", emit(func(a *Assembler) { a.Bswap32(RAX) }),
		[]byte{0x0F, 0xC8})
	// BSWAP R10D → 41 0F CA
	expectBytes(t, "BSWAP R10D", emit(func(a *Assembler) { a.Bswap32(R10) }),
		[]byte{0x41, 0x0F, 0xCA})
}

// ---------------------------------------------------------------------------
// Compare & condition
// ---------------------------------------------------------------------------

func TestCmpRegReg(t *testing.T) {
	// CMP RAX, RCX → 48 39 C8
	expectBytes(t, "CMP RAX,RCX", emit(func(a *Assembler) { a.CmpRegReg(RAX, RCX) }),
		[]byte{0x48, 0x39, 0xC8})
}

func TestCmpRegImm32(t *testing.T) {
	// CMP RAX, 0 → 48 83 F8 00
	expectBytes(t, "CMP RAX,0", emit(func(a *Assembler) { a.CmpRegImm32(RAX, 0) }),
		[]byte{0x48, 0x83, 0xF8, 0x00})
	// CMP RAX, 1000 → 48 81 F8 E8030000
	expectBytes(t, "CMP RAX,1000", emit(func(a *Assembler) { a.CmpRegImm32(RAX, 1000) }),
		[]byte{0x48, 0x81, 0xF8, 0xE8, 0x03, 0x00, 0x00})
}

func TestTestRegReg(t *testing.T) {
	// TEST RAX, RAX → 48 85 C0
	expectBytes(t, "TEST RAX,RAX", emit(func(a *Assembler) { a.TestRegReg(RAX, RAX) }),
		[]byte{0x48, 0x85, 0xC0})
}

func TestSetCC(t *testing.T) {
	// SETE AL → 0F 94 C0
	expectBytes(t, "SETE AL", emit(func(a *Assembler) { a.SetCC(CondEQ, RAX) }),
		[]byte{0x0F, 0x94, 0xC0})
	// SETL R8B → 41 0F 9C C0
	expectBytes(t, "SETL R8B", emit(func(a *Assembler) { a.SetCC(CondLT, R8) }),
		[]byte{0x41, 0x0F, 0x9C, 0xC0})
	// SETB BPL → 40 0F 92 C5 (REX required: rm=5 is CH without REX)
	expectBytes(t, "SETB BPL", emit(func(a *Assembler) { a.SetCC(CondB, RBP) }),
		[]byte{0x40, 0x0F, 0x92, 0xC5})
}

func TestCMovCC(t *testing.T) {
	// CMOVE RAX, RCX → 48 0F 44 C1
	expectBytes(t, "CMOVE RAX,RCX", emit(func(a *Assembler) { a.CMovCC(CondEQ, RAX, RCX) }),
		[]byte{0x48, 0x0F, 0x44, 0xC1})
	// CMOVB R8, R9 → 4D 0F 42 C1
	expectBytes(t, "CMOVB R8,R9", emit(func(a *Assembler) { a.CMovCC(CondB, R8, R9) }),
		[]byte{0x4D, 0x0F, 0x42, 0xC1})
}

// ---------------------------------------------------------------------------
// Jumps & flow control
// ---------------------------------------------------------------------------

func TestJmpReg(t *testing.T) {
	// JMP RAX → FF E0 (mod=11, /4, rm=0)
	expectBytes(t, "JMP RAX", emit(func(a *Assembler) { a.JmpReg(RAX) }),
		[]byte{0xFF, 0xE0})
	// JMP R11 → 41 FF E3
	expectBytes(t, "JMP R11", emit(func(a *Assembler) { a.JmpReg(R11) }),
		[]byte{0x41, 0xFF, 0xE3})
}

func TestCall(t *testing.T) {
	a := NewAssembler()
	a.Call("func1")
	_ = a.BindLabel("func1")
	a.Ret()
	code, err := a.Finalize()
	if err != nil {
		t.Fatal(err)
	}
	// CALL rel32 at 0: E8 xx xx xx xx → target=5, rel=5-(1+4)=0
	expectBytes(t, "CALL", code, []byte{
		0xE8, 0x00, 0x00, 0x00, 0x00, // CALL +0
		0xC3, // RET
	})
}

func TestCallReg(t *testing.T) {
	// CALL RAX → FF D0 (mod=11, /2, rm=0)
	expectBytes(t, "CALL RAX", emit(func(a *Assembler) { a.CallReg(RAX) }),
		[]byte{0xFF, 0xD0})
	// CALL R12 → 41 FF D4
	expectBytes(t, "CALL R12", emit(func(a *Assembler) { a.CallReg(R12) }),
		[]byte{0x41, 0xFF, 0xD4})
}

func TestJcc_Forward(t *testing.T) {
	a := NewAssembler()
	a.Jcc(CondEQ, "equal")
	a.Nop()
	_ = a.BindLabel("equal")
	a.Ret()
	code, err := a.Finalize()
	if err != nil {
		t.Fatal(err)
	}
	// JE rel32: 0F 84 xx xx xx xx → target=7, rel=7-(2+4)=1
	expectBytes(t, "JE forward", code, []byte{
		0x0F, 0x84, 0x01, 0x00, 0x00, 0x00, // JE +1
		0x90, // NOP
		0xC3, // RET
	})
}

func TestRet(t *testing.T) {
	expectBytes(t, "RET", emit(func(a *Assembler) { a.Ret() }), []byte{0xC3})
}

func TestNop(t *testing.T) {
	expectBytes(t, "NOP", emit(func(a *Assembler) { a.Nop() }), []byte{0x90})
}

// ---------------------------------------------------------------------------
// Memory load/store (different sizes)
// ---------------------------------------------------------------------------

func TestLoadByte(t *testing.T) {
	// MOVZX EAX, BYTE [RCX+4] → 0F B6 41 04
	expectBytes(t, "LoadByte", emit(func(a *Assembler) { a.LoadByte(RAX, RCX, 4) }),
		[]byte{0x0F, 0xB6, 0x41, 0x04})
}

func TestLoadSignedByte(t *testing.T) {
	// MOVSX RAX, BYTE [RCX+4] → 48 0F BE 41 04
	expectBytes(t, "LoadSignedByte", emit(func(a *Assembler) { a.LoadSignedByte(RAX, RCX, 4) }),
		[]byte{0x48, 0x0F, 0xBE, 0x41, 0x04})
}

func TestLoadWord(t *testing.T) {
	// MOVZX EAX, WORD [RCX] → 0F B7 01
	expectBytes(t, "LoadWord", emit(func(a *Assembler) { a.LoadWord(RAX, RCX, 0) }),
		[]byte{0x0F, 0xB7, 0x01})
}

func TestLoadSignedWord(t *testing.T) {
	// MOVSX RAX, WORD [RCX] → 48 0F BF 01
	expectBytes(t, "LoadSignedWord", emit(func(a *Assembler) { a.LoadSignedWord(RAX, RCX, 0) }),
		[]byte{0x48, 0x0F, 0xBF, 0x01})
}

func TestLoadDword(t *testing.T) {
	// MOV EAX, [RCX+8] → 8B 41 08
	expectBytes(t, "LoadDword", emit(func(a *Assembler) { a.LoadDword(RAX, RCX, 8) }),
		[]byte{0x8B, 0x41, 0x08})
}

func TestLoadSignedDword(t *testing.T) {
	// MOVSXD RAX, [RCX+8] → 48 63 41 08
	expectBytes(t, "LoadSignedDword", emit(func(a *Assembler) { a.LoadSignedDword(RAX, RCX, 8) }),
		[]byte{0x48, 0x63, 0x41, 0x08})
}

func TestLoadQword(t *testing.T) {
	// MOV RAX, [RCX+16] → 48 8B 41 10
	expectBytes(t, "LoadQword", emit(func(a *Assembler) { a.LoadQword(RAX, RCX, 0x10) }),
		[]byte{0x48, 0x8B, 0x41, 0x10})
}

func TestStoreByte(t *testing.T) {
	// MOV BYTE [RCX+4], AL → 88 41 04
	expectBytes(t, "StoreByte", emit(func(a *Assembler) { a.StoreByte(RCX, 4, RAX) }),
		[]byte{0x88, 0x41, 0x04})
}

func TestStoreByteBPL(t *testing.T) {
	// MOV BYTE [RCX], BPL → 40 88 29  (REX required: reg=5 is CH without REX)
	expectBytes(t, "StoreByte BPL", emit(func(a *Assembler) { a.StoreByte(RCX, 0, RBP) }),
		[]byte{0x40, 0x88, 0x29})
}

func TestStoreByteR12B(t *testing.T) {
	// MOV BYTE [RCX], R12B → 44 88 21
	expectBytes(t, "StoreByte R12B", emit(func(a *Assembler) { a.StoreByte(RCX, 0, R12) }),
		[]byte{0x44, 0x88, 0x21})
}

func TestStoreWord(t *testing.T) {
	// MOV WORD [RCX], AX → 66 89 01
	expectBytes(t, "StoreWord", emit(func(a *Assembler) { a.StoreWord(RCX, 0, RAX) }),
		[]byte{0x66, 0x89, 0x01})
}

func TestStoreDword(t *testing.T) {
	// MOV DWORD [RCX+8], EAX → 89 41 08
	expectBytes(t, "StoreDword", emit(func(a *Assembler) { a.StoreDword(RCX, 8, RAX) }),
		[]byte{0x89, 0x41, 0x08})
}

func TestStoreQword(t *testing.T) {
	// MOV QWORD [RCX+16], RAX → 48 89 41 10
	expectBytes(t, "StoreQword", emit(func(a *Assembler) { a.StoreQword(RCX, 0x10, RAX) }),
		[]byte{0x48, 0x89, 0x41, 0x10})
}

// ---------------------------------------------------------------------------
// Multiplication upper half
// ---------------------------------------------------------------------------

func TestMulHigh(t *testing.T) {
	// MUL RCX → 48 F7 E1
	expectBytes(t, "MUL RCX", emit(func(a *Assembler) { a.MulHigh(RCX) }),
		[]byte{0x48, 0xF7, 0xE1})
}

func TestIMulHigh(t *testing.T) {
	// IMUL RCX (one-operand form) → 48 F7 E9
	expectBytes(t, "IMUL RCX(1op)", emit(func(a *Assembler) { a.IMulHigh(RCX) }),
		[]byte{0x48, 0xF7, 0xE9})
}

// ---------------------------------------------------------------------------
// Extended register coverage (REX prefix correctness)
// ---------------------------------------------------------------------------

func TestExtendedRegister_AllInstructions(t *testing.T) {
	tests := []struct {
		name string
		emit func(a *Assembler)
		want []byte
	}{
		{"ADD R8,R15", func(a *Assembler) { a.AddRegReg(R8, R15) }, []byte{0x4D, 0x01, 0xF8}},
		{"SUB R14,R13", func(a *Assembler) { a.SubRegReg(R14, R13) }, []byte{0x4D, 0x29, 0xEE}},
		{"AND R8,R9", func(a *Assembler) { a.AndRegReg(R8, R9) }, []byte{0x4D, 0x21, 0xC8}},
		{"OR R10,R11", func(a *Assembler) { a.OrRegReg(R10, R11) }, []byte{0x4D, 0x09, 0xDA}},
		{"XOR R12,R13", func(a *Assembler) { a.XorRegReg(R12, R13) }, []byte{0x4D, 0x31, 0xEC}},
		{"NOT R8", func(a *Assembler) { a.Not(R8) }, []byte{0x49, 0xF7, 0xD0}},
		{"NEG R15", func(a *Assembler) { a.Neg(R15) }, []byte{0x49, 0xF7, 0xDF}},
		{"CMP R8,R9", func(a *Assembler) { a.CmpRegReg(R8, R9) }, []byte{0x4D, 0x39, 0xC8}},
		{"TEST R10,R10", func(a *Assembler) { a.TestRegReg(R10, R10) }, []byte{0x4D, 0x85, 0xD2}},
		{"IMUL R8,R9", func(a *Assembler) { a.ImulRegReg(R8, R9) }, []byte{0x4D, 0x0F, 0xAF, 0xC1}},
		{"SHL R8,CL", func(a *Assembler) { a.ShlRegCL(R8) }, []byte{0x49, 0xD3, 0xE0}},
		{"SHR R8,4", func(a *Assembler) { a.ShrRegImm(R8, 4) }, []byte{0x49, 0xC1, 0xE8, 0x04}},
		{"MOV R8,R9", func(a *Assembler) { a.MovRegToReg(R8, R9) }, []byte{0x4D, 0x89, 0xC8}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := emit(tt.emit)
			expectBytes(t, tt.name, got, tt.want)
		})
	}
}

// ---------------------------------------------------------------------------
// 32-bit instruction variants
// ---------------------------------------------------------------------------

func TestInstructions32bit(t *testing.T) {
	tests := []struct {
		name string
		emit func(a *Assembler)
		want []byte
	}{
		{"SUB32 EAX,ECX", func(a *Assembler) { a.Sub32RegReg(RAX, RCX) }, []byte{0x29, 0xC8}},
		{"IMUL32 EAX,ECX", func(a *Assembler) { a.Imul32RegReg(RAX, RCX) }, []byte{0x0F, 0xAF, 0xC1}},
		{"IDIV32 ECX", func(a *Assembler) { a.Idiv32(RCX) }, []byte{0xF7, 0xF9}},
		{"DIV32 ECX", func(a *Assembler) { a.Div32(RCX) }, []byte{0xF7, 0xF1}},
		{"SHL32 EAX,CL", func(a *Assembler) { a.Shl32RegCL(RAX) }, []byte{0xD3, 0xE0}},
		{"SHR32 EAX,4", func(a *Assembler) { a.Shr32RegImm(RAX, 4) }, []byte{0xC1, 0xE8, 0x04}},
		{"SAR32 EAX,CL", func(a *Assembler) { a.Sar32RegCL(RAX) }, []byte{0xD3, 0xF8}},
		{"ROL32 EAX,1", func(a *Assembler) { a.Rol32RegImm(RAX, 1) }, []byte{0xD1, 0xC0}},
		{"ROR32 EAX,CL", func(a *Assembler) { a.Ror32RegCL(RAX) }, []byte{0xD3, 0xC8}},
		{"POPCNT32 EAX,ECX", func(a *Assembler) { a.Popcnt32(RAX, RCX) }, []byte{0xF3, 0x0F, 0xB8, 0xC1}},
		{"LZCNT32 EAX,ECX", func(a *Assembler) { a.Lzcnt32(RAX, RCX) }, []byte{0xF3, 0x0F, 0xBD, 0xC1}},
		{"TZCNT32 EAX,ECX", func(a *Assembler) { a.Tzcnt32(RAX, RCX) }, []byte{0xF3, 0x0F, 0xBC, 0xC1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := emit(tt.emit)
			expectBytes(t, tt.name, got, tt.want)
		})
	}
}

// ---------------------------------------------------------------------------
// LEA
// ---------------------------------------------------------------------------

func TestLeaRegMem(t *testing.T) {
	// LEA RAX, [RCX+8] → 48 8D 41 08
	expectBytes(t, "LEA RAX,[RCX+8]", emit(func(a *Assembler) { a.LeaRegMem(RAX, RCX, 8) }),
		[]byte{0x48, 0x8D, 0x41, 0x08})
}

func TestLeaRIPRel(t *testing.T) {
	a := NewAssembler()
	a.LeaRIPRel(RAX, "data")
	_ = a.BindLabel("data")
	code, err := a.Finalize()
	if err != nil {
		t.Fatal(err)
	}
	// LEA RAX, [RIP+rel32]: 48 8D 05 xx xx xx xx
	// label at offset 7, fixup at offset 3, size=4
	// rel = 7 - (3+4) = 0
	expectBytes(t, "LEA RIP-rel", code, []byte{0x48, 0x8D, 0x05, 0x00, 0x00, 0x00, 0x00})
}

// ---------------------------------------------------------------------------
// SubMemImm32 (gas metering)
// ---------------------------------------------------------------------------

func TestSubMemImm32(t *testing.T) {
	// SUB QWORD [RCX+8], 5 (imm8) → 48 83 69 08 05
	expectBytes(t, "SUB [RCX+8],5", emit(func(a *Assembler) { a.SubMemImm32(RCX, 8, 5) }),
		[]byte{0x48, 0x83, 0x69, 0x08, 0x05})
	// SUB QWORD [RCX], 1000 (imm32) → 48 81 29 E8030000
	expectBytes(t, "SUB [RCX],1000", emit(func(a *Assembler) { a.SubMemImm32(RCX, 0, 1000) }),
		[]byte{0x48, 0x81, 0x29, 0xE8, 0x03, 0x00, 0x00})
}

// ---------------------------------------------------------------------------
// Memory addressing edge cases
// ---------------------------------------------------------------------------

func TestMemory_R12AsBase(t *testing.T) {
	// R12 has lo3=4 same as RSP, needs SIB
	// MOV RAX, [R12] → 49 8B 04 24
	expectBytes(t, "MOV RAX,[R12]", emit(func(a *Assembler) { a.MovMemToReg(RAX, R12, 0) }),
		[]byte{0x49, 0x8B, 0x04, 0x24})
}

func TestMemory_R13AsBase(t *testing.T) {
	// R13 has lo3=5 same as RBP, needs disp8=0
	// MOV RAX, [R13] → 49 8B 45 00
	expectBytes(t, "MOV RAX,[R13]", emit(func(a *Assembler) { a.MovMemToReg(RAX, R13, 0) }),
		[]byte{0x49, 0x8B, 0x45, 0x00})
}

func TestMemory_LargeDisp(t *testing.T) {
	// MOV RAX, [RCX+0x12345678] → 48 8B 81 78563412
	expectBytes(t, "MOV RAX,[RCX+large]", emit(func(a *Assembler) { a.MovMemToReg(RAX, RCX, 0x12345678) }),
		[]byte{0x48, 0x8B, 0x81, 0x78, 0x56, 0x34, 0x12})
}

// ---------------------------------------------------------------------------
// Assembler Finalize & Reset
// ---------------------------------------------------------------------------

func TestAssembler_Finalize(t *testing.T) {
	a := NewAssembler()
	a.Nop()
	a.Ret()
	code, err := a.Finalize()
	if err != nil {
		t.Fatal(err)
	}
	expectBytes(t, "Finalize", code, []byte{0x90, 0xC3})
}

func TestAssembler_Reset(t *testing.T) {
	a := NewAssembler()
	a.Nop()
	a.Reset()
	if a.Len() != 0 {
		t.Fatalf("Len after reset = %d", a.Len())
	}
}

// ---------------------------------------------------------------------------
// Extended registers with memory operations
// ---------------------------------------------------------------------------

func TestMemoryOps_ExtendedRegs(t *testing.T) {
	tests := []struct {
		name string
		emit func(a *Assembler)
		want []byte
	}{
		{
			"LoadByte R8,[R15+4]",
			func(a *Assembler) { a.LoadByte(R8, R15, 4) },
			[]byte{0x45, 0x0F, 0xB6, 0x47, 0x04},
		},
		{
			"StoreByte [R8+0],R9",
			func(a *Assembler) { a.StoreByte(R8, 0, R9) },
			[]byte{0x45, 0x88, 0x08},
		},
		{
			"StoreWord [R15],RAX",
			func(a *Assembler) { a.StoreWord(R15, 0, RAX) },
			[]byte{0x66, 0x41, 0x89, 0x07},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := emit(tt.emit)
			expectBytes(t, tt.name, got, tt.want)
		})
	}
}
