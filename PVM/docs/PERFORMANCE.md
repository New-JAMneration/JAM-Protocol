# Recompiler 效能：現況、問題與改進方向

本文件整理 recompiler 相對 interpreter 偏慢的原因，以及建議的解法與優先順序。  
適用於 **production fuzz path**（`BlockBasedInvoke`，無 PVMtrace / 無 `DebugSingleStepInvoke`）。

相關文件：[CONFORMANCE_TEST.md](./CONFORMANCE_TEST.md)、[bench 腳本](../../scripts/bench_fuzz_conformance.sh)。

---

## 1. 量測現況（2026-06，全量 fuzz conformance）

資料集：`pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces/`（336 folders、1179 JSON）。

| 情境 | client `test_folder` wall time | 相對 macOS interpreter |
|------|-------------------------------|------------------------|
| macOS native **interpreter** | **195 s** (~3.3 min) | 1.00× |
| Docker linux/amd64 **interpreter** | **389 s** (~6.5 min) | 1.99× |
| Docker linux/amd64 **recompiler** | **472 s** (~7.9 min) | 2.42× |

在 **相同 Docker/amd64** 環境下，recompiler 比 interpreter 多約 **84 s（+22%）**。

### 1.1 環境差異（重要）

在 **Apple Silicon Mac** 上跑 `--platform=linux/amd64` 的 Docker，x86_64 需經 **QEMU 模擬**。這會：

- 大幅拉長 Docker interpreter 相對 macOS native 的時間（約 2×）
- 放大 JIT 編譯的 CPU 成本

因此：

- **Mac Docker 上的數字**：適合驗 conformance，不適合評估 recompiler 設計上限
- **真實 linux/amd64 主機**：Docker interpreter ≈ native amd64；recompiler 才有機會在 PVM 密集、JIT 熱身後反超 interpreter

重跑 benchmark：

```bash
bash scripts/bench_fuzz_conformance.sh
```

映像須為 **production**（無 `trace` build tag），且 amd64 Mac 上建置時需：

```bash
docker buildx build --platform=linux/amd64 \
  --build-arg GP_VERSION="$(cat VERSION_GP)" \
  --build-arg TARGET_VERSION="$(cat VERSION_TARGET)" \
  --build-arg OUTPUT=new-jamneration-target \
  -t new-jamneration-target:latest -f docker/Dockerfile --load .
```

### 1.2 實測 profile breakdown（2026-06，`JIT_PROFILE`，全量 1179 JSON）

由 `JIT_PROFILE=1` 的 hot-path 計數實測（`jit_hotpath.go`；645 個 PVM-heavy invoke 累積）。**此節為硬數據，取代 §2.0/§2.6 先前的推估**。

**Wall time（Docker linux/amd64，QEMU，同一 build）：**

| backend | wall | PASSED | 備註 |
|---------|------|--------|------|
| interpreter | **434 s** | 1179/0 | |
| recompiler | **553 s** | 1179/0 | 含 `JIT_PROFILE` 計時 overhead；clean run 約 459 s（§2.0）|

QEMU 下 recompiler 比 interpreter 慢約 20–27%。

**時間去向（佔 invoke 總時間；ratio 才是重點，sums 跨並行 goroutine 會疊加）：**

| 階段 | 佔比 | 說明 |
|------|------|------|
| **compile** | **56.3%** | 其中 **mprotect W^X = 48.8%（全程）/ 87% of compile**；codegen 本身僅 7.5% |
| **exec**（native + trampoline） | **41.6%** | 多為 per-block 固定稅，非 PVM 運算（見 §4.1） |
| deblob | 1.9% | 每次重 decode |
| host（omega） | **0.08%** | 確認非熱點 |
| setup（mmap/init） | 0.09% | `MAP_NORESERVE` 4GB lazy，便宜 |

**關鍵 counts（每 invoke 平均；645 invoke）：**

| 計數 | 值 | 結論 |
|------|----|----|
| `execBlocks` | ~5634 | per-block native 往返 |
| `host` | ~35 | 與「30–40 host call」吻合 |
| `execBlocks : host` | **160 : 1** | per-block 往返 ≫ host call |
| `lock` | ~36（≈ host） | **L1 外提生效**，lock 非 per-block |
| `compile` | ~2846 | ≈ `execBlocks` 一半 → 每 invoke 重編、跑完丟（無跨-invoke 重用）|
| `mprotect` | ~5695（= compile×2） | 每 compile 一對 MakeWritable/MakeExecutable |
| `djump` | 0 | djump 已 native dispatch，不回 Go |

**兩大結論：**

1. **per-block W^X `mprotect` 是頭號成本（~49%）**，被 **QEMU TB-invalidation 放大**（`mprotect→PROT_EXEC` 會讓 QEMU 失效整塊 code 重翻譯）。真實 amd64 上 `mprotect` 是 µs 等級、佔比小很多，但仍真實；**dual-mapping 在兩種環境都贏**（§4.0）。
2. **exec 41.6% 幾乎都是 per-block 固定稅**（trampoline 暫存器存載、cgo fault-window、per-instruction gas）。JIT_PROFILE 直接量到的是 exec **總時間**；其內部「PVM 真正運算約 1/3」為**估算**（依每 block 指令數推算，見 §4.1），非直接量測。

---

## 2. 問題分析

### 2.0 關鍵澄清：host call 次數 ≠ Go↔native 次數

實務觀察（單次 accumulate invoke）：**幾萬個 instruction step，只有約 30～40 次 host call（`ecalli`）**。

這代表 **omega 派發本身不是主要熱點**；瓶頸更可能在：

```
ExecuteBlock 被呼叫的次數  ≈  本次 invoke 走過的 basic block 數
                              ≫  host call 次數（30～40）
```

每次進入 native（`executeBlockLocked`）固定做：

1. `SetFaultWindow` / `ClearFaultWindow`（仍 **per-block**）
2. entry trampoline（存 host callee-saved、設 R15、載入 13 個 PVM regs）
3. `callNative` → 跑完 **一個** block → exit trampoline 回 Go
4. `BlockBasedInvoke` Go loop 處理 fallthrough，**再呼叫下一個** `executeBlockLocked`

**`LockOSThread` 外提層級**（由內到外；數字越大越省 lock 次數）：

| 層級 | 鎖定範圍 | `LockOSThread` 次數（30k step / ~40 host call） | 狀態 |
|------|----------|--------------------------------------------------|------|
| **L0** | 每 `ExecuteBlock` | ~4 000+ | 已淘汰（改前） |
| **L1** | 每 `MachineInvoke`（`BlockBasedInvoke` 開頭） | ~40（≈ host call 段數） | **已實作** |
| **L2** | 每 service（`host.HostCall` 開頭至 HALT/PANIC/OOG） | **1** | 建議下一步 |

> L0→L1 是主要收益；L1→L2 只省 ~39 次 lock/service，絕對量小但實作成本低、語意仍安全（見 §4.1 做法 A′）。

Interpreter 在 `SingleStepInvokeDecodedBlocks` 裡可 **連續跑多個 block**，直到 `ecalli` 才離開該 loop；中間沒有 thread lock，也沒有 trampoline。

