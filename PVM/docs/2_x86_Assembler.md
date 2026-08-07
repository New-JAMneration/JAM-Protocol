# x86-64 Assembler & Code Emission

本章說明 recompiler 如何將 PVM 指令翻譯為 x86-64 機器碼。

前置閱讀：`1_Recompiler_Workflow.md`（Decode → Translate 的 black box）。

---

## 1. x86-64 指令格式

每條 x86-64 指令格式最多由以下部分組成：

```
[(Legacy) Prefixes] + [REX Prefix] + [Opcode] + [ModR/M] + [SIB] + [Displacement] + [Immediate]
```

### 1.1 REX Prefix

x86 原本只有 8 個 registers（RAX–RDI）。AMD64 擴充到 16 個（R8–R15），用 REX byte 告訴 CPU：

```
0100 W R X B
     │ │ │ └── extends ModR/M.rm 或 SIB.base（R8–R15）
     │ │ └──── extends SIB.index（R8–R15）
     │ └────── extends ModR/M.reg（R8–R15）
     └──────── 1 = 64-bit operand size
```

- 用到 R8–R15 **或** 需要 64-bit 操作 → 必須有 REX
- RAX–RDI 且 32-bit 操作 → 不需要 REX（省 1 byte）
- 這就是為什麼 register map 把高頻 PVM register 放非 REX register（RAX, RDX, RBX, RSI, RDI）

### 1.2 ModR/M（告訴 CPU operand 在哪）

```
  7  6 | 5  4  3 | 2  1  0
 [mod ] [ reg   ] [ rm    ]
```

| 欄位 | bits | 作用 |
|------|------|------|
| mod | 7:6 | 定址模式：11=reg-reg、01=mem+disp8、10=mem+disp32、00=mem |
| reg | 5:3 | 第一個 operand（register 低 3 bits） |
| rm | 2:0 | 第二個 operand（register 或 memory base 低 3 bits） |

### 1.3 SIB（複雜定址：base + index × scale）

當 ModR/M 的 `rm = 100` 時，表示後面跟著 SIB byte：

```
  7  6 | 5  4  3 | 2  1  0
[scale] [ index ] [ base  ]
```

用途：`[base + index × scale + disp]`（例如陣列存取）。

> 特例：RSP（Lo3=4=100）作為 base 時必須走 SIB（因為 rm=100 被 SIB 佔了）。

### 1.4 Displacement

- **disp8**：-128 ~ +127，1 byte（ModR/M mod=01）
- **disp32**：-2G ~ +2G，4 bytes（ModR/M mod=10）
- Control region 的 Gas（offset -48）在 disp8 範圍內 → 省 3 bytes

---

## 2. 實際範例

### 例 1：`MOV RAX, RBX`（register ← register）

```
opcode: 8B (MOV r64, r/m64)
ModR/M: mod=11, reg=RAX(000), rm=RBX(011) → 0xC3

完整: 48 8B C3
      │  │  └── ModR/M
      │  └───── opcode
      └──────── REX.W (64-bit)
```

### 例 2：`MOV RAX, [R15 - 48]`（register ← memory + disp8）

```
opcode: 8B
ModR/M: mod=01, reg=RAX(000), rm=R15.Lo3(111) → 0x47
disp8: -48 = 0xD0

完整: 49 8B 47 D0
      │        └── disp8
      └──────── REX.W + REX.B (R15 是 extended)
```

這是 recompiler 讀 control region 欄位的典型形式（例如讀 Gas 到 scratch；現行 gas check 已融合為直接 `SUB qword [R15-48], 1`，見 §4.5）。

### 例 3：`MOV [R15 - 48], RBX`（memory ← register）

```
opcode: 89 (注意方向反了：MOV r/m64, r64)
ModR/M: mod=01, reg=RBX(011), rm=R15.Lo3(111) → 0x5F

完整: 49 89 5F D0
```

### 例 4：`MOV RAX, [RSP + 8]`（base 是 RSP → 需要 SIB）

