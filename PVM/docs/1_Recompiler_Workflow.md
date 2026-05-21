# Recompiler Workflow

整體流程：**Decode → Translate → Execute**。

看似三步，實際每個箭頭之間都有 black box。本章逐一拆開說明。

---

## 總覽

```
Psi_M_recompiler
  Setup（見 0_Recompiler_Setup.md）
  InitFromProgram     ← 記憶體 decode（segment + 初始 regs）
  DeBlobProgramCode   ← 指令 decode（blob → InstrMeta / BlockMeta）
    └─ preDecodeBlocks()
  寫入 control region（registers / gas / exitPC）
  NewRecompiler
  host.HostCall(pc)   ← 進入主執行迴圈
```

兩層 loop 驅動整個 workflow：

```
host.HostCall ─────────────────── 外層（≈ host call 段數，~30–40 次/invoke）
  └─ BlockBasedInvoke ────────── 內層（≈ block 數，~4000+ 次/invoke）
       lookupOrCompileBlock       ← Decode + Translate（lazy）
       executeBlockLocked         ← Execute（trampoline + native）
```

---

## 1. Decode

靜態階段，整份 program 只做一次。有兩種 decode（不要混淆）：

### Memory decode — `InitFromProgram`

- `DecodeSerializedValues(blob)` → `o/w/z/s`
- 依 Graypaper 佈局 mapSegment（read-only / read-write / stack / argument）
- 回傳初始 registers + program code（instruction bytes）

### Instruction decode — `DeBlobProgramCode`

- 解出 jump table + instruction bytes + bitmask
- 末尾 `preDecodeBlocks()` 一次掃描整份 blob，產出：
  - `InstrMeta[]`（PC, opcode, operands, skipLen）
  - `BlockMeta[]`（StartPC, EndPC, GasCost, 指令範圍）
  - `BlockAt[pc]`, `InstrIdxAt[pc]` 索引表
- 之後 runtime **不再逐 byte 解碼**

---

## 2. Decode → Translate（black box）

觸發點：`lookupOrCompileBlock(pc)` 首次遇到某 block 時觸發 `CompileBasicBlock`。

### CompileBasicBlock 流程

```
CompileBasicBlock(startPC)
  1. 從 Program.BlockAt[startPC] 取 BlockMeta + 指令切片
  2. 對每條指令：
       emitGasCheck (GP v0.7.2: load / test / sub / OOG label)
       opcodeHandlers[opcode](c, asm, instr) → emit x86 指令
  3. Block epilogue：
       fallthrough 目標已編譯 → JMP NativeAddr（block linking）
       否則 → emitExitToPC（寫 CONTINUE + 下一 PC → exit_trampoline）
  4. 每指令 OOG landing pad + EmitExitTrampoline
  5. Assembler.Finalize() → []byte（機器碼）
  6. em.Write(code) → 寫入 ExecutableMemory
  7. CodeCache.Put + registerDispatch（djump dispatch table）
```

### 寫入 ExecutableMemory（Dual Mapping，零 mprotect）

```
offset, err := em.Write(code)       // 透過 rwMem view append native code
nativeAddr := em.GetPtr(offset)     // 從 rxMem view 取得可執行位址
```

因為 dual mapping（rwMem 與 rxMem 指向同一實體頁），寫入後**立刻可執行**：
- **不需要** `MakeWritable` / `MakeExecutable`
- **不需要** 任何 `mprotect` 切換
- Compile 的成本只剩純 codegen（~7.5%），不再有 mprotect 瓶頸

> **歷史**：舊版每 compile 一個 block = 2× `mprotect` 整塊 16MB arena，實測佔全程 ~49%。
> Dual mapping 將此降為零，是目前最大的單項效能改進。

### 附加決策

- **On-demand compile**：不是整份 program 一次編完，跑到哪編到哪
- **Fallthrough linking**：epilogue 前可 eager 編下一 block（depth cap = 256）
- **Forward reference**：目標 block 尚未存在 → fallback `emitExitToPC` 回 Go loop 再查

---

## 3. Translate → Execute（black box）

### CompiledBlock 交接

Translate 產出 `CompiledBlock`：

```go
type CompiledBlock struct {
    PVMStartPC   ProgramCounter
    PVMEndPC     ProgramCounter
    NativeAddr   uintptr       // ← 唯一的「translate → execute」銜接點
    NativeOffset int
    NativeSize   int
    GasCost      int64
    InstrCount   int
}
```

`NativeAddr` = `em.GetPtr(offset)`，即 **rxMem view**（可執行映射）中的位址。Go 不再 interpret 指令，只負責：
1. 把 `NativeAddr` 傳給 `callNative`
2. 執行完讀 `ExitReason` 決定下一步

**INVARIANT**：所有可呼叫/儲存的位址（`NativeAddr`、djump dispatch entries、trampoline、signal handler fault window）必須來自 `GetPtr`（rxBase view），因為那才是真正執行的映射。

### BlockBasedInvoke 執行迴圈

