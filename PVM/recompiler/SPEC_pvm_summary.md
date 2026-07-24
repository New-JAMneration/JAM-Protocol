# PVM Specification Summary（AI Agent Rebuild Reference）

本文件摘要 PVM（Polka Virtual Machine）的關鍵規格，供 AI agent 重建 recompiler 時參照。
完整規格見 Graypaper v0.7.2 Appendix A。

**Interpreter Reference**：每條 instruction 的精確語意可參考 `PVM/instructions.go`，
decode 邏輯在 `PVM/decode.go`，djump 解析在 `PVM/branch.go`。
這些檔案是 test vector 驗證通過的 reference implementation。

---

## 1. Registers

| Index | 名稱 | 用途 |
|-------|------|------|
| 0 | RA | Return address / accumulator |
| 1 | SP | Stack pointer |
| 2 | T0 | Temporary 0 |
| 3 | T1 | Temporary 1 |
| 4 | T2 | Temporary 2 |
| 5 | S0 | Saved 0 |
| 6 | S1 | Saved 1 |
| 7 | A0 | Argument 0 |
| 8 | A1 | Argument 1 |
| 9 | A2 | Argument 2 |
| 10 | A3 | Argument 3 |
| 11 | A4 | Argument 4 |
| 12 | A5 | Argument 5 |

- 13 個 64-bit unsigned registers
- 初始值由 `InitFromProgram` 設定（RA = 2^32 - 2^16, SP = stackEnd, A0 = argStart, A1 = argLen）

---

## 2. Memory Layout

### 2.1 常數

| 符號 | 值 | 意義 |
|------|---|------|
| ZZ | 65536 (2^16) | 區段間距 |
| ZP | 4096 | Page size |
| ZI | 16 | Initial gap |
| ZA | 2 | Jump table address alignment |
| Z(n) | `((n + ZP - 1) / ZP) * ZP` | Page-aligned ceiling |
| P(n) | `Z(n)` | 同上，另一寫法 |

### 2.2 Segment Layout（4GB address space）

```
Address 0x00000000 ────────────────────────────────────────────

  [0, ZZ)                              — 低保護區（PROT_NONE）

  readOnlyStart  = ZZ
  readOnlyEnd    = ZZ + len(code)
  readOnlyPad    = ZZ + P(len(code))   — Read-only segment（program code/data）

  readWriteStart = 2*ZZ + Z(len(code))
  readWriteEnd   = readWriteStart + len(data)
  readWritePad   = readWriteStart + P(len(data)) + z_pages * ZP
                                        — Read-Write segment（初始化 data + z-page）

  heapStart      = readWritePad        — Heap 起始（sbrk 可擴展到 stackStart）
  ...（PROT_NONE，sbrk 時 mprotect 啟用）

  stackStart     = 2^32 - 2*ZZ - ZI - P(stackSize)
  stackEnd       = 2^32 - 2*ZZ - ZI   — Stack（RW）

  argStart       = 2^32 - ZZ - ZI
  argEnd         = argStart + len(argument)
  argPad         = argStart + P(len(argument))  — Argument（Read-only）

Address 0xFFFFFFFF ────────────────────────────────────────────
```

### 2.3 Special Addresses

| 位址 | 意義 |
|------|------|
| 0x00000000–0x0000FFFF | Null trap zone（access → Panic） |
| 0xFFFF0000 | HALT sentinel（jump_ind 跳到此 = HALT） |

---

## 3. Instruction Encoding

### 3.1 Program Blob 結構

```
StandardCodeFormat = jumpTable || instructions || bitmasks
```

- **jumpTable**：variable-length entries（每 entry 1-8 bytes），儲存 indirect jump 目標 PC
- **instructions**：sequential opcode stream，每條 1-17 bytes
- **bitmasks**：per-PC byte，`0x03` = basic block start

### 3.2 Instruction Format

```
[opcode: 1 byte] [operands: 0-16 bytes]
```

Operand encoding 依 `InstrCategory` 決定：