#### 數量級估算（30k step / 40 host call 的情境）

| 量 | 粗算 |
|----|------|
| 平均每段 host-call 區間 | 30 000 / 40 ≈ **750** 條指令 |
| 若平均每 block 5～10 條指令 | 每段約 **75～150** 個 block |
| 整次 invoke 的 `ExecuteBlock` 次數 | 30 000 / 7 ≈ **4 000+**（視 branch 密度而定） |

因此單次 service invoke（改前 L0）可能是：

- **~4 000 次** `LockOSThread` + trampoline 往返
- **~40 次** 完整 host call（omega + snapshot registers）

**實測（2026-06，Docker linux/amd64，1179 JSON fuzz conformance）**：

| 情境 | real (s) | PASSED | 備註 |
|------|----------|--------|------|
| Docker interpreter | 321 | 1179 | baseline |
| Recompiler L1（`executeBlockLocked`，無 block linking） | **459** | 1179 | 較先前 ~472s 略快（~3%） |
| Recompiler L1 + fallthrough linking（策略 a） | 575 | 1179 | **已修復**（見 §4.1 B 根因） |
| Recompiler L1、無 linking（對照） | 459 | 1179 | linking 修復前 baseline |

> **實測修正（§1.2）**：頭號瓶頸是 **per-block W^X mprotect（compile 的 87%、全程 ~49%）**，其次 per-block trampoline（exec ~42%）；兩者皆 per-block。lock 差確認次要，單靠 lock 外提無法反超 interpreter。解法：§4.0 dual-mapping（先）+ §4.1 block linking（後）。

`host.HostCall` 外層 loop 每輪呼叫一次 `MachineInvoke`；`MachineInvoke` 內部仍為每個 block 進一次 native（trampoline），host call 少並不能免除 block 級往返。

---

### 2.1 每次 `Psi_M` invoke 都 cold-start JIT（冷啟動最大宗）

`Psi_M_recompiler` 每次 accumulation / service invoke 都會：

1. `NewJITContext()` — 新建 4GB mmap guest sandbox
2. `NewExecutableMemory()` — 新建 JIT code arena
3. `DeBlobProgramCode()` — 重新 decode program
4. `NewRecompiler()` → `NewCodeCache()` — **空的** per-invoke cache
5. 每個 basic block **lazy compile**，invoke 結束後全部丟棄

程式路徑：`psi_m_recompiler.go`、`code_cache.go`（cache 註解為 per-instance、非跨 invoke）。

**後果**：同一 `codeHash` 的 service 在每個 block、每輪 accumulation 都重付 **DeBlob + codegen** 成本。  
Interpreter 同樣 DeBlob，但 **沒有 codegen**，直接在 Go 執行 `InstrMeta.Exec`。

```
accumulate invoke
    → 依 service 找 codeHash、載入 program blob
    → Psi_M_recompiler（全新 ctx + 空 cache）
    → 邊跑邊 compile
    → defer Close()，compiled native code 消失
```

### 2.2 每個 basic block 的 trampoline（exec ~42%，**次於 mprotect**；Lock 已外提至 L1）

`BlockBasedInvoke` 每進入一個 block 呼叫 `executeBlockLocked`（見 `recompiler.go` + `execute.go`）。**`LockOSThread` 已在 L1 外提**（每 `MachineInvoke` 一次），但 **entry/exit trampoline 仍 per-block**。

**後果**：兩次 host call 之間若有 N 個 block，產生 **N 次** trampoline 往返（lock 僅 1 次/段，若升至 L2 則整個 service 1 次）。  
以 30k step / 40 host call 估算，trampoline 總和仍達 **数千** 量級。

**與 interpreter 對比**：同樣區間 interpreter 全程留在 Go，零次 trampoline、零次 `LockOSThread`。

### 2.3 Host call 必須回 Go（次數少，短期非優化重點）

`ecalli` 固定 `JMP exit_trampoline`（`emit_basic.go`）。約 30～40 次/invoke 的 omega 開銷相對 block 級往返通常 **次要**。

JAM 語意仍要求 host call 在 Go 執行；短期 **不必** 投入 native omega。

其他較少見的 runtime exit：

- **sbrk 跨頁** → `HandleSbrk`（仍回 Go `mprotect`）
- **djump**（`djump_native.go`）→ **多數在 native 解析**：dispatch **hit** → native `JmpReg`（不離開）；dispatch **miss** → `emitDjumpMiss` 寫 `CONTINUE + resolvedPC`，回 Go 由 loop `lookupOrCompileBlock` 編譯（**不走 `DjumpResolve`**）；**空 jump table** → `emitDjumpExit` fallback `DjumpResolve`（少見）。§1.2 `djump=0` 指 Go-resolve 計數器幾乎不觸發
- OOG / HALT / PANIC → exit trampoline（終止，各一次）

### 2.4 其他成本（其中 mprotect 實測為頭號，見 §1.2）

| 項目 | 說明 |
|------|------|
| Per-instruction gas（GP v0.7.2） | 每條 PVM 指令 **4 條 inline x86**（`gas.go` `emitGasCheck`）+ block 尾 **每指令一個 OOG 落點**（`emitOutOfGasExit`）；block-based 已備好但未啟用 |
| JIT 編譯本身 | 首次進入 block PC 時 `CompileBasicBlock` |
| Signal handler / fault window | 每 `ExecuteBlock` 設定 TLS、fault window |
| **每次 compile 的 W^X 翻轉**（**實測頭號成本 ~49%**，§1.2） | 每個 `CompileBasicBlock`：`MakeWritable` → `Write` → `MakeExecutable` = **2× `mprotect` 整塊 16MB arena**（`executable.go`）。實測 mprotect = compile time 的 87%、全程 ~49%；QEMU `PROT_EXEC` 觸發 TB-invalidation 放大。**解法：§4.0 dual-mapping（消除所有 mprotect）** |
| Mac Docker QEMU | 放大上述所有 CPU 密集操作 |

### 2.5 何時 recompiler 理論上該贏？

```
recompiler 優勢 ≈ (native 執行節省)
                − (JIT 編譯成本)
                − (ExecuteBlock 次數 × 每 block 固定成本)
```

其中 **每 block 固定成本** ≈ **W^X mprotect（實測 #1，§1.2）** + fault window + entry/exit trampoline（L1 後 lock 已非 per-block）。

| 情境 | 預期 |
|------|------|
| 冷 invoke（首次 compile） | JIT 編譯（含 mprotect）+ block 往返雙重吃虧 → interpreter 贏 |
| 暖 invoke L1（已 compile）、仍每 block mprotect+trampoline | **仍輸 interpreter**（實測 §1.2，~+27%）|
| L1 + **§4.0 dual-mapping** + block linking | 才有機會縮小差距；單靠 lock 外提不足 |
| Fuzz 大量快速 validation fail、PVM 跑得少 | interpreter 贏（JIT 摊不回） |

### 2.6 建議的驗證方式（確認 LockOSThread 假設）

在 `ExecuteBlock` / `BlockBasedInvoke` 加計數（或 `testing.B` microbench）：

