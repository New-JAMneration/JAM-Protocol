# Opcode Emit 翻譯規則（AI Agent Rebuild Reference）

本文件列出 PVM opcode → x86-64 native code 的翻譯策略，聚焦於 non-trivial 的 emit 邏輯。
簡單的（如 `add_64` = MOV + ADD + MOV）可從 pattern 推導，不重複列出。

**語意驗證**：每條指令的精確語意請對照 `PVM/instructions.go` 的 interpreter 實作。
特別是 edge cases（div-by-zero、overflow、register-update-before-panic）都以 test vector 驗證過。

---

## 通用慣例

| 慣例 | 說明 |
|------|------|
| `RegScratch` = RCX | 所有 emit 可自由使用的 scratch register |
| `RegGuestBase` = R15 | guest memory base + control region 指標 |
| `dReg` / `aReg` / `bReg` | 來自 `InstrMeta` 的 PVM register，已映射為 x86 register |
| 三 operand pattern | `MOV scratch, aReg` → `OP scratch, bReg` → `MOV dReg, scratch`（避免 dst==src aliasing） |
| 32-bit 結果 | 需要 `emitSignExt32(dReg)` = `MOVSXD dReg, dReg`（sext_4 語意） |
| `fallthroughPC(instr)` | = `instr.PC + ProgramCounter(instr.SkipLen) + 1` |

---

## 1. Division（DIV / IDIV）— Register Conflict

x86 的 `DIV` / `IDIV` **固定使用 RDX:RAX**，但 RAX 和 RDX 分別映射了 PVM RA(0) 和 SP(1)。

### 問題

如果 `dReg` 不是 RAX，需要在 DIV 前後 save/restore RAX 和 RDX。

### 解法：`emitSaveRAXRDX` / `emitRestoreDivResult`

```
emitSaveRAXRDX(dReg):
    if dReg != RAX: PUSH RAX    // save PVM RA
    if dReg != RDX: PUSH RDX    // save PVM SP

DIV / IDIV:
    MOV RAX, dividend
    XOR RDX, RDX  (unsigned) / CQO (signed)
    DIV/IDIV divisor
    // 商在 RAX，餘數在 RDX

emitRestoreDivResult(dReg):     // quotient 版
    MOV dReg, RAX               // 把結果搬到目標
    if dReg != RDX: POP RDX     // restore PVM SP
    if dReg != RAX: POP RAX     // restore PVM RA

emitRestoreDivRemainder(dReg):  // remainder 版
    MOV dReg, RDX               // 餘數在 RDX
    if dReg != RDX: POP RDX
    if dReg != RAX: POP RAX
```

### Edge cases

- **div by zero**：先 TEST divisor → JCC divByZero → `dReg = 2^64-1`（quotient）或 `dReg = dividend`（remainder）
- **signed overflow**（`INT64_MIN / -1`）：先 CMP divisor, -1 → JCC overflow → `dReg = INT64_MIN`（quotient）或 `dReg = 0`（remainder）
- 32-bit 版：`Div32` / `Idiv32`，dividend 做 zero-extend 或 sign-extend 到 32-bit

---

## 2. MUL_UPPER — 128-bit 結果

`mul_upper_u` / `mul_upper_s`：結果 = (Reg[rA] × Reg[rB]) 的高 64 bits。

x86 `MUL r64` = RDX:RAX = RAX × operand（高 64 在 RDX）。

```
emitSaveRAXRDX(dReg)
MOV RAX, aReg
MUL bReg           // unsigned: RDX:RAX = RAX × bReg
                   // signed:  IMUL bReg (single-operand form)
MOV dReg, RDX     // 高 64 bits 在 RDX
emitRestoreDivResult (reuse: POP RDX, POP RAX，但取 RDX 不取 RAX)
```

特殊處理：`emitRestoreMulHighResult` 先 `MOV dReg, RDX` 再 POP。

---

## 3. Shift（SHL / SHR / SAR / ROL / ROR）— CL 限制

x86 shift 的動態 shift amount **必須在 CL**（RCX 低 8 bits）。RCX 恰好是 `RegScratch`。

```
MOV RegScratch, bReg    // shift amount → RCX
MOV dReg, aReg          // 被 shift 的值
SHL/SHR/SAR dReg, CL   // x86 自動 mask: %64 (64-bit) 或 %32 (32-bit)
```