| Category | Format | 例 |
|----------|--------|---|
| NoArgs | `[opcode]` | trap, fallthrough |
| OneImm | `[opcode][imm: 1-lX bytes]` | ecalli |
| LoadImm64 | `[opcode][reg_byte][imm: 8 bytes]` | load_imm_64 |
| TwoImm | `[opcode][lX_byte][vX: lX bytes][vY: lY bytes]` | store_imm_u8 |
| OneOffset | `[opcode][offset: lX bytes]` | jump |
| OneRegImm | `[opcode][reg_byte][imm: lX bytes]` | load_imm, jump_ind |
| OneRegTwoImm | `[opcode][reg_byte][vX: lX bytes][vY: lY bytes]` | store_imm_ind |
| BranchOneRegImm | `[opcode][reg_byte][vX: lX bytes][offset: lY bytes]` | branch_eq_imm |
| TwoReg | `[opcode][reg_byte]` | move_reg, sbrk |
| TwoRegImm | `[opcode][reg_byte][imm: lX bytes]` | add_imm_32 |
| TwoRegOffset | `[opcode][reg_byte][offset: lX bytes]` | branch_eq |
| TwoRegTwoImm | `[opcode][reg_byte][lX_byte][vX: lX bytes][vY: lY bytes]` | load_imm_jump_ind |
| ThreeReg | `[opcode][reg_byte][rD_byte]` | add_64, div_u_64 |

### 3.3 Decode 公式（精確）

**reg_byte 解析**：
```
rA = min(12, instructionCode[pc+1] % 16)     // 低 4 bits
rB = min(12, instructionCode[pc+1] >> 4)     // 高 4 bits（或 floor index）
```

**ThreeReg 的第三暫存器**：
```
rD = min(12, instructionCode[pc+2])          // 第三 byte
```

**Immediate 長度計算**（`skipLength` = instruction 總長 - 1，即 opcode 後的 byte 數）：
```
OneImm:         lX = min(4, skipLength)
OneRegImm:      lX = min(4, max(0, skipLength - 1))
TwoRegImm:      lX = min(4, max(0, skipLength - 1))
OneOffset:      lX = min(4, skipLength)
TwoRegOffset:   lX = min(4, max(0, skipLength - 1))
BranchOneRegImm: lX = min(4, reg_byte>>4 % 8),  lY = min(4, max(0, skipLength - lX - 1))
TwoImm:         lX = min(4, instructionCode[pc+1]),  lY = min(4, max(0, skipLength - lX - 1))
OneRegTwoImm:   lX = min(4, reg_byte>>4),  lY = min(4, max(0, skipLength - lX - 1))
TwoRegTwoImm:   lX = min(4, instructionCode[pc+2] % 8),  lY = min(4, max(0, skipLength - lX - 2))
```

**Immediate 符號擴展**：
- 大多數 immediate 做 **sign-extend**（`ReadUintSignExtended`）
- `TwoRegTwoImm`（opcode 180）的 vX, vY 做 **zero-extend**（`ReadUintFixed`）
- `LoadImm64` 的 imm 是原始 8 bytes little-endian（`DeserializeFixedLength`）

**Offset 語意**：
```
targetPC = pc + offset    // offset 是 signed，相對於「當前 PC」
```

### 3.4 Block Terminators

以下 opcode 結束一個 basic block（由 `IsBlockTerminator()` 查表判定）：
- `trap`（opcode 0）
- `fallthrough`（opcode 1）— block 邊界標記，不是真的跳，fallthrough 到下一 block
- `jump`（opcode 40）
- `jump_ind`（opcode 50）
- `load_imm_jump`（opcode 80）
- `branch_*_imm`（opcode 81-90）
- `branch_*`（opcode 170-175）
- `load_imm_jump_ind`（opcode 180）

非 terminator 的指令在 block 內 sequential 執行。

---

## 4. Gas Model (v0.7.2)

- 每條指令 cost = 1 gas
- Gas < 0 → ExitOOG（在指令執行**前**檢查）
- Graypaper v0.8.0 改為 per-block pipeline simulation（尚未實作）

---

## 5. Exit Reasons