| 計數器 | 預期數量級（30k step invoke） |
|--------|------------------------------|
| `execute_block_calls` | ~10³–10⁴ |
| `lock_os_thread_calls` | L0：≈ block 數；**L1：≈ host call 段數（~40）**；L2：1 |
| `host_call_exits` | ~30–40 |
| `compile_basic_block_calls` | 首次 ≈ block 數；暖 cache ≈ 0 |

對照實驗：

1. ~~**L1** 外提 `LockOSThread`~~（**已完成**；conformance 1179/0，實測次要）
2. ~~fallthrough block linking~~（**已修復並啟用**，fallthrough only，§4.1 B；但 §4.0 dual-mapping 前 eager linking 會 **+25% regression**）
3. **§4.0 dual-mapping**（消 mprotect，§1.2 實測最大塊）→ 再 jump/branch linking、（可選）L2

L1 實測約 **~3%** wall time 改善（lock 確認次要）；**實測頭號瓶頸是 mprotect（§1.2 ~49%），其次 trampoline**——非 lock，見 §1.2 / §4.0。

---

## 3. 已知問題與前置條件（效能優化前必讀）

以下為程式審查確認的 **既有 bug / 語意缺口**。  
實作 §4 的 LockOSThread 外提、block linking、cross-invoke cache **之前或同時** 應先處理標為 **阻擋** 的項目，否則可能引入新 regression 或讓優化失效。

| # | 問題 | 嚴重度 | 阻擋 §4 優化？ |
|---|------|--------|----------------|
| 3.1 | Fault window stale | **中（防禦性）**；僅在錯誤外提 SetFaultWindow 時理論上有 SIGFPE 風險 | **不阻擋** LockOSThread 外提（方案 A：fault window 仍 per-block） |
| 3.2 | Executable arena 16MB 用盡 → 假 PANIC | 高 | **阻擋** cross-invoke cache；trace 模式已危險 |
| 3.3 | 跨頁 fault 位址 / 部分寫入 | 中（邊界 case） | 不阻擋，但語意對齊須知 |
| 3.4 | `host.HostCall` sbrk 死碼 | 低 | ✅ **已處理**（分支移除、留 unreachable 註解） |
| 3.5 | HALT/PANIC `Counter(PC)` 不一致 | 中 | ✅ **解法已實作**（剩單元測試） |
| 3.6 | Entry trampoline 在 `JITContext`；跨 invoke 會重複 emit | 高（§4.2） | **阻擋** cross-invoke cache 正確實作（見 §4.2 B3） |
| 3.7 | 全域 cache + 共享 arena 無 W^X 序列化 | 高（若多 goroutine 並行 compile/execute） | **阻擋** §4.2 直到確認並行模型（見 §4.2 B4） |

**建議前置順序（Step 0）：**

```
0a. arena 滿 → 明確錯誤（非 ExitPanic）+ trace single-step 停止無限 em.Write
0b. LockOSThread 外提至 L1（已完成）；可選 L2 外提至 `host.HostCall`（§4.1 A′）；SetFaultWindow 仍 per-block（§3.1）
0c. 刪 host.go dead sbrk；HALT emit 補寫 ExitPC（見 §3.4、§3.5）
0d. §4.2 前：entry trampoline 從 `JITContext` 移到 `ProgramJITEntry`（見 §4.2 B3）
→ 再進行 block linking、codeHash cache（§4.2 另需 B4 並行/W^X 策略）
```

---

### 3.1 Fault window `codeEnd` stale

**現況**

- `ExecuteBlock` 每 block 呼叫 `SetFaultWindow(guestBase, codeStart, codeStart+em.Used())`（`execute.go`）。
- Signal handler（`x86signal/x86_signal_linux.c`）：
  - **SIGFPE**（x86 `#DE`）用 `is_jit_code(rip)`，範圍為 `jit_code_start..jit_code_end`。
  - **SIGSEGV/SIGBUS** 用 `is_jit_fault(si_addr)` 查 guest 區，**不依** code 範圍。

**審查驗證（已確認，2026-06）**

JIT 內會觸發 **SIGFPE / `#DE`** 的 x86 指令只有 **`DIV` / `IDIV`**（`emit_arith_three.go` opcodes 193–196、203–206）。  
目前每個 emit 路徑在執行除法前已攔截：

| 案例 | 守衛 |
|------|------|
| `div_u_*` / `rem_u_*` | `divisor == 0` → 不執行 `Div`/`Div32` |
| `div_s_32` | `divisor == 0`；`divisor == -1 && dividend == INT32_MIN` → 不執行 `Idiv` |
| `div_s_64` | `divisor == 0`；`divisor == -1` → `INT64_MIN` 特例或 `-a`，不執行 `Idiv` |
| `rem_s_*` | `divisor == 0`；`divisor == -1` → 直接結果，不執行 `Idiv` |

**結論**：在現有 opcode emit 下，JIT 實務上 **幾乎不可能** 打到未處理的 `IDIV`/`DIV` 而觸發 SIGFPE。  
因此 `codeEnd` stale 導致 **process crash** 的機率極低；此項屬 **defense-in-depth**，不是 conformance 已測路徑上的熱點風險。

**理論風險（僅當錯誤外提 SetFaultWindow 時）**

若 LockOSThread 外提時 **也把 SetFaultWindow 外提一次**（只設當下 `em.Used()`），且未來某條路徑讓 `IDIV` 無守衛執行：

- 新 compile 的 code 可能落在 **大於舊 `codeEnd`** 的位址。
- 該處 SIGFPE 不會被 `is_jit_code` 辨識 → `chain_to_old` → 可能 crash。

**解法（擇一）**

| 方案 | 做法 |
|------|------|
| **A（建議）** | **只外提 `LockOSThread`**；**`SetFaultWindow` / `ClearFaultWindow` 仍 per-`ExecuteBlock`**。方案 A/B/C 在現況下 correctness 幾乎等價，選最簡單即可。 |
| B | 外提 SetFaultWindow 時改 `codeEnd = codeStart + em.Size()`（整個 16MB arena 上界）。 |
| C | 外提 fault window 時 **不可** per-block `ClearFaultWindow()`（會清 `guest_base_ptr`）；僅在整次 `BlockBasedInvoke` 結束 clear。 |

**Backlog（可選簡化，非短期）**

既然 SIGFPE 已被 div/rem emit 守死，可考慮：

- 移除 handler 的 **SIGFPE** 分支與 `jit_code_start/end`（`is_jit_code`）
- `set_fault_window` 僅保留 **`guest_base_ptr`**（供 SIGSEGV page fault），整段 invoke 設一次

這會讓 LockOSThread 外提更乾淨，但需另做 signal handler 回歸測試；**不影響目前 conformance 路徑的急迫性**。

**相關檔案：** `execute.go`、`recompiler.go`、`x86signal/x86_signal_linux.c`、`emit_arith_three.go`

---

### 3.2 Executable arena 16MB 用盡 → 假 PANIC

**現況**

- `ExecutableMemory` 固定 **16MB**（`executable.go` `DefaultExecutableSize`），無成長、無 LRU 驅逐。
- `em.Write` 滿了回 `executable memory full` error（`compiler.go` / `compiler_debug.go`）。
- `lookupOrCompileBlock` 失敗 → `BlockBasedInvoke` 回 **`ExitPanic, 0`**（`recompiler.go`），與真實 program panic 無法區分。