**PVM 語意**：shift amount 也是 `% 32` 或 `% 64`，恰好與 x86 行為一致，不需額外 AND。

---

## 4. Branch — CMP + Jcc

```
// emitBranchImm: branch_eq_imm / branch_ne_imm / branch_lt_u_imm / ...
if imm fits int32:
    CMP xReg, imm32
else:
    MOV RCX, imm64
    CMP xReg, RCX

Jcc takenLabel           // CondEQ / CondNE / CondB / CondLT / ...

// Not taken → fallthrough
emitLinkOrExit(linkFallthrough, nextPC)

// Taken → target
takenLabel:
    emitLinkOrExit(linkTaken, targetPC)
```

### Branch 兩 register 版

```
// emitBranchTwoReg: branch_eq / branch_ne / ...
CMP aReg, bReg
Jcc takenLabel
// 其餘同上
```

### Condition Code 對照

| PVM 語意 | x86 ConditionCode |
|----------|------------------|
| `eq` | CondEQ (0x04) |
| `ne` | CondNE (0x05) |
| `lt_u` | CondB (0x02) — unsigned below |
| `le_u` | CondBE (0x06) |
| `ge_u` | CondAE (0x03) |
| `gt_u` | CondA (0x07) |
| `lt_s` | CondLT (0x0C) — signed less |
| `le_s` | CondLE (0x0E) |
| `ge_s` | CondGE (0x0D) |
| `gt_s` | CondGT (0x0F) |

---

## 5. Memory Load / Store — Address Calculation

PVM address = 32-bit（wrap-around）。翻譯模式：

```
// load_u32 Reg[rD], offset(Reg[rA])
MOV ECX, Reg[rA]_32bit          // 取低 32 bits (zero-extend)
ADD ECX, imm32(offset)          // 32-bit ADD (自動 wrap-around)
MOV dReg_32, [R15 + RCX]       // base(R15) + index(RCX)

// store_u32 offset(Reg[rA]), Reg[rB]
MOV ECX, Reg[rA]_32bit
ADD ECX, imm32(offset)
MOV [R15 + RCX], bReg_32
```

不同大小用不同 MOV variant：
- 1 byte: `MOVZX` (zero-ext load) / `MOV byte` (store)
- 2 bytes: `MOVZX word` / `MOV word` (需 0x66 prefix)
- 4 bytes: `MOV dword`
- 8 bytes: `MOV qword` (REX.W)

Signed load（`load_i8` / `load_i16` / `load_i32`）用 `MOVSX` / `MOVSXD`。

**Bounds check**：無。依賴硬體 MMU（PROT_NONE → SIGSEGV → signal handler）。

---

## 6. sbrk — 三路 Inline + Runtime Exit

```
TEST aReg, aReg                 // amount == 0?
JE queryLabel                   // → rD = heapPointer

// amount != 0:
oldHP = [R15 - OffsetHeapPointer]
newHP = oldHP + amount
if newHP < oldHP → overflow → rD = 0
if newHP > heapLimit → rD = 0

nextPageBoundary = (oldHP + 0xFFF) & ~0xFFF
if newHP <= nextPageBoundary:
    // 同一頁，不需 mprotect
    [R15 - OffsetHeapPointer] = newHP
    rD = newHP
else:
    // 跨頁：MUST exit to Go for mprotect
    emitRuntimeExit(SbrkCallID = 0xFF)
    // Go 側 HandleSbrk → unix.Mprotect → 寫回 heapPointer + rD

queryLabel:
    rD = [R15 - OffsetHeapPointer]
```

---

## 7. Gas Check — Per-Instruction (GP v0.7.2)

每條 PVM 指令前：

```
MOV RCX, [R15 - 48]              // 讀 Gas
TEST RCX, RCX                    // SF = (Gas < 0)
JS out_of_gas_pc_XXXXXXXX        // 負數 → OOG
SUB qword [R15 - 48], 1          // 扣 1 gas
```

Block epilogue 為每條指令 emit OOG landing pad：

```
out_of_gas_pc_XXXXXXXX:
    MOV dword [R15-32], instrPC  // ExitPC
    MOV RCX, ExitOOG
    MOV [R15-40], RCX            // ExitReason
    JMP exit_trampoline
```

