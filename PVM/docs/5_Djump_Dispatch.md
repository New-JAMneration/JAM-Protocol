# djump Native Dispatch

本章說明 PVM indirect jump（`jump_ind` / `load_imm_jump_ind`）如何在 native code 內完成位址解析，避免回到 Go。

前置閱讀：`1_Recompiler_Workflow.md`（djump miss）、`4_HostCall_Integration.md`（MachineInvoke 內部消化）。

---

## 1. 什麼是 djump

PVM 的 `jump_ind`（indirect jump）和 `load_imm_jump_ind` 的目標位址在 **runtime 才能確定**（register 中的值）。這跟 static jump/branch 不同——static target 在 compile time 已知，可以直接 link。

djump 需要：
1. 驗證目標位址合法（alignment、bounds、是 basic-block start）
2. 查詢 jump table 拿到實際 PC
3. 跳到對應的 compiled native code

如果每次都回 Go 做這些 → 開銷太大（djump 在 loop 中很常見）。

---

## 2. 設計目標：三條出口

| 路徑 | 條件 | 動作 | 離開 native？ |
|------|------|------|-------------|
| **hit** | dispatch table 有 native addr | `JMP reg`（直接跳） | 否 |
| **miss** | 合法 PC 但尚未 compile | exit CONTINUE + resolvedPC → Go compile → 回 native | 是（短暫） |
| **panic** | 非法位址（alignment/bounds/非 block start） | ExitPanic | 是 |

目標：**大多數 djump 走 hit 路徑，完全不離開 native**。

---

## 3. 資料結構

### 3.1 djumpSupport

```go
type djumpSupport struct {
    tableLen   uint32      // jump table entry 的 byte 長度（1/2/3/4/5/6/7/8）
    tableSize  uint32      // jump table 的 entry 數量
    maxAddr    uint32      // tableSize * ZA（合法位址上界）
    bitmaskLen uint32      // bitmask 長度 = program code 長度
    dispatch   []uintptr   // PC → native address 的 dispatch table
}
```

### 3.2 三塊記憶體

| 資料 | 儲存位置 | 用途 |
|------|---------|------|
| Jump Table（rodata） | ExecutableMemory（`em.Write`） | 位址 → PC 的 lookup 表 |
| Bitmask（rodata） | ExecutableMemory（緊接 jump table） | `bitmask[pc] == 0x03` 判斷是否為 basic-block start |
| Dispatch Table | Go heap（`[]uintptr`） | PC → native addr，由 `registerDispatch` 填入 |

前兩者放在 ExecutableMemory 是因為 native code 需要直接讀取（`[R15 - offset]` 指標取得）。

### 3.3 Control Region 指標

```
[R15 - 176]  OffsetDjumpTable      → jump table rodata 起始位址
[R15 - 184]  OffsetDjumpBitmask    → bitmask rodata 起始位址
[R15 - 192]  OffsetDjumpDispatch   → dispatch table (Go heap) 起始位址
```

這三個值在 `ensureDjumpSupport()` 時寫入 control region，native code 透過 `MOV reg, [R15-offset]` 讀取。

---

## 4. 初始化流程

```go
func (c *Compiler) ensureDjumpSupport() error {
    // 1. 把 jump table + bitmask 寫入 ExecutableMemory 作為 rodata
    blob := jumpTable.Data + bitmasks
    offset, _ := em.Write(blob)

    // 2. 建立 dispatch table（Go heap，長度 = bitmask 長度 = code 長度）
    dispatch := make([]uintptr, len(bitmasks))

    // 3. 把三個指標寫入 control region
    ctx.setDjumpPointers(tableAddr, bitmaskAddr, &dispatch[0])
}
```

每次 compile 一個 block 時，呼叫 `registerDispatch(block)` 把 `block.NativeAddr` 填入 dispatch table：

```go
func (c *Compiler) registerDispatch(block *CompiledBlock) {
    pc := int(block.PVMStartPC)
    if pc >= 0 && pc < len(c.djump.dispatch) {
        c.djump.dispatch[pc] = block.NativeAddr
    }
}
```

---

## 5. Native Code 解析流程（emitDjumpNative）

`emitDjumpNative` emit 的 x86 序列大約 50-70 bytes，在 native 內完成完整的 djump 解析。

### 5.1 驗證階段（Panic 檢查）

```
targetAddr = register 值（32-bit PVM 位址）

1. targetAddr == 0 ?                     → Panic（null jump）
2. targetAddr & 1 != 0 ?                 → Panic（未對齊，ZA=2）
3. targetAddr > jumpTable.Size * ZA ?    → Panic（超出 jump table 範圍）
4. index = (targetAddr >> 1) - 1
5. index >= tableSize ?                  → Panic
```

### 5.2 Jump Table Lookup

```
6. byteOffset = index * tableLen
7. tablePtr = [R15 - OffsetDjumpTable] + byteOffset
8. destPC = readFixedWidth(tablePtr, tableLen)    // 1-8 bytes little-endian
```

