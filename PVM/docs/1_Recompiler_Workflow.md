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
  └─ BlockBasedInvoke ────────── 內層（≈ Go→native round-trip 呼叫數）
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
  2. blockGas = blockGasCostAt(startPC)（A.9，compile 時 bake 進 native code）
  3. emitBlockGasCheck(blockGas) + block OOG landing pad
  4. 對每條指令：
       opcodeHandlers[opcode](c, asm, instr) → emit x86 指令
       terminator 最後一條 → emitGasCharged(false)（離開 block 重置 flag）
  5. Block epilogue：
       fallthrough 目標已編譯 → JMP NativeAddr（compile-time link）
       否則 → emitChainOrExit：runtime 查 PC→native dispatch table，
              hit 直接 jmp 進目標；miss 才寫 CONTINUE + 下一 PC → exit_trampoline
  6. EmitExitTrampoline
  7. Assembler.Finalize() → []byte（機器碼）
  8. em.Write(code) → 寫入 ExecutableMemory
  9. CodeCache.Put + registerDispatch（djump dispatch table）
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
- **Lazy-only linking**：`compileForLink` 只連結「已編譯」的目標，**不** eager 預編下一個
  block。早期 eager 版（沿 fallthrough 遞迴預編，depth cap 256）實測多編了 ~38% 從未
  執行到的 block，而 codegen 正是 compile 的主成本 → 改為目標已存在才 compile-time `JMP`
- **Forward reference → runtime chaining**：目標尚未編譯時 emit `emitChainOrExit`
  （`block_link.go`）：查 PC→native dispatch table，hit 直接 `jmp` 進目標；miss 才
  `emitExitToPC` 回 Go。Go dispatcher 編譯目標並 `registerDispatch` 填表後，同一出口
  從此走 native（miss 自癒），loop back-edge 因此收斂成全 native。適用 fallthrough /
  靜態 branch 兩路 / `jump`；indirect jump 另走 djump dispatch（見 `5_Djump_Dispatch.md`）

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

### Chaining 對迴圈次數的影響

Chaining 前，每個 basic block 執行完都經 exit trampoline 回 Go，由本 loop 查下一個
block 再進 native——round-trip 次數 ≈ 走過的 block 數（conformance dataset 實測
**~4000+ 次/invoke**）。

Chaining 後，block epilogue 在 native 內直接 `jmp` 進下一個已編譯 block
（compile-time link 或 dispatch table hit）：PVM registers 全程留在 x86 register、
不經 trampoline、不回本 loop。只有 **host call、chain/djump miss（冷啟一次性）、
sbrk 跨頁、終止類出口** 才回 Go。同一 dataset 實測 round-trip 降至
**~40 次/invoke**（roundTrips 3,310,512 → 28,302，117×↓）——一次 invoke 的內層
loop 幾乎只在必要出口才轉一圈。

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
| chain miss（static 目標未編譯） | `BlockBasedInvoke` 編譯 + 填 dispatch table | 否（一次性；之後同出口 native chain） |
| block link / chain hit JMP | native 內 | 否（完全不回 Go） |
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
| 跨 invoke 重用 | deblob 後的 `*Program`（`PVM/program_cache.go`，兩 backend 共用） | 同左 + per-CodeHash 編譯產物（`compiled_program.go`：`ExecutableMemory` / `CodeCache` / djump dispatch table；single-flight + LRU eviction） |

### 效能演進 timeline

1. ~~per-block W^X toggle~~ → **已解決：dual mapping**（mprotect 歸零，原頭號成本 ~49%）
2. ~~per-block trampoline~~ → **已解決：native block chaining**（fallthrough / 靜態 branch /
   jump 全走 dispatch table chain；roundTrips 3.31M → 28.3k（117×↓），迴圈收斂成全 native）
3. ~~per-invoke cold start（重編）~~ → **已解決：cross-invoke cache**（per-CodeHash 共用
   `ExecutableMemory` / `CodeCache` / djump dispatch，single-flight + LRU eviction；
   `JITContext` 仍 per-invoke）
4. **現況最大可挖項**：per-invocation mmap/munmap（`NewJITContext`，setup ~195ms）→
   評估 JITContext pooling（TODO #1(b)），並將 `SetFaultWindow` 從 per-block 移到 per-invocation
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
| `block_link.go` | Block chaining：compile-time link + runtime dispatch-table chain |
| `compiled_program.go` | 跨 invocation code cache（per-CodeHash artifact、single-flight、LRU eviction） |
| `djump_native.go` | Native djump dispatch |
| `code_cache.go` | PC → CompiledBlock 快取 |