---

## 8. Trampoline 完整組語

### Entry Trampoline（Go → Native）

```asm
; 入口：RAX = guestBase, RBX = target native addr
PUSH RBX                         ; save callee-saved
PUSH RBP
PUSH R12
PUSH R13
PUSH R14
PUSH R15

MOV  RCX, RBX                   ; save target addr (RBX will be overwritten)
MOV  R15, RAX                   ; R15 = guestBase

LEA  RAX, [RIP + return_label]  ; 計算 return address
MOV  [R15 - 16], RAX            ; ReturnAddr
MOV  [R15 - 8], RSP             ; ReturnStack

; 載入 13 個 PVM registers（pvmRegSlot 決定 offset）
MOV  RAX, [R15 + regOffset(0)]  ; PVM RA → x86 RAX
MOV  RDX, [R15 + regOffset(1)]  ; PVM SP → x86 RDX
MOV  RBX, [R15 + regOffset(2)]  ; PVM T0 → x86 RBX
MOV  RSI, [R15 + regOffset(3)]  ; ...
MOV  RDI, [R15 + regOffset(4)]
MOV  R8,  [R15 + regOffset(5)]
MOV  R9,  [R15 + regOffset(6)]
MOV  R10, [R15 + regOffset(7)]
MOV  R11, [R15 + regOffset(8)]
MOV  R12, [R15 + regOffset(9)]
MOV  R13, [R15 + regOffset(10)]
MOV  R14, [R15 + regOffset(11)]
MOV  RBP, [R15 + regOffset(12)] ; PVM A5 → x86 RBP

JMP  RCX                         ; 跳到 native block

return_label:
POP  R15
POP  R14
POP  R13
POP  R12
POP  RBP
POP  RBX
RET                              ; 回到 Go
```

### Exit Trampoline（Native → Go）

```asm
exit_trampoline:
; 存回 13 個 PVM registers（反向映射）
MOV  [R15 + regOffset(0)], RAX
MOV  [R15 + regOffset(1)], RDX
MOV  [R15 + regOffset(2)], RBX
MOV  [R15 + regOffset(3)], RSI
MOV  [R15 + regOffset(4)], RDI
MOV  [R15 + regOffset(5)], R8
MOV  [R15 + regOffset(6)], R9
MOV  [R15 + regOffset(7)], R10
MOV  [R15 + regOffset(8)], R11
MOV  [R15 + regOffset(9)], R12
MOV  [R15 + regOffset(10)], R13
MOV  [R15 + regOffset(11)], R14
MOV  [R15 + regOffset(12)], RBP

MOV  RSP, [R15 - 8]             ; 恢復 Go stack
JMP  [R15 - 16]                  ; 跳到 return_label
```

### pvmRegSlot 映射

```
PVM index:  0  1  2  3  4  5  6  7  8  9 10 11 12
Slot:      10 11  2  3  4  5  6  7  8  9  0  1 12
regOffset(i) = -OffsetRegisters + pvmRegSlot[i] * 8
             = -152 + slot * 8
```

RA(0)和SP(1)放在 slot 10,11（offset -72,-64），因為 DIV 路徑頻繁 save/restore RAX/RDX，disp8 範圍更有效率。

---

## 9. callNative（Go Assembly）

```go
// go:nosplit
// func callNative(guestBase, blockAddr, trampolineAddr uintptr)
// Go internal ABI (1.17+): args in RAX, RBX, RCX
//   RAX = guestBase
//   RBX = blockAddr
//   RCX = trampolineAddr
// CALL RCX → 進入 entry trampoline
TEXT ·callNative(SB), NOSPLIT, $0-24
    MOVQ guestBase+0(FP), AX
    MOVQ blockAddr+8(FP), BX
    MOVQ trampolineAddr+16(FP), CX
    CALL CX
    RET
```

注意：Go 1.17+ internal ABI 用 register 傳參，但 `go:nosplit` 函式使用 stack-based FP 存取（plan9 assembly 慣例）。

---

## 10. Opcode Handler 完整註冊表（opcodeHandlers [231]）

