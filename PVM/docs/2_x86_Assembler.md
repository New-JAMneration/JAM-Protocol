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

這就是 recompiler 讀取 Gas 的指令：`MOV RCX, [R15-48]`。

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

## 3. Assembler 架構（`PVM/asm/` package）

### 3.1 分層

```
Assembler (高階 API)
  ├── 呼叫 emitREX / emitMemOp / modRM / sib 等 encoding helpers
  └── 寫入 CodeBuffer

CodeBuffer (低階 byte buffer + label 管理)
  ├── data []byte          — 累積的機器碼
  ├── labels map[string]int — label name → offset
  └── fixups []fixup       — 待回填的前向參照
```

### 3.2 CodeBuffer

```go
type CodeBuffer struct {
    data   []byte
    labels map[string]int
    fixups []fixup        // {label, offset, size(1 or 4)}
}
```

關鍵操作：
- `Emit(bytes...)` — append raw bytes
- `BindLabel(name)` — 記錄 label 的 offset
- `UseLabel32(name)` — emit 4-byte placeholder + 註冊 fixup
- `ResolveFixups()` — 回填所有 `rel = target - (fixupPos + fixupSize)`

### 3.3 Label 與前向參照

Compiler 是 **single-pass**，遇到跳轉目標可能還沒 emit（forward reference）：

```
emitGasCheck → Jcc("out_of_gas_pc_00001234")  ← 目標尚未存在
...（更多指令）...
BindLabel("out_of_gas_pc_00001234")            ← 現在存在了
...
Finalize() → ResolveFixups()                   ← 回填所有 placeholder
```

`Jcc` 會 emit `0F 8x [placeholder_4bytes]`，`ResolveFixups` 最後算出相對距離填回去。

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

### 4.1 每條 PVM 指令的 emit 流程

```go
// compiler.go 主迴圈
for i := range instrs {
    instr := &instrs[i]
    c.emitGasCheck(a, instr.PC)           // 4 條 x86（load/test/jcc/sub）
    handler := opcodeHandlers[instr.Opcode]
    handler(c, a, instr)                   // PVM opcode → x86 序列
}
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
| `emit_two_reg.go` | 兩 reg 特殊操作（move_reg、sbrk、clz、ctz、popcnt、bswap） |
| `emit_branch.go` | branch（條件跳轉）、djump（indirect jump） |
| `gas.go` | per-instruction / block-based gas check emit |
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

### 4.5 Gas Check emit（GP v0.7.2 per-instruction）

每條 PVM 指令前插入 4 條 x86：

```asm
MOV RCX, [R15 - 48]           // 讀 Gas
TEST RCX, RCX                  // 檢查 < 0
JS out_of_gas_pc_XXXXXXXX      // 負數 → 跳 OOG exit
SUB qword [R15 - 48], 1        // 扣 1 gas
```

Block epilogue 再 emit 每指令的 OOG landing pad：

```asm
out_of_gas_pc_XXXXXXXX:
  MOV dword [R15-32], pc       // 設 ExitPC
  MOV RCX, ExitOOG             // 設 ExitReason
  MOV [R15-40], RCX
  JMP exit_trampoline
```

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
| `PVM/asm/registers.go` | Register type + ConditionCode 定義 |
| `PVM/asm/encoding.go` | REX / ModR/M / SIB encoding helpers |
| `PVM/asm/buffer.go` | CodeBuffer：byte emit + label + fixup |
| `PVM/asm/assembler.go` | Assembler：high-level API wrapper |
| `PVM/asm/instructions.go` | 所有 x86 指令 emit 方法（MOV/ADD/JMP/...） |
| `PVM/recompiler/compiler.go` | CompileBasicBlock + opcodeHandlers dispatch |
| `PVM/recompiler/emit_*.go` | 各類 PVM opcode 的翻譯實作 |