```
ModR/M: mod=01, reg=RAX(000), rm=100(SIB!) → 0x44
SIB: scale=00, index=RSP(100=none), base=RSP(100) → 0x24

完整: 48 8B 44 24 08
           │  │  └── disp8=8
           │  └───── SIB
           └──────── ModR/M
```

---

## 3. Assembler 架構（`PVM/recompiler/asm/` package）

### 3.1 分層

```
Assembler (高階 API)
  ├── 呼叫 emitREX / emitMemOp / modRM / sib 等 encoding helpers
  └── 寫入 CodeBuffer

CodeBuffer (低階 byte buffer + label 管理)
  ├── data []byte          — 累積的機器碼
  ├── labelPos []int32     — Label handle → offset（-1 = 未 bind）
  └── fixups []fixup       — 待回填的前向參照
```

### 3.2 CodeBuffer

```go
type CodeBuffer struct {
    data     []byte
    labelPos []int32   // index = Label；值 = byte offset；-1 = 未 bind
    fixups   []fixup   // {label Label, offset, size(1 or 4)}
}
```

關鍵操作：
- `Emit(bytes...)` — append raw bytes
- `NewLabel()` — 配發一個未 bind 的整數 handle
- `BindLabel(l)` — 記錄 label 的 offset
- `UseLabel32(l)` — emit 4-byte placeholder + 註冊 fixup
- `ResolveFixups()` — 回填所有 `rel = target - (fixupPos + fixupSize)`

### 3.3 Label 與前向參照

Compiler 是 **single-pass**，遇到跳轉目標可能還沒 emit（forward reference）：

```
blockOOG := a.NewLabel()
emitBlockGasCheck → Jcc(blockOOG)   ← 目標尚未存在（unbound handle）
...（block 內指令 emit）...
BindLabel(blockOOG)                 ← OOG landing pad
...
Finalize() → ResolveFixups()        ← 回填所有 placeholder
```

`Jcc` 會 emit `0F 8x [placeholder_4bytes]`，`ResolveFixups` 最後算出相對距離填回去。

### 3.3.1 Label handle 化（已實作）

早期版本 label 是字串：每個 call site 用 `fmt.Sprintf` 造名（gas check 每條 PVM 指令 2 次：Jcc 端 + BindLabel 端，各一次 heap alloc），`labels` 是 `map[string]int`（Bind insert、Finalize resolve 各 hash 一次字串）。這些全發生在 compile 階段，emitted bytes 不受影響。**實測**：handle 化後 conformance 的 compile 桶 70ms → 37ms（avg 15.8µs → 8.4µs / block，-47%）；run 桶不變。

**核心**：label 是 assembler 發放的整數 handle，查找是 slice 索引：

```go
type Label int32

// CodeBuffer:
labelPos []int32   // index = Label；值 = byte offset；-1 = 未 bind
fixups   []fixup   // {label Label, pos int32}（原本存 string）

NewLabel() Label    // append(labelPos, -1)，O(1) 零 alloc
BindLabel(l Label)  // labelPos[l] = len(data)
Jcc(cc, l Label)    // fixups append {l, 洞位置}
// Finalize：labelPos[f.label] < 0 → unbound error；否則回填 rel32
// Reset：labelPos = labelPos[:0]（保留 capacity，跨 block 重用）
```

**管理方式**——原字串名承擔三種不同的溝通方式，各有對應：

| 類型 | 舊字串做法 | 現行做法 |
|------|------|------|
| ① 區域 label（最大宗）：taken、halt、djump_miss/panic、chain_miss、div 系列 | 名字編入 PC 保唯一 | 產生與使用在同一 emit 函式內 → `l := a.NewLabel()` 區域變數，`NewLabel` 天生唯一，**不需任何註冊** |
| ② 跨 emit 函式、per-block 共用：exit trampoline | 各 emit 函式用字串約定 `"exit_trampoline"` | Assembler 欄位，`Reset()` 時預配，call site 用 `a.Jmp(a.ExitTrampoline())`（return_label 只在 `EmitEntryTrampoline` 內自產自用，屬 ①） |
| ③ block 入口 OOG landing pad | hot 端先引用、cold 端後 bind | block 開頭 `emitBlockGasCheck` 引用 `blockOOG`；指令 loop 後 `emitBlockOutOfGasExit` bind 同一 label |