**風險**

| 路徑 | 行為 |
|------|------|
| Production `BlockBasedInvoke` | 每 PC block cache 一次；單 invoke 通常夠用，超大 program 仍可能滿 |
| **`DebugSingleStepInvoke`**（trace） | 每指令 `CompileSingleInstruction` → **每步 `em.Write`**、不進 cache、每 blob 含 `exit_trampoline`；30k step 量級 **極易爆 16MB** |
| **未來 cross-invoke cache（§4.2）** | arena **跨 invoke 累積**，只增不減 → 必須搭配驅逐或成長 |

**解法**

1. **區分錯誤**：compile / `em.Write` 失敗改 distinct handling（log、metric、或內部 error），**不要**映射成 guest `ExitPanic`。
2. **Trace 模式**：single-step block 進 `CodeCache`（key=PC）重用；或共用 **一份** shared `exit_trampoline`，避免每指令複製。
3. **Cross-invoke cache 前**：實作 **LRU 驅逐** `ProgramJITEntry`、或可擴展 arena、或 per-entry 大小上限。
4. **測試**：人造小 arena（測試 hook）assert 滿了回正確錯誤而非 conformance panic。

**相關檔案：** `executable.go`、`compiler.go`、`compiler_debug.go`、`debug_single_step.go`、`psi_m_recompiler.go`

---

### 3.3 跨頁存取：fault 位址與部分寫入

**現況**

- **Interpreter**（`decode.go` `storeIntoMemory` / `loadFromMemory`）：跨頁前檢查 **兩頁** 權限；fault 回報 **起始位址** `memIndex`；第二頁不可寫則 **完全不寫**。
- **Recompiler**（`emit_memory.go`）：`emitMemoryBoundsCheck` **只檢查 `addr >= 0x10000`**，無 size、無跨頁；實際靠單條 `StoreQword` / `LoadQword` + SIGSEGV。
- Signal handler 的 PAGE_FAULT payload 用 `si_addr`（**實際 fault 位址**，常為第二頁），非 instruction 起始位址。

**風險**

僅在存取 **剛好跨 mapped / PROT_NONE 邊界** 時與 interpreter 分歧：

- 回報 fault PC 不同（起始 vs 第二頁）
- native store 在 fault 前是否部分寫入第一頁，**不能**假設與 interpreter all-or-nothing 一致

Conformance fuzz 目前多半不測此 PAGE_FAULT 細節；**PVMtrace 對齊** 可能受影響。

**解法（中長期）**

1. Emit 時加 **page-aware check**（對齊 interpreter 兩頁邏輯），通過才 store/load；失敗 emit `ExitPageFault` + **起始位址**。
2. 或 fault handler 由 `rip` 反查 PVM instruction，回報 **instruction 的 mem 起始位址**（成本較高）。

**相關檔案：** `emit_memory.go`、`decode.go`、`x86_signal_linux.c`

---

### 3.4 `host.HostCall` 的 sbrk 分支為 dead code　— ✅ 已處理

**現況**

- Production：`MachineInvoke` = `BlockBasedInvoke`（`invoke_mode.go`）；sbrk 在 `BlockBasedInvoke` 內 `resolveSbrk`，**不會**以 `SbrkCallID` 回傳給 host。
- Trace：`DebugSingleStepInvoke` 內部自己 `HandleSbrk`（`debug_single_step.go`），同樣不回傳 `SbrkCallID`。
- `host.go:60-93` 的 `IsSbrkExit` 分支在兩條路徑皆 **unreachable**。（host 亦無 djump 分支；djump 已在 machine 層處理。）

**狀態：✅ 已完成**

- `host.go` 的 sbrk 分支本體**已移除**，僅保留 `// unreachable: sbrk is resolved inside BlockBasedInvoke / DebugSingleStepInvoke`（`host.go:61`）。
- （可選）補一個 assert/test 確認 `MachineInvoke` 不會回傳 `SbrkCallID`。

**相關檔案：** `host.go`、`recompiler.go`、`debug_single_step.go`

---

### 3.5 HALT / PANIC 的 `Counter(PC)` 與 interpreter 不一致　— ✅ 已對齊（解法 #1/#2 已實作）

**現況（已修）**

| 路徑 | `Psi_H.Counter` | 狀態 |
|------|-----------------|------|
| Interpreter `jump_ind` HALT | `instr.PC` | baseline |
| Recompiler native HALT（`emitHaltAtPC`，`emit_branch.go:22`） | **寫 `ExitPC = instrPC`** 再 `ExitHalt` | ✅ 已修 |
| Recompiler `djump` HALT/PANIC | native `emitDjumpPanic` 寫 `instrPC`（§2.3） | ✅ |
| Recompiler 其他 HALT/PANIC（`recompiler.go` switch） | 改回 `ReadExitPC()`（行 87/98） | ✅ 已修 |

Conformance **未比對** exit 時 Counter（故 1179/0 不保證此項）；上述為程式碼層級已對齊。

**解法狀態**

1. ✅ 所有 HALT emit 路徑寫 `ExitPC = instr.PC`（`emitHaltAtPC`）。
2. ✅ `BlockBasedInvoke` HALT/PANIC 回傳改用 `ReadExitPC()`（`recompiler.go:87,98`）。djump 已 native，無 `emitDjumpExit`-target-PC 混淆問題。
3. ⬜ **待補**：單元測試 `jump_ind` + HALT sentinel → Counter == instr.PC（conformance 測不到此項）。

**相關檔案：** `emit_branch.go`、`recompiler.go`、`host.go`

---

## 4. 解決方法（建議優先順序）

**實測（§1.2）後的優先序**——攻擊點依數據排序：

| 優先 | 解法 | 打哪塊（§1.2 佔比） |
|------|------|---------------------|
| **P0** | **§4.0 Dual-mapping**：消除 per-block W^X `mprotect` | compile 中的 mprotect（~49%）|
| **P1** | **§4.1 Block linking**（jump/branch；**需先做 P0**） | exec 的 trampoline（~40% of exec）|
| P3 | §4.3 Block-based gas（卡 GP 版本同步） | exec 的 per-instr gas（~25% of exec）|
| **最後（先量再決定）** | **§4.2 Cross-invoke compile cache** | P0 後僅剩 codegen ~7.5%（且僅暖命中）|

> `LockOSThread` 外提（L1）**已完成**且實測為次要（§1.2，`lock ≈ host`）；L2 可選但收益小。host call、setup、deblob 實測皆 <2%，**非優化重點**。
>
> ⚠️ **Cross-invoke cache（§4.2）刻意排在最後、做完 P0/P1 並量測無誤後再評估是否值得**：
> 1. **P0 會吃掉它大部分效益**——compile 56% 裡 **87% 是 mprotect**；P0 dual-mapping 把 mprotect 歸零後，「重編一次」只剩 codegen **~7.5%**，cache 能省的從「56%」縮到「~7.5% 且僅暖命中」。
> 2. 它是**結構動最大、正確性面最廣**的一項（`ExecutableMemory`/`CodeCache`/trampoline 擁有權搬家、並行 singleflight、lifetime/eviction、code upgrade 語意；見 §4.2）。
> 3. fuzz 場景多為**不同 codeHash、命中率低**。
> → P0/P1 都不依賴 cache，可獨立實作與驗證；cache 等其餘穩定後再量、再決定。