```
Opcode  Handler                    Category
------  -------                    --------
0       emitTrap                   NoArgs (→ PANIC)
1       emitFallthrough            NoArgs (block boundary marker)
10      emitEcalli                 OneImm (host call)
20      emitLoadImm64              LoadImm64
30-33   emitStoreImm(1/2/4/8)     TwoImm (store imm to [imm addr])
40      emitJump                   OneOffset (unconditional static)
50      emitJumpInd                OneReg+Imm (indirect via djump)
51      emitLoadImm                OneReg+Imm (load imm → reg)
52-58   emitLoad(1u/1s/2u/2s/4u/4s/8) OneReg+Imm (memory load)
59-62   emitStore(1/2/4/8)        OneReg+Imm (memory store)
70-73   emitStoreImmInd(1/2/4/8)  OneReg+TwoImm
80      emitLoadImmJump            Branch (load + static jump)
81-90   emitBranchImm(EQ/NE/B/BE/AE/A/LT/LE/GE/GT)  Branch
100     emitMoveReg                TwoReg
101     emitSbrk                   TwoReg (heap expansion)
102-107 emitCountSetBits/LeadingZero/TrailingZero (64/32)  TwoReg (bit manipulation)
108-109 emitSignExtend(8/16)       TwoReg
110     emitZeroExtend16           TwoReg
111     emitReverseBytes           TwoReg
120-123 emitStoreInd(1/2/4/8)     TwoReg+Imm
124-130 emitLoadInd(1u/1s/2u/2s/4u/4s/8)  TwoReg+Imm
131     emitAddImm32               TwoReg+Imm (32-bit add imm)
132-134 emitAndImm/XorImm/OrImm   TwoReg+Imm
135     emitMulImm32               TwoReg+Imm
136-137 emitSetLtUImm/SetLtSImm   TwoReg+Imm (compare + set)
138-140 emitShloL/R/SharR Imm32   TwoReg+Imm (shift by imm, 32-bit)
141     emitNegAddImm32            TwoReg+Imm (imm - reg)
142-143 emitSetGtUImm/SetGtSImm   TwoReg+Imm
144-146 emitShloL/R/SharR ImmAlt32  TwoReg+Imm (reg >> imm, 32-bit)
147-148 emitCmovIzImm/CmovNzImm   TwoReg+Imm (conditional move)
149     emitAddImm64               TwoReg+Imm (64-bit add imm)
150     emitMulImm64               TwoReg+Imm
151-153 emitShloL/R/SharR Imm64   TwoReg+Imm (64-bit)
154     emitNegAddImm64            TwoReg+Imm
155-157 emitShloL/R/SharR ImmAlt64  TwoReg+Imm
158-161 emitRotRImm(64/32)/Alt    TwoReg+Imm (rotate)
170-175 emitBranch(EQ/NE/B/LT/AE/GE)  TwoReg+Offset (conditional branch)
180     emitLoadImmJumpInd         TwoReg+TwoImm (load + djump)
190-199 add/sub/mul/divU/divS/remU/remS/shlL/shlR/sarR _32  ThreeReg
200-209 add/sub/mul/divU/divS/remU/remS/shlL/shlR/sarR _64  ThreeReg
210-212 and/xor/or               ThreeReg (bitwise)
213-215 mulUpperSS/UU/SU         ThreeReg (128-bit high)
216-217 setLtU/setLtS            ThreeReg (compare + set)
218-219 cmovIz/cmovNz            ThreeReg (conditional move)
220-223 rotL64/rotL32/rotR64/rotR32  ThreeReg
224-226 andInv/orInv/xnor        ThreeReg
227-230 max/maxU/min/minU        ThreeReg
```

### Compilation Flow

```
CompileBasicBlock(startPC):
  1. ensureDjumpSupport() — lazy init jump table rodata
  2. 找到 BlockMeta (pre-decoded block boundaries)
  3. Pre-compile link targets (strategy-a block linking):
     - compileForLink(fallthroughPC)  → linkFallthrough
     - compileForLink(branchTarget)   → linkTaken
  4. Loop: for each instruction in block:
     a. emitGasCheck(instrPC)         → gas decrement + OOG check
     b. opcodeHandlers[opcode](...)   → emit native code
  5. Epilogue: JMP block_epilogue
  6. Emit per-instruction OOG landing pads
  7. block_epilogue: emitFallthroughEpilogue → ExitTrampoline
  8. Finalize → resolve labels → write to ExecutableMemory
  9. cache.Put(block) + registerDispatch(block)
```