**成本對照**（每個 label 引用點）：

| 步驟 | string 版（舊） | int 版（現行） |
|------|-----------|--------|
| 造名 | `Sprintf` ×2（heap alloc ×2） | slice 取值，~0 |
| Bind | map insert（hash 整條字串） | `labelPos[l] = off`，一次 store |
| Finalize resolve | map lookup（再 hash 一次） | `labelPos[f.label]`，一次 load |
| fixup 大小 | 24B（string ptr+len+pos） | 8B |
| Reset | `clear(map)` 走遍 bucket | `[:0]`，~0 |

**注意**：錯誤訊息以數字 id 呈現（`unresolved label 3`）；handle 只在同一個 Reset 週期內有效。

### 3.4 Assembler 高階 API（摘錄）

| 方法 | x86 指令 | 用途舉例 |
|------|----------|---------|
| `MovRegToReg(dst, src)` | `MOV r64, r64` | PVM `move_reg` |
| `MovMemToReg(dst, base, disp)` | `MOV r64, [base+disp]` | 讀 Gas / 讀 control region |
| `MovRegToMem(base, disp, src)` | `MOV [base+disp], r64` | 存 PVM reg 回 control region |
| `MovImm64ToReg(dst, imm)` | `MOV r64, imm64` | PVM `load_imm_64` |
| `AddRegReg(dst, src)` | `ADD r64, r64` | PVM `add` |
| `SubMemImm32(base, disp, imm)` | `SUB qword [base+disp], imm32` | Gas 扣除 |
| `Jcc(cond, label)` | `Jcc rel32` | branch / OOG check |
| `Jmp(label)` | `JMP rel32` | block epilogue / linking |
| `JmpReg(reg)` | `JMP r64` | djump dispatch hit |

---

## 4. Emit 設計模式（Compiler 層）

### 4.1 Block emit 流程

```go
// compiler.go（GP 0.8.0 block gas）
blockOOG := a.NewLabel()
blockGas := c.blockGasCostAt(startPC)   // A.9 GasCostFromPC / GasCostForBlock
c.emitBlockGasCheck(a, blockOOG, blockGas)

for i := range instrs {
    instr := &instrs[i]
    if i == len(instrs)-1 && IsBlockTerminator(instr.Opcode) {
        emitGasCharged(a, false)       // 離開 block 重置 gaschargedflag
    }
    handler := opcodeHandlers[instr.Opcode]
    handler(c, a, instr)               // PVM opcode → x86 序列
}
emitBlockOutOfGasExit(a, blockOOG, blockMeta.StartPC, blockGas)
```

### 4.2 opcodeHandlers dispatch table

```go
var opcodeHandlers [231]opcodeHandler

// 4.3 No-argument
opcodeHandlers[0] = (*Compiler).emitTrap       // trap → ExitPanic
opcodeHandlers[1] = (*Compiler).emitFallthrough // nop

// 4.4 Immediate
opcodeHandlers[10] = (*Compiler).emitEcalli     // host call exit
opcodeHandlers[20] = (*Compiler).emitLoadImm64  // MOV r64, imm64

// 4.5 Memory
opcodeHandlers[52] = makeLoad(1, false)         // load_u8
opcodeHandlers[59] = makeStore(1)               // store_u8

// 4.6 Arithmetic
opcodeHandlers[131] = (*Compiler).emitAddImm32  // add_imm
...
```

每個 handler 簽名統一：`func(c *Compiler, a *asm.Assembler, instr *InstrMeta) error`

### 4.3 Emit 分類