---

### 4.0 [P0] Dual-mapping：消除 per-block W^X `mprotect`（實測 #1）

**問題（實測 §1.2）**：`mprotect` 佔全程 ~49%、compile time 的 87%。每個 block compile 都 `MakeWritable → Write → MakeExecutable`，對整塊 16MB arena 翻轉兩次；QEMU 下 `PROT_EXEC` 觸發 TB 失效（整塊 code 重翻譯），成本被放大。

**做法**：code arena 開成 **dual mapping**——同一塊實體記憶體映射兩個別名：

```
fd := memfd_create("jit", 0)             // 或 shm_open / MAP_SHARED 暫存檔
ftruncate(fd, size)
rwPtr := mmap(fd, PROT_READ|PROT_WRITE)   // 寫入 native code 走這個
rxPtr := mmap(fd, PROT_READ|PROT_EXEC)    // 執行（NativeAddr / 跳轉 / trampoline）走這個
```

- `Write` 寫進 `rwPtr`；`NativeAddr` / `GetPtr` / entry trampoline 用 `rxPtr` 對應同一 offset。
- **完全不再 `mprotect`**：W 與 X 是同一份實體頁的兩個 view。
- 移除 `MakeWritable`/`MakeExecutable`/`writable` 狀態與其所有呼叫點。

**注意**：

- **icache**：x86-64 對自修改碼 coherent，新寫入透過 `rxPtr` 執行前同核心不需顯式 flush；跨核心由硬體保證。
- **W^X 安全性**：dual-map = 同份實體頁同時可寫可執行（不同位址），略弱於嚴格 W^X。此 JIT sandbox 可接受（guest 取不到 `rwPtr`）。若要保留嚴格 W^X，退而求其次用 **batched toggle**（一串 eager-link block 共用一次翻轉），效益遠不如 dual-map。

**預期效益**：直接回收 compile 的 ~49%（QEMU 尤其大，消除 TB-invalidation thrash），真實 amd64 也省 syscall。**並解鎖 §4.1 linking**——linking 的 eager 編譯上次 +25% regression，正是因為多編→多 mprotect（§1.2 證實 mprotect 是主成本）；mprotect 變免費後 linking 才能淨賺。

**相關檔案**：`executable.go`（核心改寫）、`compiler.go`/`execute.go`（移除 `Make*` 呼叫）、`context.go`（trampoline 也走 `rxPtr`）

---

### 4.1 [P1] Block linking：per-block trampoline 是設計落差（需先做 §4.0）

**設計落差（exec 實測 ~42%，§1.2）**：exec 多為 per-block 固定稅——entry/exit trampoline 暫存器存載、cgo fault-window、per-instruction gas、Go dispatcher。**估算**（非直接量測，依 ~5.3 instr/block）：trampoline ~33 條 + gas ~21 條 + 真正運算 ~27 條 x86/block → **PVM 真正運算約佔 native 1/3**，其餘是固定稅。

**預期 vs 實際**（穩態、非終止時回 Go 的時機）：

| | 回 Go 時機 | 次數/invoke |
|--|-----------|-------------|
| **預期** | 只有 `ecalli`(host call) + `sbrk 跨頁` | **~35–40** |
| **實際** | **每個 basic block** | **~5634** |

終止類（`OOG`/`HALT`/`PANIC`/`PAGE_FAULT`）各只發生一次；**`djump` 已是 native dispatch**（§1.2 `djump=0`）。所以可被 linking 留在 native 的是 **fallthrough / static `jump` / static `branch` 目標**（目標 PC 編譯期已知，理應 native `JMP` 串起，不該每 block 回 Go）。

**前提**：**必須先做 §4.0 dual-mapping**。linking 的 eager 預編譯會增加 compile→mprotect（§1.2 證實是目前主成本），未做 P0 前 linking 會像上次一樣 **+25% regression**。`LockOSThread` 外提（L1）已完成、確認次要（§1.2）。

#### 做法 A — L1：`LockOSThread` 外提至 `MachineInvoke`（**已實作**）

**不可**在公開 `ExecuteBlock` 內直接移除 `LockOSThread`——測試依賴內建鎖（TLS → signal handler 失效）。

**三層拆分**（測試零改動；machine 層可再外提一層）：

| 函式 | 鎖 | 呼叫者 |
|------|-----|--------|
| `ExecuteBlock`（exported） | 內建 `LockOSThread` | 測試、standalone |
| `executeBlockLocked` | 不鎖；caller 持鎖 | `BlockBasedInvoke` / `DebugSingleStepInvoke` loop |
| `blockBasedInvokeLocked`（建議，未實作） | 不鎖；caller 持鎖 | `host.HostCall` loop（L2） |

L1 現況：`BlockBasedInvoke` 開頭 `LockOSThread`；loop 內 `executeBlockLocked`；`SetFaultWindow` **仍 per-block**。

#### 做法 A′ — L2：`LockOSThread` 外提至 **整個 service**（`host.HostCall`）

**建議實作**：一次 `LockOSThread` 從 `host.HostCall` 進入到 HALT/PANIC/OOG/PAGE_FAULT 返回，中間所有 `MachineInvoke`（含 ~40 次 omega 派發）都在**同一 OS thread**。

```
Psi_M_recompiler → host.HostCall(pc)
  LockOSThread()                         // L2：整個 service 一次
  defer UnlockOSThread()
  loop:
    MachineInvoke(pc) → blockBasedInvokeLocked(...)
    omega / snapshot（Go，仍在此 OS thread）
    continue with pcPrime
```

**為何安全（§3.1 / fault window 問題是否仍存在）**：

| 議題 | L2 外提後 |
|------|-----------|
| **§3.1 `codeEnd` stale** | **與 lock 層級無關**。只要 **`SetFaultWindow` 仍 per-`executeBlockLocked`**，就不會因 L2 惡化。錯誤外提 SetFaultWindow 到 `HostCall` 才危險（與 L2 無關）。 |
| **SIGFPE / `is_jit_code`** | 仍防禦性（§3.1 div emit 守衛）；L2 不改變 |
| **TLS `guest_base_ptr`** | 整個 service 固定在同一 OS thread → **更符合** signal handler 假設 |
| **omega 在 Go 執行** | 在 locked thread 跑 Go 沒問題；期間不進 native，不觸發 JIT signal |
| **`lookupOrCompileBlock` 編譯** | 可在 locked thread 做 `MakeWritable`/`Write`；與 execute 同 goroutine，無 §4.2 B4 跨 goroutine 問題 |

**取捨**：

- **優點**：每 service 少 ~39 次 lock/unlock（40 段 → 1）；實作與 L1 同模式，風險低
- **缺點**：整段 service 執行期間 **pin 住一個 OS thread**（含 omega 耗時）；對 accumulate 單 goroutine 路徑通常可接受；不建議在多 service 並行共用同一 goroutine 時再疊更長的 critical section
- **收益預期**：**小於 L0→L1**（實測 L1 僅 ~3%）；仍值得做，因幾乎無 conformance 風險