`emitLoadJumpEntry` 處理不同 entry 長度（1/2/4/8 bytes 走單條 MOV，3/5/6/7 走逐 byte 組裝）。

### 5.3 Bitmask 驗證

```
9.  destPC >= bitmaskLen ?                → Panic
10. bitmaskPtr = [R15 - OffsetDjumpBitmask] + destPC
11. bitmask[destPC] != 0x03 ?             → Panic（不是 basic-block start）
```

`0x03` 表示該 PC 是一個 basic block 的起始位置（Graypaper 定義）。

### 5.4 Dispatch

```
12. dispatchPtr = [R15 - OffsetDjumpDispatch] + destPC * 8
13. nativeAddr = *(uint64*)dispatchPtr
14. nativeAddr == 0 ?                     → Miss（尚未 compile）
15. JMP nativeAddr                        → Hit！直接跳過去
```

### 5.5 流程圖

```
targetAddr (register)
    │
    ├── == 0 ?                      → Panic
    ├── & 1 != 0 ?                  → Panic
    ├── > maxAddr ?                 → Panic
    │
    ├── index = (addr >> 1) - 1
    ├── index >= tableSize ?        → Panic
    │
    ├── destPC = jumpTable[index]
    ├── destPC >= bitmaskLen ?      → Panic
    ├── bitmask[destPC] != 0x03 ?   → Panic
    │
    ├── nativeAddr = dispatch[destPC]
    ├── nativeAddr == 0 ?           → Miss（exit to Go, compile, re-enter）
    │
    └── JMP nativeAddr              → Hit（完全不離開 native）
```

---

## 6. Miss 路徑（回 Go compile）

```go
// emitDjumpMiss — native 側
MOV [R15 - ExitPC], destPC       // 告訴 Go 要跳到哪
MOV qword [R15 - ExitReason], 0  // CONTINUE（= 0）
JMP exit_trampoline
```

Go 側 `BlockBasedInvoke` 收到 `DjumpCallID`：

```go
// recompiler.go
if exitReason.GetHostCallID() == DjumpCallID {
    exitReason, pc = r.resolveDjump(blockStartPC, exitPC)
    // resolveDjump 用 Go 側 DjumpResolve 再驗一次
    // 回傳 CONTINUE + resolved PC → continue loop
    // → lookupOrCompileBlock(resolvedPC) → 下次 djump hit
}
```

**下次**同樣的 djump 到相同 destPC → dispatch table 已填好 → hit → 不再回 Go。

---

## 7. Panic 路徑

```go
// emitDjumpPanic — native 側
MOV RCX, ExitPanic
MOV [R15 - ExitReason], RCX
MOV dword [R15 - ExitPC], instrPC    // djump 指令本身的 PC
JMP exit_trampoline
```

Go 側收到 PANIC → 程式結束。

---

## 8. 空 Jump Table Fallback

如果 program 的 jump table 為空（`jt.Length == 0`）：

```go
if jt.Length == 0 || jt.Size == 0 {
    emitDjumpExit(a, targetReg)    // 直接走 Go-side DjumpResolve（legacy 路徑）
    return nil
}
```

這是少見的退化情況（program 沒有 jump table），走舊版的 Go 全解析路徑。

---

## 9. 效能特性

| 指標 | 說明 |
|------|------|
| Hit 路徑指令數 | ~30-40 條 x86（含所有 validation） |
| Miss 次數 | 只在首次遇到新 destPC 時 miss，之後永遠 hit |
| 典型 loop | 第一次 miss，之後全 hit → 等同 native indirect jump |
| Go-side resolve 計數 | 實測接近 0（PERFORMANCE.md `djump=0`） |

djump native dispatch 是 recompiler 能跑 tight loop 的關鍵：如果每次 indirect jump 都回 Go，loop 效能會直接打折。

---

## 10. 與 Block Linking 的對比

| | Block Linking（static） | Djump Dispatch（dynamic） |
|---|---|---|
| 目標 PC | compile time 已知 | runtime 才知道 |
| 解析方式 | emit `JMP rel32` 直接跳 | emit lookup + validate + `JMP reg` |
| Miss 處理 | 遞迴 compile（`compileForLink`） | exit Go → compile → registerDispatch |
| 適用指令 | `jump`、`branch_*` | `jump_ind`、`load_imm_jump_ind` |

兩者互補：static 走 linking，dynamic 走 dispatch table。

---

## 相關檔案索引

| 檔案 | 職責 |
|------|------|
| `PVM/recompiler/djump_native.go` | `djumpSupport`、`ensureDjumpSupport`、`emitDjumpNative`、dispatch 管理 |
| `PVM/recompiler/recompiler.go` | `resolveDjump`（miss 時 Go-side 解析） |
| `PVM/recompiler/context.go` | `OffsetDjumpTable` / `OffsetDjumpBitmask` / `OffsetDjumpDispatch` |
| `PVM/recompiler/compiler.go` | `registerDispatch`（每次 compile 後更新 dispatch table） |
| `PVM/recompiler/emit_branch.go` | `emitJumpInd` / `emitLoadImmJumpInd`（呼叫 `emitDjumpNative`） |