| 檔案 | 負責的 PVM 指令類別 |
|------|-------------------|
| `emit_basic.go` | trap、fallthrough、ecalli、load_imm、jump |
| `emit_memory.go` | load/store（1/2/4/8 byte，直接 / indirect） |
| `emit_arith_imm.go` | 一個 reg + 一個 imm 的算術（add_imm、sub_imm、mul_imm…） |
| `emit_arith_three.go` | 兩個 reg 的算術（add、sub、mul、div、shift、bitwise） |
| `emit_two_reg.go` | 兩 reg 特殊操作（move_reg、clz、ctz、popcnt、bswap） |
| `emit_branch.go` | branch（條件跳轉）、djump（indirect jump） |
| `gas.go` | block-level gas check emit（`emitBlockGasCheck` / `emitBlockOutOfGasExit`） |
| `emit_record_mem.go` | debug trace 的 memory access 記錄 |

### 4.4 Memory 存取的 emit 模式

PVM memory access 翻譯成 `[R15 + guest_addr]`：

```
// load_u32: Reg[dst] = *(uint32*)(guestMem + addr)
MOV ECX, addr_reg        // 取 PVM 地址（32-bit zero-extend）
MOV dst32, [R15 + RCX]   // base(R15) + index(RCX) 定址

// store_u32: *(uint32*)(guestMem + addr) = Reg[src]
MOV ECX, addr_reg
MOV [R15 + RCX], src32
```

如果地址越界（碰到 PROT_NONE page）→ 硬體 SIGSEGV → signal handler 捕獲 → ExitPageFault。

### 4.5 Gas Check emit（GP 0.8.0 block-level，A.4 / A.9）

GP 0.8.0 改為 **basic block 入口一次性 pre-charge**（`gascostforblock`），recompiler 在 compile 時用 `blockGasCostAt(startPC)` 算出成本並 bake 進 native code。mid-block resume（host call 返回等）編譯 suffix block，gas 同樣在 suffix 入口一次扣除。

Block 開頭插入 gas check（扣費與檢查融合）：

```asm
; gaschargedflag == 0 時才扣費
TEST byte [R15 - gasChargedOff], 0
JNE  charged
SUB  qword [R15 - 48], blockGas    // 扣整段 block gas
JS   block_oog                     // 結果 < 0 → OOG
MOV  byte [R15 - gasChargedOff], 1
charged:
; ... block 指令 payload ...
```

OOG landing pad（interpreter OOG 不扣費，需補回）：

```asm
block_oog:
  SUB qword [R15-48], -blockGas    // 把剛扣的 block gas 補回
  MOV dword [R15-32], blockStartPC // 設 ExitPC = block 起始 PC（A.4）
  MOV RCX, ExitOOG
  MOV [R15-40], RCX
  JMP exit_trampoline
```

> **歷史**：v0.7.2 曾用 per-instruction `SUB [gas], 1` + 每指令 OOG pad；GP 0.8.0 改 block gas 後已移除。

---

## 5. 與 recompiler 的關係總覽

```
PVM InstrMeta                          x86-64 machine code
┌─────────────┐                        ┌─────────────────┐
│ opcode: 131 │  opcodeHandlers[131]   │ REX ADD r64,imm │
│ dst: 7      │ ─────────────────────► │ (R10 += imm32)  │
│ imm: 42     │  emitAddImm32          │                 │
└─────────────┘                        └─────────────────┘
       ↑                                       │
  preDecodeBlocks                          em.Write
  (靜態 decode)                         (dual mapping)
```

---

## 相關檔案索引

| 檔案 | 職責 |
|------|------|
| `PVM/recompiler/asm/registers.go` | Register type + ConditionCode 定義 |
| `PVM/recompiler/asm/encoding.go` | REX / ModR/M / SIB encoding helpers |
| `PVM/recompiler/asm/buffer.go` | CodeBuffer：byte emit + label + fixup |
| `PVM/recompiler/asm/assembler.go` | Assembler：high-level API wrapper |
| `PVM/recompiler/asm/instructions.go` | 所有 x86 指令 emit 方法（MOV/ADD/JMP/...） |
| `PVM/recompiler/compiler.go` | CompileBasicBlock + opcodeHandlers dispatch |
| `PVM/recompiler/emit_*.go` | 各類 PVM opcode 的翻譯實作 |