**實作要點**：

- `host.HostCall` 開頭 `SetupSignalHandler()` + `LockOSThread()`（`defer Unlock`）
- `BlockBasedInvoke` / `DebugSingleStepInvoke` 改為 **不鎖**（或拆 `blockBasedInvokeLocked`）
- 公開 `BlockBasedInvoke` 可保留薄包裝 `LockOSThread` → `blockBasedInvokeLocked`，供 `compiler_test` 直接呼叫
- trace 路徑：`DebugSingleStepInvoke` 經 `MachineInvoke` → 同樣由 `HostCall` L2 覆蓋，移除其內部 lock

#### 做法 B — Block linking（**已修復並啟用**，fallthrough epilogue only）

> **根因（2026-06，183 fail）**：舊版在 block **epilogue 組裝中途**呼叫 `compileForLink` → 遞迴 `CompileBasicBlock` → **`c.asm.Reset()` 清掉父 block 已 emit 的指令**，父 block 只剩殘缺 native code。
>
> **修復**：
> 1. 在 `a.Reset()` **之前** `compileForLink(fallthroughPC)` 預編譯目標
> 2. epilogue 只 emit（`emitFallthroughEpilogue`），不再遞迴 compile
> 3. `CompileBasicBlock` 開頭標記 `linking[startPC]`，back-edge 到父 block 仍編譯中時 fallback `emitExitToPC`
>
> **驗證**：conformance **1179/0**；wall time 575s vs 無 linking 459s = **+25% regression**。**根因（§1.2 已查明）**：eager pre-compile 增加 compile → 增加 mprotect（目前主成本），不是「暖 cache 待重測」。**須先做 §4.0 dual-mapping**，mprotect 變免費後 linking 才淨賺。
>
> ⚠️ **在 P0 dual-mapping 之前，575s 不代表 linking 的穩態表現**——它含被 mprotect 放大的 eager-compile 稅；linking 的真實效益要在 mprotect 歸零後才量得準。

Block epilogue：**靜態可解析目標 PC** 時 → `JMP nextNativeAddr`，不走 `emitExitToPC` + Go loop。現況：

- **已實作：fallthrough only**（`emitFallthroughEpilogue` + `compileForLink`，`block_link.go`）
- **待辦（§5 #3）：`jump`(op40) / `branch`(op81–90,170–175) 目標 linking**——目前 `emitJump` / `emitBranch*`（`emit_branch.go`）的 taken/static 目標**仍走 `emitExitToPC` 回 Go dispatcher**，未用 `linkTarget`

**難點 — 前向參照 / lazy compile**（工作量常被低估）

現行 `lookupOrCompileBlock` 為 **on-demand**：編譯 block A 時，fallthrough / branch 目標 block B 的 `NativeAddr` **通常尚未存在**；loop back-edge 目標甚至可能是 **block 自身**（編譯中途尚無 callable 位址）。僅寫「編譯期已知 PC → `JMP nativeAddr`」不足以指導實作。

| 策略 | 做法 | 取捨 |
|------|------|------|
| **(a) 先編譯目標再 link** | 編譯 A 前遞迴或拓樸序 compile 所有靜態可達目標，再 emit `JMP rel32` | 實作較直覺、無 runtime patch；可能 compile 從未執行的 cold block |
| **(b) stub + runtime patch** | 先 emit `JMP` placeholder 或暫走 `exit_trampoline`；首次執行到邊時回 Go compile 目標 → `MakeWritable` → patch rel32 → `MakeExecutable` | 只 compile 熱路徑；每次 patch 需 **W^X 翻轉**（§2.4），且與並行 execute **互斥**（§4.2 B4） |
| **(c) PC→native dispatch table** | 邊界 `JMP [table[PC]]` 或寄存器間接跳；目標 compile 後填表項 | 多一次間接跳；可漸進填表、避免 rel32 patch；djump 可收斂至此 |

**(a′) whole-program compile（策略 (a) 的極端變體）**：一次把整個 program blob 的所有 block 編完（一輪 `preDecodeBlocks` → 全 compile），再回填 dispatch / `JMP rel32`。優點：**所有 `NativeAddr` 一次到位 → 消除前向參照問題，jump/branch 也能直接 link**；且若仍用 W^X toggle，整個 program **只翻一次**（dual-mapping 後此點變 moot）。缺點：cold path 也被編（同 (a)）、首次 invoke 編譯延遲集中。**與 §4.2 cache 最契合**（編一次、跨 invoke 重用），可作 §4.2 的前置實作方式。

短期建議：**fallthrough only + 策略 (a)**（同一 compile pass 內先 compile 下一 block）；conditional branch / back-edge 與 (b) 可二期；(a′) 待 §4.2 cache 一起做。選定策略前勿低估 §4.1 工時。
- **必須離開 native 或無法靜態 link 的情況**：
  - **ecalli** — 必須回 Go 跑 omega（語意硬需求）
  - **sbrk 跨頁** — 必須 Go `mprotect`（目前設計）
  - **OOG / HALT / PANIC** — 寫 ExitReason 後結束執行
  - **djump**（`jump_ind` / `load_imm_jump_ind`）— **已改 native dispatch（§1.2 `djump=0`，不再離開 native）**；下方為背景/歷史

**djump 現況（已 native，`djump_native.go`）**

djump 目標 PC 在 runtime 才算出，無法在 compile 時靜態 link，但**已不必回 Go**——`emitDjumpNative` 在 native 內完成：讀 jump table（control region 指標）、bounds + alignment + basic-block-start 檢查，再查 `PC → native addr` dispatch table。三條出口：

| 路徑 | 行為 |
|------|------|
| dispatch **hit** | native `JmpReg`，**完全不離開 native** |
| dispatch **miss**（目標尚未編譯） | `emitDjumpMiss`：寫 `CONTINUE + resolvedPC` → 回 Go，loop `lookupOrCompileBlock(resolvedPC)` 編譯後續編入 dispatch；**不走 `DjumpResolve`** |
| **空 jump table** | `emitDjumpExit` fallback → Go `DjumpResolve`（少見） |
| 非法目標（misalign / 越界 / 非 block start） | `emitDjumpPanic`（native 內） |

§1.2 `djump=0` 指 **Go-resolve 計數器幾乎不觸發**（hit 與 miss 都不算它）。

> （歷史）djump 曾做成假 host call（`DjumpCallID = 0xFE`）+ `exit_trampoline` 回 Go `DjumpResolve`；那是實作選擇、非語意硬需求，現已收斂為上述 native dispatch。

```
L0 改前:    ~4000× LockOSThread + ~4000× trampoline + ~8000× mprotect + ~40× host call
L1 現況:    ~40× LockOSThread  + ~4000× trampoline + ~8000× mprotect + ~40× host call  （~3%）
+P0(§4.0):  dual-mapping → mprotect 歸零（§1.2 實測最大塊 ~49%）
+linking:   trampoline/Go-loop 降至 ≈ host call + sbrk
            （目前僅 fallthrough；jump/branch 待辦，§4.1 做法 B）
```