```
BlockBasedInvoke(pc)                          [LockOSThread 一次, L1]
  loop:
    block = lookupOrCompileBlock(pc)          ← cache hit 或觸發 translate
    ctx.WriteExitPC(pc)
    ctx.WriteExitReason(CONTINUE)

    executeBlockLocked(block):
      trampolineAddr = getTrampolineAddr()    ← 首次 lazy emit
      SetFaultWindow(guestBase, codeStart, codeEnd)
      callNative(guestBase, block.NativeAddr, trampolineAddr)
      return ReadExitReason()

    依 exitReason 分支：
      CONTINUE       → pc = exitPC, continue
      HOST_CALL      → break → 外層 host.HostCall 跑 omega → 再 MachineInvoke
      sbrk (0xFF)    → HandleSbrk (Go mprotect) → continue
      djump miss     → compile target → continue
      HALT/PANIC/OOG/PAGE_FAULT → 結束
```

### callNative 內部

```asm
callNative(guestBase, blockAddr, trampolineAddr):
  MOVQ guestBase → RAX
  MOVQ blockAddr → RBX
  MOVQ trampoline → CX
  CALL CX

  [entry_trampoline]
    PUSH host callee-saved
    R15 = RAX (guestBase)
    存 ReturnAddr / ReturnStack 到 control region
    從 control region 載入 13 PVM regs
    JMP RCX (block native code)

  [native block 本體]
    算術/邏輯: 純 register 操作
    記憶體 load/store: [R15 + PVM_addr]
    gas: sub qword [R15-48], 1
    branch: JMP 其他 native block（若已 link）或 exit
    ecalli: 寫 ExitReason + ExitPC → JMP exit_trampoline

  [exit_trampoline]
    存 13 PVM regs 回 control region
    RSP = [R15 - ReturnStack]
    JMP [R15 - ReturnAddr] → return_label
    POP host callee-saved
    RET → 回 Go
```

### Signal Handler（硬體保護後盾）

Native code 存取 `PROT_NONE` 頁面 → CPU page fault → SIGSEGV → signal handler 攔截 → 回存 PVM state → 修改 CPU context 回到 Go → 回報 ExitPageFault 或 ExitPanic。

這讓 memory access **不需要** software bounds check（零成本），正常存取完全無 overhead。

> 完整機制（RIP/RSP 切換、ucontext 修改、Fault Window 等）見 `3_Signal_Handler.md`。

### 執行期特殊出口

| 出口 | 處理位置 | 是否離開 native loop |
|------|----------|----------------------|
| fallthrough CONTINUE | `BlockBasedInvoke` | 否 |
| block linking JMP | native 內 | 否（完全不回 Go） |
| `ecalli` | `host.HostCall` → omega | 是（外層 loop） |
| sbrk 跨頁 | `HandleSbrk`（Go mprotect） | 否（resolve 後 continue） |
| djump hit | native `JmpReg`（dispatch table） | 否 |
| djump miss | Go compile + dispatch 更新 | 否 |
| OOG / HALT / PANIC | 結束 invoke | 是 |
| PAGE_FAULT | signal handler → exit | 是 |

---

## 補充

### 三個「狀態存放處」（心智模型）

| 狀態 | 執行中在哪 | Go 怎麼讀 |
|------|-----------|-----------|
| PVM registers | x86 regs（邊界才刷 control region） | `ctx.ReadRegisters()` |
| Gas | control region `[R15-48]` | `ctx.ReadGas()` |
| Guest RAM | `guestMem[]` = `[R15+addr]` | `ctx.GuestMemory()` |

### 與 interpreter 對照

| | Interpreter | Recompiler |
|--|-------------|------------|
| 熱路徑 | Go `InstrMeta.Exec` dispatch | native x86 |
| Block 邊界 | Go loop（零 trampoline） | trampoline + exit reason |
| Memory | paged map (`Memory`) | flat 4GB mmap + mprotect |
| Registers | Go struct | control region + x86 register |
| Gas | Go 變數 | control region `[R15-48]` |
| 跨 invoke 重用 | 無 | 目前無（規劃中：by codeHash） |

### 效能演進 timeline

1. ~~per-block W^X toggle~~ → **已解決：dual mapping**（mprotect 歸零，原頭號成本 ~49%）
2. **現況**：per-invoke cold start + per-block trampoline（exec ~42% 為次大成本）
3. **近期**：block linking 全面化（jump/branch 也 link，減少 Go 往返）
4. **中期**：cross-invoke cache（`ExecutableMemory` by codeHash；`JITContext` 仍 per-invoke）
5. **可選**：L2 LockOSThread 外提到 `host.HostCall` 整段

### Graypaper 語意 vs Recompiler 特有

| 分類 | 項目 |
|------|------|
| Graypaper 語意（另章） | Gas model、Host Call (omega)、sbrk、djump、PVMtrace |
| Recompiler 特有 | mmap layout、dual mapping、trampoline、signal handler、block linking、register map |
| 兩邊共用 | `DeBlobProgramCode`、`preDecodeBlocks`、`GuestMemory` interface |

---

## 相關檔案索引

| 檔案 | 職責 |
|------|------|
| `compiler.go` | CompileBasicBlock 主體 |
| `recompiler.go` | BlockBasedInvoke 執行迴圈 |
| `host.go` | host-call dispatch layer |
| `execute.go` | executeBlockLocked + HandleSbrk |
| `x86signal/` | Signal handler（C stub） |
| `gas.go` | Per-instruction / block-based gas emit |
| `block_link.go` | Fallthrough linking |
| `djump_native.go` | Native djump dispatch |
| `code_cache.go` | PC → CompiledBlock 快取 |