ExitReason 是 uint64，type 編碼在**最高 byte**（bit 63-56）：

```
ExitReason = ExitReasonType(type) << 56 | payload
```

| Type | 值 (type << 56) | payload | 觸發條件 |
|------|-----------------|---------|---------|
| CONTINUE | 0x00_00000000000000 | 無 | block fallthrough（內部用） |
| HALT | 0x01_00000000000000 | 無 | jump_ind 到 0xFFFF0000 |
| PANIC | 0x02_00000000000000 | 無 | trap / invalid jump target / signal |
| OUT_OF_GAS | 0x03_00000000000000 | 無 | Gas < 0 |
| PAGE_FAULT | 0x04_000000XXXXXXXX | uint32 fault addr | 存取 PROT_NONE 頁面 |
| HOST_CALL | 0x05_0000000000XX | uint8 callID | ecalli 指令 |

```go
func (e ExitReason) GetReasonType() ExitReasonType { return ExitReasonType(e >> 56) }
func (e ExitReason) GetHostCallID() uint8           { return uint8(e) }
func (e ExitReason) GetPageFaultAddress() uint32    { return uint32(e) }
```

注意：recompiler 內部 control region 的 ExitReason 用簡化編碼（見 SPEC_architecture.md），
與 PVM package 的 `ExitReason` 在 `host.go` 做轉換。

---

## 6. Jump Table & Bitmask

### Jump Table

- Entry 數量 = `jumpTable.Size`
- Entry byte 長度 = `jumpTable.Length`（1-8）
- `jumpTable.Data` = sequential entries，每 entry 存一個 PC 值
- 合法 jump address = `[2, tableSize * ZA]` 且 aligned to ZA=2

### Djump 解析（jump_ind 語意）

```
targetAddr = uint32(Reg[rA] + vX)

if targetAddr == 0xFFFF0000 → HALT
if targetAddr == 0 → PANIC
if targetAddr % ZA != 0 → PANIC
if targetAddr > tableSize * ZA → PANIC

index = (targetAddr / ZA) - 1
destPC = jumpTable[index]      // read tableLen bytes, little-endian

if destPC >= len(bitmasks) → PANIC
if bitmasks[destPC] != 0x03 → PANIC (not a basic block start)

jump to destPC
```

### load_imm_jump_ind（opcode 180）特殊語意

```
decode: rA, rB, vX, vY
dest = uint32(Reg[rB] + vY)
reason, newPC = djump(dest)

Reg[rA] = vX                  // ⚠️ register update 在 djump 結果之前執行！
                               //    即使 djump → PANIC，Reg[rA] 仍然被修改

if reason == PANIC → return PANIC
if reason == HALT  → return HALT
else               → jump to newPC
```

這是 test vector 驗證的行為（參見 interpreter 的 `instLoadImmJumpInd`）。

---

## 7. sbrk 語意

```
sbrk(rD, rA):
    amount = Reg[rA]
    oldHP = heapPointer

    if amount == 0:
        Reg[rD] = oldHP
        return

    newHP = oldHP + amount
    if newHP < oldHP (overflow) or newHP > heapLimit:
        Reg[rD] = 0
        return

    // activate pages from oldHP to newHP
    for page in [oldHP, pageCeil(newHP)):
        mprotect(page, PROT_READ|PROT_WRITE)

    heapPointer = newHP
    Reg[rD] = newHP
```

---

## 8. Memory Access 語意

### 地址驗證（A.8 / A.9）

```
invalidAddresses = []
for each read/write access at addr:
    pageNum = addr / ZP
    if !readable[pageNum] (read) or !writeable[pageNum] (write):
        invalidAddresses.append(addr)

if any invalidAddress < ZZ (2^16):
    return ExitPanic                        // null zone access → Panic (NOT PageFault)
else if invalidAddresses is not empty:
    return ExitPageFault | min(invalidAddresses)  // 回報最小的 fault 地址
else:
    access is valid
```

### Recompiler 的實作方式