**預期效益**：對「幾萬 step、幾十次 host call」的 accumulate，可能是 **最大單項優化**。

---

### 4.2 [最後，先量再決定] Cross-invoke compile cache（key = codeHash）

> **排序決定（2026-06）**：本項**延到 P0/P1 之後**再評估——理由見 §4 開頭 ⚠️（P0 後效益剩 ~7.5%、結構動最大、命中率低）。以下設計仍保留，供確定要做時參考。

**問題**：每次 `Psi_M` 重建空 `CodeCache`，同一 program 重複 codegen（其中 mprotect 那塊由 §4.0 解；本項只省剩餘 codegen）。

**正確性前提（key 與 upgrade）**：
- **key 必須是 content-addressed `codeHash`**（= 餵給 `Psi_M` 的 `code` 的 hash，亦即 `FetchCodeByHash` 用的那個）；`PreimageLookup[codeHash]` 保證 `codeHash ⟺ code bytes` 不可變。**嚴禁用 `serviceID` 當 key**（upgrade 後會 stale）。
- `upgrade`（`host_call_accumulate.go`）立即改 `ServiceInfo.CodeHash`；下一輪 accumulate 以新 codeHash 取 code → **cache miss → 重編**，不會 stale-hit。`MinItemGas`/`MinMemoGas` 不進 codegen，不必併入 key。
- 因 entry immutable → **不需 invalidation**；lifetime 純為記憶體問題（LRU / size cap）。
- 須確認 **compiled code 為 `f(blob)` 純函數**（R15-relative、per-instr gas 常數、djump dispatch 對固定 blob 穩定）→ 同 codeHash 可跨 service 共用。

**做法**：

```
codeHash → ProgramJITEntry {
    Program           // DeBlob 結果（immutable，可共享）
    ExecutableMemory  // 已 emit 的 native code arena
    CodeCache         // PC → CompiledBlock
    TrampolineAddr    // entry trampoline，整個 codeHash 只 emit 一次（見 B3）
    compileMu         // 可選：序列化 MakeWritable→Write→MakeExecutable（見 B4）
}
```

**每次 invoke 仍新建**（狀態每輪不同）：

- `JITContext`（guest memory、registers、gas、heap）
- `InitFromProgram(code, argument)`

**重用**（同一 codeHash）：

- 已編譯的 `CompiledBlock.NativeAddr`
- **共用** `ProgramJITEntry.TrampolineAddr`（不再 per-invoke emit）
- 避免重複 `CompileBasicBlock` / `DeBlob`

Compiled code 使用 R15-relative 定址（control region + guest），同一 program blob 的 layout 固定，**可跨 invoke 重用**（code 不變則不需 invalidate）。

**B3 — trampoline 擁有權必須搬家（已驗證，否則直接漏 arena）**

現況：

- `trampolineAddr` 快取在 **`JITContext`**（`context.go:64`）
- `getTrampolineAddr()` 把 entry trampoline **寫入** `ExecutableMemory`（`execute.go:51-84`）
- 每次 `SetExecutableMemory` 會 **`trampolineAddr = 0`**（`context.go:181`）

§4.2 若 `ExecutableMemory` 由 `ProgramJITEntry` 跨 invoke 共享、而 `JITContext` 仍 per-invoke：

- 每次 invoke 都會把 trampoline **重新 emit** 進共享 arena → arena **只增不減**，直接放大 §3.2 爆滿風險

**必須**：trampoline 從 `JITContext` 移到 **`ProgramJITEntry`**，每個 codeHash **只 emit 一次**；per-invoke `JITContext` 僅持有指標引用。

**B4 — 跨 invoke 全域 cache 的並行 / W^X 前置（已驗證）**

`code_cache.go:23` 明寫：`CodeCache` **Not thread-safe: each PVM instance runs on a single goroutine**。

改成全域 `sync.Map` + 共享 `ExecutableMemory` 後，若 accumulate invoke 之間有任何並行（多 service、多 core、背景 compile）：

| 競態 | 後果 |
|------|------|
| `CodeCache.blocks` 並行寫入 | map corruption / lost block |
| `em.used` 並行 `Write` | 覆寫 native code |
| 一 goroutine `MakeWritable`、另一 goroutine 正在 `callNative` | **process crash**（執行到被改寫或非 EXEC 頁） |

`sync.Map` 只解決 **lookup map**，不解決 **arena 游標** 與 **W^X 生命週期**。

**前置**（§4.2 實作前）：

1. **並行模型已確認：accumulate 是並行的**——`ParallelizedAccumulation` 用 `errgroup` + `singleflight` 同時跑多個 `Psi_M`（`accumulation.go:447/473`，djump crash stack trace 證實）。全域共享 cache/arena **必然並行存取**。**dual-mapping（§4.0）順帶消除最致命的 W^X-toggle race**（不再有 `MakeWritable` vs `callNative`），但 `CodeCache.blocks` 寫入與 `em.used` 游標 race 仍在 → 仍需 per-entry mutex 或 compile-once-readonly
2. 若有並行：`ProgramJITEntry.compileMu` 序列化 `MakeWritable → Write → MakeExecutable`（與 execute 互斥），或 **compile-once 後 arena 唯讀**、不再 patch
3. §4.1 策略 (b) stub patch 與此同一 mutex 約束

**實作步驟**：

1. Phase A：`sync.Map` 或 LRU：`[32]byte` → `*ProgramJITEntry`（含 `TrampolineAddr`）
2. Phase B：`Psi_M_recompiler` 結束只 `ctx.Close()`，不清除全局 cache
3. 可選：cache `DeBlob` 後的 `Program` 指標，避免重複 decode

**前置**：§3.2 arena 策略 **+** §4.2 B3（trampoline 搬家）**+** §4.2 B4（並行/W^X 策略）；否則 cache 會加速累積至 16MB 爆滿或並行 crash。

**預期效益**：消除 **冷啟動** 的重複 DeBlob + codegen；與 §4.1 互補（cache 解決 compile，外提/linking 解決執行）。

---

### 4.3 [P3] 啟用 block-based gas（GP v0.8.0 路徑）

**問題（已驗證）**：v0.7.2 per-instruction gas 每條 PVM 指令在 block 內 emit：

- **4 條 inline**（`emitGasCheck`：`MovMemToReg` / `TestRegReg` / `Jcc` / `SubMemImm32`，見 `gas.go:19-27`）
- Block 尾 **每指令一個 OOG 落點**（`emitOutOfGasExit`：寫 ExitPC/ExitReason + `JMP exit_trampoline`）

估算 code 膨脹或 block-based 節省時，應以 **4+N 落點/block** 為準，而非「3 條」。

**做法**（已備好，見 `gas.go`、`compiler.go` 註解處）：

1. `preDecodeBlocks` 計算 `BlockMeta.GasCost`（未來可接 opcode gas table）
2. Block 入口 `emitBlockGasCheck`：**2 條**（`SUB [R15-48], N` + `JS`）+ **一個** block 級 OOG 落點
3. 關閉 per-instruction `emitGasCheck` loop 與 per-instruction `emitOutOfGasExit` loop

Gas 以 **compile-time immediate** 燒進 machine code，不需 runtime 從 Go 傳參。