Recompiler 不做軟體 bounds check，而是依賴硬體：
- `[0, ZZ)` 區域已 mprotect(PROT_NONE) → access 觸發 SIGSEGV
- Signal handler 判斷 fault address：
  - `< ZZ` → ExitPanic
  - `>= ZZ` → ExitPageFault | faultAddr

### Store/Load 計算

```
addr = uint32(Reg[rA] + offset)     // 32-bit wrap-around
// recompiler: MOV ECX, rA; ADD ECX, offset (32-bit ADD wraps automatically)
// access: [R15 + ECX] with appropriate width
```

---

## 9. Host Call (ecalli) 語意

```
ecalli N:
    exit to host with operation = N
    host executes omega(N) which may:
      - read/write guest memory
      - modify registers
      - modify gas
      - return ExitContinue (resume) or other (terminate)
    if ExitContinue:
        PC = next instruction after ecalli
        continue execution
```

---

## 10. Arithmetic 語意摘要

### 32-bit 版（opcode 190-199 等）

- 操作在 uint32 上進行
- 結果做 `sext_4`（sign-extend 32→64）
- `div_u_32` div-by-zero → result = 2^64 - 1
- `div_s_32` div-by-zero → result = 2^64 - 1; INT32_MIN / -1 → INT32_MIN
- `rem_u_32` div-by-zero → result = dividend
- `rem_s_32` div-by-zero → result = dividend; x % -1 → 0

### 64-bit 版（opcode 200-209 等）

- 操作在 uint64 上進行
- 相同的 div-by-zero / overflow 規則（但用 2^64-1 / INT64_MIN）

### mul_upper 三種變體

```
mul_upper_u_u (opcode 214):
    hi, _ = bits.Mul64(Reg[rA], Reg[rB])    // unsigned × unsigned
    Reg[rD] = hi

mul_upper_s_s (opcode 213):
    hi, _ = bits.Mul64(abs(int64(rA)), abs(int64(rB)))
    if (rA < 0) == (rB < 0):  Reg[rD] = hi
    else:                      Reg[rD] = -hi

mul_upper_s_u (opcode 215):
    signedA = int64(Reg[rA])
    hi, lo = bits.Mul64(abs(signedA), Reg[rB])   // |signed| × unsigned
    if signedA < 0:
        hi = -hi
        if lo != 0: hi--    // 2's complement borrow
    Reg[rD] = hi
```

### 其他運算

- shift amount 自動 mask：32-bit 版 %32，64-bit 版 %64
- `set_lt_u` / `set_lt_s`: result = (a < b) ? 1 : 0
- `cmov_iz(rA, rB, rD)`: if Reg[rB] == 0 then Reg[rD] = Reg[rA]
- `cmov_nz(rA, rB, rD)`: if Reg[rB] != 0 then Reg[rD] = Reg[rA]
- `cmov_iz_imm(rA, rB, vX)`: if Reg[rB] == 0 then Reg[rA] = vX
- `cmov_nz_imm(rA, rB, vX)`: if Reg[rB] != 0 then Reg[rA] = vX
- `and_inv(rA, rB, rD)`: Reg[rD] = Reg[rA] & ^Reg[rB]（AND NOT）
- `or_inv(rA, rB, rD)`: Reg[rD] = Reg[rA] | ^Reg[rB]（OR NOT）
- `xnor(rA, rB, rD)`: Reg[rD] = ^(Reg[rA] ^ Reg[rB])
- `max(rA, rB, rD)`: Reg[rD] = max(int64(rA), int64(rB))（signed）
- `max_u(rA, rB, rD)`: Reg[rD] = max(uint64(rA), uint64(rB))（unsigned）
- `min(rA, rB, rD)`: Reg[rD] = min(int64(rA), int64(rB))（signed）
- `min_u(rA, rB, rD)`: Reg[rD] = min(uint64(rA), uint64(rB))（unsigned）
- `rot_l_64/32(rA, rB, rD)`: Reg[rD] = RotateLeft(Reg[rA], Reg[rB] % 64/32)
- `rot_r_64/32(rA, rB, rD)`: Reg[rD] = RotateRight(Reg[rA], Reg[rB] % 64/32)