**實測佔比**：per-instruction gas 約佔 exec 的 ~25%（§1.2；每 PVM instr 4 條 x86 vs 真正運算 ~5–6 條）。

**⚠️ 卡 GP 版本同步**：block-based 是 **GP v0.8.0** 語意（block 入口一次性扣 gas）。recompiler **單方面**改、interpreter 仍 v0.7.2 per-instruction，會讓 **gas 餘額 / OOG 的 ExitPC** 分歧 → conformance 掛。**必須 interpreter 同步切換**才能啟用；v0.7.2 路徑下此項暫不可動。

---

### 4.4 [P3] 擴大 inline sbrk

**問題**：sbrk 跨頁必須 `HandleSbrk`（Go + mprotect）。

**做法**：

- 同頁內 heap 擴張維持 inline path（`emit_two_reg.go` 已有部分實作）
- 僅在 `newHP` 跨 page boundary 時 exit 到 Go

---

### 4.5 [長期，非短期] Native host-call 減少往返

僅在 profiling 證明某幾個 omega（如高頻 lookup）是瓶頸時考慮：

- 極少數 hot host call 的 native stub（函數指標 + 精簡 calling convention）
- Guarded specialization（特定 service code pattern）

Storage 語意複雜，**不建議**短期全面移入 native。

---

## 5. 建議實作順序

**Step 0 — 前置（§3，優化前或與優化並行）**

```
0a. §3.2 arena 滿：distinct error + trace single-step cache/shared trampoline（待辦）
0b. ✓ §3.4 host.go dead sbrk 已移除
0c. ✓ §3.5 HALT/PANIC ExitPC 已對齊（剩單元測試）
```

**實測後的主線（依 §1.2 數據排序）**：

```
已完成：
  ✓ L1 LockOSThread 外提（實測次要）
  ✓ fallthrough block linking（已修復；但 P0 前會 regression，見 §4.1）
  ✓ native djump（djump=0，不回 Go resolve）
  ✓ JIT_PROFILE 計數器（jit_hotpath.go，gated by env）

待辦（依收益 + 風險排序）：
  1. §4.0 dual-mapping ............... 消 mprotect ~49%（最大、低風險；同時解鎖 linking）
  2. §4.1 block linking (jump/branch)  消 trampoline ~40% of exec（需 #1 先做）
  3. amd64 上重跑 benchmark + PVMtrace 對照（確認 #1/#2 正確且有效）
  ── 以上穩定無誤後，再評估是否做 cache ──
  4. §4.2 cross-invoke cache ......... 結構最大、正確性面最廣；P0 後僅剩 codegen ~7.5%、命中率低 → 最後做、先量再決定
  5. §4.3 block-based gas ............ 卡 GP 同步，暫不可動
  （可選）§3.3 跨頁 page-aware emit；§4.1 A′ L2 lock
```

若 fuzz 場景 **每個 JSON 都是不同 codeHash、PVM 跑得短**：

```
- §4.2 cache 效益有限（命中率低），但 §4.0 dual-mapping 仍全面有效（每次 compile 都省 mprotect）
- 仍應優先 §4.0
```

---

## 6. 相關程式碼索引

| 主題 | 檔案 |
|------|------|
| 每次 invoke 生命週期 | `psi_m_recompiler.go` |
| Per-invoke 空 cache | `recompiler.go`（`NewCodeCache`）、`code_cache.go` |
| Go↔native 每 block | `execute.go`（`ExecuteBlock`、`executeBlockLocked`） |
| LockOSThread L1 | `recompiler.go`（`BlockBasedInvoke`）、`debug_single_step.go` |
| LockOSThread L2（建議） | `host.go`（`HostCall`） |
| Block linking（fallthrough） | `block_link.go`、`compiler.go`（epilogue 前 pre-compile） |
| Entry trampoline 快取 | `context.go:64`（`trampolineAddr`）、`context.go:181`（`SetExecutableMemory` reset） |
| W^X / mprotect（**§4.0 dual-mapping 目標**） | `executable.go`（`MakeWritable`/`MakeExecutable`、整塊 arena）、`compiler.go`（`Make*` 呼叫點） |
| JIT_PROFILE 計數器 | `metrics.go`（`jitProfile` / `jm`，gated by `JIT_PROFILE=1`） |
| CodeCache 執行緒假設 | `code_cache.go:23` |
| Host call 迴圈 | `host.go` |
| ecalli exit | `emit_basic.go` |
| Block gas（待啟用） | `gas.go`、`compiler.go` |
| preDecode / BlockMeta.GasCost | `block_info.go` |
| Interpreter 連續 block | `invocation.go`（`SingleStepInvokeDecodedBlocks`） |
| Fault window / SIGFPE | `x86signal/x86_signal_linux.c`、`execute.go` |
| Executable arena | `executable.go`、`compiler.go`、`compiler_debug.go` |
| 跨頁 memory 語意 | `decode.go`、`emit_memory.go` |
| HALT ExitPC | `emit_branch.go`、`recompiler.go` |

---

## 7. 待驗證項目

**前置（§3）**

- [x] L1 LockOSThread 外提；conformance 1179/0（SetFaultWindow 仍 per-block）
- [ ] L2 外提至 `host.HostCall` 後 conformance 仍 1179/0；§3.1 不因 lock 層級惡化
- [ ] arena 滿時不回 `ExitPanic`（§3.2）
- [ ] trace single-step 長程序不爆 16MB（§3.2）
- [ ] `jump_ind` HALT Counter == instr.PC（§3.5）

**效能（§4）**

- [x] `execBlocks : host` = 160:1（§1.2，per-block 往返 ≫ host call）
- [x] `executeBlockLocked` 兩層拆分；`compiler_test` PASS
- [x] L1 實測（§1.2）：recomp 553s（profiled）/ ~459s clean vs interp 434s
- [x] **mprotect = 頭號瓶頸**：49% of total、87% of compile（§1.2 / §2.4 B5）
- [x] 並行模型確認：accumulate 並行（`ParallelizedAccumulation`，§4.2 B4）
- [ ] **§4.0 dual-mapping 後：`mprotect` 計數歸零、wall time 下降**
- [ ] §4.2 同一 `codeHash` 第二次 invoke：無 `CompileBasicBlock`、無重複 trampoline emit（B3）
- [ ] §4.2 並行下 per-entry mutex / compile-once-readonly 無 crash（B4）
- [ ] §4.1 block linking（在 §4.0 後）：trampoline/Go-loop 次數降至 ≈ host call + sbrk
- [ ] 真實 amd64 Linux 上 recompiler vs interpreter 時間差（QEMU 校正後）
- [ ] PVM-heavy case（如 `228429307` / `1767896003_2013`）單獨 benchmark

---

*最後更新：2026-06-14（重排優先序為 **P0 dual-map → P1 block linking →（穩定後再評估）cross-invoke cache**：cache 延後因 P0 後其效益僅剩 codegen ~7.5%、且結構/正確性風險最高；§4.2 補上 content-hash key 與 code-upgrade 正確性前提）。先前（06-11）：JIT_PROFILE 實測 mprotect=頭號 ~49%、新增 §1.2/§4.0、§4.1 設計落差、§4.2 B4 並行。*
