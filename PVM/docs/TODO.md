# Recompiler 優化 TODO

前提：**效能 baseline** 以 production mode 為準（`linux/amd64/cgo`、**無** `pvmtrace` tag、**不**開下列 env）。**正確性**不只看 production：`pvmtrace` build 仍須能編譯、debug 路徑不能被優化改壞。Debug tag **統一為 `pvmtrace`**（recompiler 現有 `trace` 檔在 1a 一併改名；不再使用 `trace`）。

目前沒有正式 v0.8.0 jam-conformance accumulate corpus。過渡資料是 `PVM/consistent-testdata`（Psi_M input：code blob + argument + gas）。

**驗收順序（每個改動都走一遍）**：
1. 正確性（見步驟 0 的 gate；`-count=1` 避免 Go test cache）。
2. 效率：production-mode benchmark（見步驟 0）。驗收時 unset：`JIT_PROFILE`、`JIT_CPUPROFILE`、`JIT_PPROF_ADDR`、`JIT_PERFMAP`。

診斷工具（皆 env-gated，預設關；**overhead 不同，不要混為同一句**）：
- `JIT_PROFILE=1`：額外 `time.Now` + atomic counters。
- `JIT_CPUPROFILE`：sampling profiler。
- `JIT_PERFMAP=1`：compile 時 mutex + 寫 `/tmp/perf-*.map`，不是 per-op `time.Now`。
- `JIT_PPROF_ADDR`：HTTP pprof。
只有「不知道時間花在哪」才開。紀錄每次 benchmark 的 CPU / Go / kernel / `GOMAXPROCS` / `JIT_CACHE_MAX_PROGRAMS` / `-benchtime` / `-count`。

優先序：有把握的 production 熱路徑先做；可能優化列在後面，**實作前先討論**（見文末）。

---

## 歷史狀態（2026-07-05，v0.7.2 conformance，645 invokes）

當時 JIT_PROFILE（**過時，不可當 v0.8.0 優先序**）：invoke ~1086ms，run ~838ms，setup ~195ms，host ~125ms，compile ~35ms。setup 當時被標成最大可挖項。v0.8.0 已改 block gas / host-call / opcode，必須用 consistent-testdata 重測。

對 interpreter：tiny corpus UpdateAccumulate 1.57×+。full dataset 兩 backend 等速——PVM 佔 STF <4%；Safrole ring-VRF ~51%，其中 ~107s/run 是 fork-restore 清 ring-verifier cache（**非 PVM**）。

---

## ✅ 已完成（已 commit，v0.7.2 週期）

- **lazy block linking**：`compileForLink` 只連已編好的 block（`block_link.go`）。
- **跨 invocation code cache**（`compiled_program.go`）：per-CodeHash 共用編譯產物，single-flight、`cp.mu` 序列化編譯。
  - Phase 1：cache deblob 的 `*Program`（`PVM/program_cache.go`，兩 backend 共用）。
  - Phase 2：cache `ExecutableMemory` / `CodeCache` / djump dispatch table（recompiler-only）。
  - 結果：recompiler accumulate 反超 interpreter；fuzz 1179/1179、`-race` 0 races。
- **backend 選擇修正**（`cmd/fuzz`）：`--pvm-backend` flag 預設曾蓋掉 `JAM_PVM_BACKEND` env；已修。
- **JIT_PROFILE 工具**：counters / pprof / perf + `FlushProfile` + `jit_perfmap.go`。
- **fuzz-target STF timing**：`internal/fuzz` 累計 per-ImportBlock STF timing，`make test-timing-fuzz-trace` 驅動。
- **eviction（cross-invoke cache Phase 3 v1）**——已完成（`compiled_program.go` + `compiled_program_test.go`）：
  - 觸發：reactive——只在「編了新 distinct CodeHash、insert 進 cache」且超過 cap 時。
  - victim：`refCount==0` 中 `lastUsed` 最小者（LRU）。
  - 保護：`refCount>0` 絕不 Munmap；全在用 → soft cap。
  - cap：`JIT_CACHE_MAX_PROGRAMS`（預設 256，count-based）。
- **native block chaining（原待辦 #1(a)）**——已完成（`block_link.go` `emitChainOrExit`）：
  - static exit 查 PC→native dispatch table；miss 才回 Go，自癒後迴圈全 native。
  - 驗收（0.7.2，645 invokes）：roundTrips 3,310,512 → 28,302（117×↓）。
- **label `fmt.Sprintf` → int handle**——已完成。
- **v0.7.2 per-instruction gas fusion**——已完成；**v0.8.0 已由 block pre-charge + `gaschargedflag` 取代**（`gas.go` `emitBlockGasCheck`）。舊描述 `sub [R15-48],1` 不再適用。

---

## [Task: pvm-recompiler-opt-v080] Recompiler production-mode 優化（GP 0.8.0）

對 `PVM/docs/TODO_review.md` 的採納已併入本 Task，**不改該 review 檔**。第一次 review 多數已收；以下針對**第二次 review**（同日覆寫）核對程式後採納或反駁。

收尾 commit：原分支 `b4def943` **不是**目前 HEAD ancestor；本分支等價 cherry-pick 是 `e70cce5a`（`git merge-base --is-ancestor` 已確認）。

### 第二次 review：採納／反駁

**採納**

- **P0.1 tagged gate**：現列 `go test -tags=pvmtrace ./PVM/recompiler/...` 確實會 PASS，但 (1) 不含根套件 `PVM`，漏掉 `host_call_mem_trace_test.go` 的 `stubGuestMemory`；(2) 1a 完成前 recompiler debug 仍是 `trace`/`!trace`，只加 `pvmtrace` 測到的是 production `!trace` 路徑。本機已重現：`go test -tags=pvmtrace -c ./PVM/` 因缺 `GrowHeapTo` 編譯失敗。永久 gate 改 `./PVM/...`；遷移前用 `trace,pvmtrace`。遷移後 `rg` 確認 `PVM` 下無殘留 `trace` constraint。
- **P0.2 矛盾（gate vs 延後修正）**：步驟 0 若把「同 block 第二條 mem fault 的 ExitPC」當 **passing** 斷言，現在理應失敗（進 native 前 `WriteExitPC(blockStart)`；C handler 只寫 regs + `ExitReason`，不寫 `ExitPC`）。必須把 **passing gate** 與 **known-failing characterization** 拆開，跨頁語意亦然。
- **P0.3 候選 E**：`OffsetExitPC=32` 註解的 4B padding 是 `[R15-28, R15-24)`。`emitDjumpExit` 用 `MovRegToMem`（**64-bit** `MOV [R15-32], r64`）會蓋掉整段 padding。C header **沒有** `OffsetGasCharged`；只搬進既有 padding、且不改 C 可見欄位時不必改 C offset。前置：djump/`jump_ind` 的 ExitPC 改明確 32-bit store + layout overlap test + `jump_ind` 後 `GasCharged` 不被改寫。
- **P0.4 cold/warm 隔離**：`programCache`（PVM，never-evict）與 `theProgramStore`（recompiler）都是 process-global。Cold 的**每次被計時 invocation** 前都要在 timer 外重設兩層；compiled arena 在 `refCount==0` 時 close，cleanup 不算 ns/op。Warm 兩邊要處於相同 deblob-cache 狀態。不要 production exported reset；in-process 用 `export_test.go`，或各 case 獨立 process + `-benchtime=1x`。`code_hash` 用 `types.OpaqueHash` JSON（record 已是 `0x` + 64 hex）。缺值／長度錯／（corpus 路徑）zero hash 應 fail，不可靜默 uncached。`UncachedZeroHash` 是**另外構造**的 bypass case，不是把壞 record 吞掉。門檻：baseline 後、優化開始前寫回相對 interpreter、相對「改前 recompiler」regression、`allocs/op`、benchstat 判定。另加能打中各改動的 microbenchmark（見步驟 0）。
- **P1.5 pool cap**：`Psi_M` 亦服務 Psi_I / Psi_R，不只 Psi_A；`types.MaxWorkers = runtime.NumCPU()*2`。不要只綁 Psi_A worker。可配置 + VA budget；滿池策略與 reset 失敗即 `Close` 丟棄；control region 傾向整頁 `clear`。
- **P1.6 GrowHeapTo 用詞**：改善是 **N 次 syscall → 1 次**；`ctx.pages` 在候選 D 前仍 O(N) map write。不再存在「中途第 k 頁 mprotect 失敗」；改測整段 syscall 失敗時 heap pointer 與 `pages` 都未更新。`n≤h` / `n>b` 屬 host-call 層，與 OS failure unit test 分開。A/D/跨頁 bitmap/2a 要有單一 permission 更新入口。
- **P1.7** `NewAccumulateTraceContextIfEnabled` 每次 Psi_A 都 `Getenv("JAM_PVM_TRACE_DIR")`，且無 build tag（`accumulate_trace_context.go`）。`!pvmtrace` 可拆成直接 `nil`。數字與 JIT 分開報。
- **P1.8** 1b / 1c / 2d 可共用 rel32／cold-stub 基礎設施，但須獨立 commit 或連續 benchmark checkpoint。
- **P1.9** `program_cache.go` 寫明 Nothing is evicted；G 會讓 entry 更大。低優先：count/bytes cap，與 compiled-program eviction 分開。
- **Race gate**：涉及 cache/pool/single-flight 的 checkpoint 加 `go test -race ./PVM/...`。本機若因 toolchain 無法建 race（review 見 Go 1.26.4 `runtime/race: cannot find package`），改在可用 toolchain / CI 跑，不省略。

**反駁**

- **不把 PAGE_FAULT `ExitPC` 修正提前到 1a 之前。** Review 建議的「每筆可能 fault 的 mem op 前寫 `ExitPC=instr.PC`」正是 1a 要從 production 拿掉的那類 hot-path store，會把 1a 的數字汙染掉。RIP→PVM-PC metadata 是另一套 signal-safe 設計，不該擋 1a。已拍板：修正仍在優先序 1 **之後**；步驟 0 只放 characterization（預期目前 ExitPC = block start）。1a **禁止**順便加上 per-op `ExitPC` store。完整策略（RIP map vs 日後 store）實作前再定，並獨立量測，不與 1a 混報。
- **`jump_ind` + HALT sentinel 不是同一個 bug，不該和 PAGE_FAULT ExitPC 捆成「現在必失敗、必須移到 1a 前」。** `DjumpResolve` 對 `0xffff0000` 回 `(ExitHalt, pc)`；recompiler `emitHaltAtPC` 已寫 `ExitPC=instr.PC`。這是軟體路徑，可當步驟 0 的 **passing** 測試。若實測失敗再修，不預設延後。
- **P1.7 熱度不可與 B 等同、不插隊。** B 是**每次 omega**（含算好的 `HostCallName`）。`JAM_PVM_TRACE_DIR` 是**每次 Psi_A 一次**。列為 B 之後的低優先 Psi_A 路徑，不算 recompiler JIT 改善。

### 0. 量測與正確性骨架（先做）

Corpus（寫死數字易過期；benchmark 應印 inventory）：目前 **13** 筆 record、**2** 個 distinct blob、13 筆 dump 的 `exit_reason_type` 皆 **Halt**、`gas_in` 最大 19,999,999。工作量不要只看 `gas_in`（那可能只是避免 OOG）；應看 `gas_consumed = gas_in - gas_out`、dynamic blocks / mem ops / host-call 種類與次數、round-trips、是否 `grow_heap`。

- [x] **修 consistency parser**：record JSON 有 `code_hash`（`0x` + 64 hex），但 `psiADumpRecordFile` 沒讀、`minimalAccumulateHostArgs()` 沒設 `HostCallArgs.CodeHash`。直接用 `types.OpaqueHash` 的 JSON unmarshal。zero hash 會 bypass `GetOrDeblobProgram` 與 `acquireCompiledProgram`（每次 throwaway `cp.close()`）。現有 `TestInterpreterVsRecompilerPsiADump` 因此是 **每次 uncached**，不能當 warm。必須用 record 的 **`code_hash`（service CodeHash）**，不可用 `code_sha256`。corpus 路徑：缺值、長度錯誤、zero hash **fail**，不可靜默退回 uncached。
- [x] **Benchmark 三 case**（不要捏假 hash；**不要** production exported reset API）：
  1. `UncachedZeroHash`：**另外構造**的 bypass（不是把壞 record 吞成 zero）。
  2. `ColdNamedHash`：**每次被計時的 invocation** 前，timer 停止時重設 `programCache` **與** `theProgramStore`；compiled arena 在 `refCount==0` 時 close；cleanup 不算 ns/op。不能只在整個 benchmark 開始前清一次。
  3. `WarmNamedHash`：timer 外各自 prewarm；interpreter / recompiler 的 shared `programCache` 狀態必須相同。
  - in-process：`export_test.go` hook。更乾淨：每 case 獨立 process + `-benchtime=1x`，再 `-count` 取 cold samples。benchmark **不得**平行跑 cache reset。
  - `b.ReportAllocs()`；每次 iteration 用 fresh `HostCallArgs` / state。`count≥6` + benchstat。
- [ ] 寫下 v0.8.0 baseline。優化開始前把 acceptance 寫回文件：相對 interpreter、相對「改前 recompiler」允許的 regression、`allocs/op` 是否可增、benchstat 信賴區間 vs 固定百分比。
- [ ] **Microbenchmark**（corpus 只有 2 blob、全 Halt，不保證走到各路徑）：
  - 1a / 1b / C / E：memory-heavy block
  - B：log 關閉時的 host-call loop
  - A：一次成長 1 頁／多頁的 `GrowHeapTo`，並記 syscall 數
  - 1c / 2d：back-edge、emitted bytes、cold compile

**Passing correctness gate**（每個改動都要過；Psi_M dump **不夠**當唯一 gate）：
- `R()` 只把 `HALT` / `OUT_OF_GAS` 當特殊；`PAGE_FAULT` / `PANIC` / 其他一律折成 **`PANIC`**，且不比較 PC、fault addr、registers、`GasCharged`、memory、完整 `Addition`。兩邊 `BlockMeta.GasCost` 同源，一致 ≠ A.9 符合 Graypaper。
- [x] Psi_M corpus：補真實 CodeHash；可行範圍比 `Addition`。
- [x] **Machine-level** 小 program（passing）：exit reason、gas、`GasCharged`、regs、指定 memory；ecalli suffix、back-edge、OOG；同頁 PAGE_FAULT **payload**（起始 addr）。
- [x] `jump_ind` + HALT sentinel → `Psi_H.Counter == instr.PC`（軟體路徑，預期現在就過）。
- [ ] A.9：repo 既有 `pkg/test_data/new-gas-cost-model`（過渡／draft vectors，**不是**正式 v0.8.0 accumulate corpus）。
- [x] 先修 `stubGuestMemory`：補 `HeapPages` / `HeapMaxPages` / `GrowHeapTo`（與 1a 無關，但擋住 tagged gate）。

**Known-failing characterization**（步驟 0 要寫下現況，**不算** passing gate；修正在優先序 1 之後）：
- [x] PAGE_FAULT `ExitPC`：進 native 前是 block/suffix 入口；handler 不寫 faulting PC。同 block 第二條 mem fault：recompiler 報 block start，interpreter 報 `instr.PC`。斷言「目前 ≠ instr.PC」，不要當回歸綠燈。
- [x] 跨頁 access：payload / 部分寫入語意與 interpreter 已知不同；1 之後才修。

Tagged gate（`-count=1`）。**1a 完成前**：

```
go test ./PVM/... -count=1
go test ./PVM/ -run '^TestInterpreterVsRecompilerPsiADump$' -count=1
go test -tags='trace pvmtrace' ./PVM/... -count=1
```

**1a 統一 `pvmtrace` 後（永久）**：

```
go test ./PVM/... -count=1
go test -tags=pvmtrace ./PVM/... -count=1
go test ./PVM/ -run '^TestInterpreterVsRecompilerPsiADump$' -count=1
```

遷移後用 `rg` 確認 `PVM` 下無殘留 `//go:build` `trace`。涉及 cache / pool / single-flight 的 checkpoint 另加 `go test -race ./PVM/... -count=1`（本機 race toolchain 壞掉時改 CI，不省略）。

`./PVM/` 不含 `PVM/recompiler` 的 unit tests；因此永久 tagged gate 必須是 `./PVM/...`，不能只測 recompiler。

### 1. 有把握（production 熱路徑）

- [x] **(1a) 拿掉 production 的 mem-trace emit**
  - 六條 memory emit 都寫 `R15-160/-168`；production 無讀取者。
  - **不能只**給 `emit_record_mem.go` 加 tag（`emit_memory.go` 仍會呼叫）。要 `!pvmtrace` no-op，或呼叫點依 tag 分開。
  - **Debug tag 已拍板：統一 `pvmtrace`。** 把 recompiler 現有 `//go:build … && trace` / `!trace` 改成 `pvmtrace` / `!pvmtrace`（`debug_single_step.go`、`compiler_debug.go`、`invoke_mode*.go`、`host_call_{trace,notrace}.go`、`trace_mem_trace.go`）。Makefile 本來就只傳 `BUILD_TAGS=pvmtrace`，改完才會編進 single-step。`docs/6_PVMtrace.md` 仍寫舊 `trace` tag，一併改。
  - 驗收：production emitted code 不應對 `R15-160/-168` store；`-tags=pvmtrace` 仍要有。
  - **禁止**在 1a 順便於每筆 mem op 前寫 `ExitPC`（那是 PAGE_FAULT PC 修正，且會汙染 1a 數字）。遷移後永久 gate 改 `./PVM/...`。
- [x] **(B) `LogHostCallDispatchEnv` 不要每次 `os.Getenv`**（緊接 1a；高把握、glue 熱路徑）
  - 現況：每次 omega **之前** `host.go` 都呼叫此函式。即使 `JAM_PVM_HOSTCALL_LOG` 未設（production 預設關），函式仍：`Getenv` + `TrimSpace`，而且 **call site 已先算** `HostCallName` 與 serviceID。chaining 後 round-trip 幾乎只剩 host call，這段會被跑非常多次。
  - 作法：與 `jitProfile` 相同，`init`（或 `sync.Once`）快照 enabled bool；call site `if enabled { 才算 name 並呼叫 }`。
  - 代價：process 啟動後 `os.Setenv` 不會即時生效。除非產品明確要求 runtime 切換，否則可接受。
- [x] **(A) `grow_heap` 一次 `mprotect` 整段**（緊接 B；**N 次 syscall → 1 次**；排在 pooling 前）
  - 現況：`GrowHeapTo` 對每個新頁呼叫 `SetPageAccess` → 每頁一次 `mprotect` + 一次 map 更新。guest 一次 `grow_heap` 若要 N 頁，就是 N 次進出 kernel。v0.8.0 heap 成長是正式 host call，這條會打在 **host** 桶，不是 compile。
  - 新頁是連續 `[oldBound, newBound)`，一次 `unix.Mprotect(range, RW)`，**成功才**改 heap pointer 並更新 `ctx.pages`。改善是 syscall 次數；候選 D 前 `pages` 仍是 O(N) map write，不要寫成「metadata 一次更新」。
  - 要測：多頁／1 頁／零 growth；以可注入的 protection helper 模擬**整段** syscall 失敗 → heap pointer 與 `pages` 均未更新。`n≤h` / `n>b` 屬 host-call 層，與這支 OS test 分開。
  - A / D / 跨頁 bitmap / 2a 共用同一個 permission 更新入口，避免四套狀態。
- [x] **(1b) memory happy-path 不要每筆 `JMP` 越過 panic pad**
  - 只把 `jb` 反向會變成 happy path **每次 taken jump**。正確 layout：`jb panic_N` → mem op fall-through → block 末端才 `jmp epilogue`；各 mem op 的 `{label, instrPC}` 在 hot body 後統一 emit，**不可共用固定 PC 的 pad**。
  - 與 (2d) / 共用 cold stub **共用基礎設施**，但 1b / 1c / 2d 須獨立 commit 或連續 benchmark checkpoint，保留各項 delta。
- [x] **(1c) compile-time 已 link 的目標改 `JMP rel32`**
  - 同 16MiB arena 保證 signed rel32；12B `MOV imm64; JMP reg` → 5B。
  - **用詞更正**：runtime dispatch **hit** 才是 load slot 後 `JMP reg`；**miss** 是 `emitExitToPC` 回 Go。
  - 現有 assembler 只有 **同 buffer label** fixup。跨已寫入 block 的 rel32 要用預定 `em.Used()` + branch 結束位計算 displacement（`cp.mu` 已序列化 compile）。做 range check，超出則保留 abs jmp。
  - Lazy compile 下第一次 forward edge 多半尚未編好，走 dispatch；direct link 較常是 back-edge / 已編譯目標。記錄 direct-link 邊數與動態命中率，不只看 bytes。

### 2. 可能優化（優先序 1 之後；2a 設計已收斂、仍晚做）

- [ ] **(2a) pooling / 重用 `JITContext`**（**1 做完且 mmap 仍是瓶頸再開始**）
  - 現況：每次 `Psi_M` `mmap`/`munmap` 4GB+8KB（`MAP_NORESERVE` = 虛擬保留，不是立刻 4GB 實體）。`InitFromProgram` 典型最多 **6** 次 initial `mprotect`（RO content 2、RW 1、stack 1、arg content 2；空 segment 更少），不是 4–8。v0.7.2 曾量到 setup；v0.8.0 **待重測**，不能寫成「成本確定在」。
  - Reset（已拍板方向 + review 補齊）：`MADV_DONTNEED` **不會**把 protection 改回 `PROT_NONE`。還 context 時必須：
    - 前次 **全部** 可存取範圍（RO / RW+padding / heap / stack / arg+padding）`mprotect(PROT_NONE)`，即使新 program 比較小；
    - 再 `MADV_DONTNEED`（先一次整段 guest；若實測貴再改成上述連續區間）。**不要**逐頁 syscall；
    - 清零語意：舊 RW/stack/arg/heap 不能只覆寫新 content 長度；
    - 清 `pages`、`seg`、`heapLimit`、heap pointer、regs/gas/exit/`GasCharged`/mem-trace/djump/return addr+stack；
    - re-bind `executableMem` / trampoline，避免 stale pointer；
    - 等 `PVM.R()` 完成（`jitGuestMemory.Read` 已 copy）才回 pool。
  - **Cap**：可配置且受 VA budget 限制；**不要**只綁 Psi_A `types.MaxWorkers`（`NumCPU()*2`；Psi_I / Psi_R 也走同一 `Psi_M`）。`sync.Pool` 無硬上限 → 用 bounded retained pool。滿池：`Close` 多出的 context，或限制併發（實作前定一種）。記 allocated / in-use / retained / miss / drop。
  - Reset 中任一 `mprotect` / `madvise` 失敗：該 context `Close` 丟棄，不回池。Control region 傾向整頁 `clear(ctx.controlMem)` 再 re-bind。
  - 測試：同／不同 CodeHash 併發借還；舊 RW/stack/arg/heap 資料與 protection 不可沿用；舊 executable/djump pointer 不可沿用；reset 失敗不回池。
  - Sticky per-CodeHash 可選，但要處理同 hash 併發與 eviction。未證明 mmap 仍是瓶頸前不先做生命週期複雜度。
- [ ] **(2b) `LockOSThread` 外提**（可能；**不預設更快**）
  - 現已是每 MachineInvoke（host-call 段）lock 一次，不是每 block。提到整個 `HostCall` 會讓 omega 也釘在同一 OS thread，可能增加 scheduler 壓力。
  - 拆成兩個實驗：只外提 Lock（每次 native 仍 Set/Clear window）；另量 Set/Clear C-call。Throughput + latency + OS-thread 數，不只單執行緒 ns/op。
  - Window **不可**包 omega。現行 SIGSEGV 主要看 fault addr、**沒同時查 RIP**；window 開著時 Go 誤碰 guest 可能被當成 JIT fault。
- [ ] **(2c) `gaschargedflag` entry specialization**（1a–1c 有數字後再決定）
  - `ecalli` **不是** terminator。Chain 時已知 flag=false：應 **略過 load/test，但仍無條件 charge**，不是略過扣費。
  - 三種語境：(1) CONTINUE chain → unconditional-charge entry；(2) ecalli suffix → 尊重 omega 寫回的實際 flag；(3) 外部 fresh mid-block 且 flag=false → 仍走 generic 檢查（扣完整 containing-block gas）。
  - 較安全：保留 generic entry，另暴露 charge-entry / body-entry；只有控制流能證明 flag 時才走 specialized。測試：fresh mid-block false、ecalli resume true、terminator clear、OOG 不設 flag、panic/fault 保留 flag。做 2c 前先修仍把 `ecalli` 寫成 terminator 的舊文件。
- [x] **(2d) 每 arena 共用 exit trampoline**（中優先：code size / arena / cold compile，不只 i-cache）
  - 每 block 不再內嵌 13-reg spill。`buildCompiledProgram` 預寫一份；uncached / `em.Reset()` 路徑在第一次 compile 寫入。`JmpExit` 走 (1c) 的 `JMP rel32`，超出 signed rel32 才 abs jmp。
  - 驗收：`TestSharedExitTrampolineIsOutsideBlocksAndReused`、`TestSharedExitTrampolineReemitsAfterEmReset`；production + `pvmtrace` `./PVM/...` 已過。16MB/service high-water 仍待 corpus 變大後量。
- [ ] **(2e) assembler 少一次 copy**（低）
  - `Finalize` 可回傳 Reset 前有效的只讀 view，由 `em.Write` 做唯一 copy。

### 3. 低優先 / 先不做

- 任意常數折疊：要 benchmark 證明（常數 **addr** 的 ZZ check 見候選 C，較具體）。
- 合併連續 memory access：**不做**。
- eviction **效能**：2 個 CodeHash 測不到。eviction **正確性**已有 unit test，不要寫成完全沒測。
- hot omega native stub：要 pprof。
- arena-full 優雅處理：先看 `em.Used()/Size()`。
- [ ] shared `programCache` count/bytes cap（never-evict；與 compiled-program eviction 分開設計）。G 會讓單 entry 更大，先量 bytes。
- [ ] **(B′) `JAM_PVM_TRACE_DIR` constructor 依 `pvmtrace` 拆開**（Psi_A 路徑，**不是** JIT 熱路徑，不插在 B 與 A 之間）
  - `NewAccumulateTraceContextIfEnabled` 每次 Psi_A 呼叫 `os.Getenv`，檔案無 build tag。`pvmtrace`：讀設定；`!pvmtrace`：直接 `nil`，不查 env。數字與 B / JIT 分開報。

### 4. 非優化（正確性；**優先序 1 之後**，不當效能驗收）

- [ ] **跨頁 memory access**
  - (3.1) payload 應為起始 addr；(3.2) 須兩頁合法才寫。不跨頁：單一 `MOV` + 硬體 fault。
  - 跨頁：native 權限 bitmap 先查兩頁（~256KB，與 mprotect 同步）。**不**每筆都查表。
  - Signal **留下**當後盾（同頁 PROT_NONE/RO、漏檢查、guard、SIGFPE、非 JIT SIGSEGV）。不要整段 RW 只信 bitmap。
  - 不做：退回 Go；先寫第二頁；fault rollback。
- [ ] PAGE_FAULT ExitPC **修正**（characterization 見步驟 0；**不當** 1a 的一部分）。
  - 候選：每筆 mem op 前 store（hot path，須獨立量測）vs compile-time native-RIP → PVM-PC、handler signal-safe 查表。實作前再定。

### 5. Review 其餘候選（量測後再插入；A/B 已進 §1）

- [ ] **C. 常數地址 compile-time 去掉 ZZ `CMP`**  
  `addr < ZZ` → 直接 PANIC（該 instr PC）；`addr >= ZZ` → 不 emit lower-bound compare，權限仍靠硬體。indirect 保留 runtime check。
- [ ] **D. 用 `guestSegments` 區間取代 `ctx.pages` map**（與 2a 相關）  
  `seg` 目前只寫不讀；`IsReadable`/`IsWriteable` 逐頁 map。外層 Psi_M 可存取區是少數連續區間；grow_heap 只更新 heap end。先確認沒有 host call 任意改此外層 page mode（refine 的 pages 是 integrated PVM `Memory`，不是此 JITContext）。
- [ ] **E. `GasCharged` 改 disp8 + byte store**（要 benchmark）  
  `OffsetGasCharged=200` 超出 disp8；`emitGasCharged` 寫 4 bytes。disp8 內唯一明顯空隙是 ExitPC 後 4B padding（`[R15-28, R15-24)`）。**不可**在 djump 仍做 64-bit ExitPC store 時使用該 padding。前置：改 32-bit store + control-layout overlap test + `jump_ind` 後 `GasCharged` 語意。C header 目前不知道 `OffsetGasCharged`；只進 padding、不移動 C 可見欄位則不必改 C。只有 `OffsetExitPC` / `OffsetExitReason` / registers 被移動時才同步 C。
- [ ] **F. 共用 cold PANIC/HALT stub**（與 1b/2d 一起）  
  pad 只寫 instr PC 再跳 arena stub。OOG / mem PANIC 不要過度共用。
- [ ] **G. cache immutable decoded layout**（profile 後）  
  cache hit 仍 `DecodeSerializedValues` + memcpy static segments。
- [ ] **H. 拿掉 hot path `SetupSignalHandler()`**（低；`sync.Once` 已是 atomic）  
  `init` 已安裝。

建議實作順序：  
0 量測／passing gate／characterization（含修 stub、CodeHash、cache 隔離、microbench）→ 1a（`trace`→`pvmtrace`，**不加** ExitPC store）→ **B** → **A** → C → 1b 然後 1c 然後 2d（可共用基礎設施，分開 checkpoint）→ 再量 2c/E（E 先修 32-bit ExitPC store）→ 再決定 D 與 2a → 2b/2e/G/B′/programCache cap profile-driven。跨頁 bitmap 與 PAGE_FAULT PC **修正**跟在 1 之後、不當效能驗收。

---

## 已拍板

- **成功指標**：cold + warm。Warm 為主。Cold 不要輸太多。**具體 threshold 在 baseline 後、優化前寫回**（相對 interpreter、相對改前 recompiler、allocs、benchstat）。
- **(2a)**：整段 guest `MADV_DONTNEED` **加上** 舊 mapping `mprotect(PROT_NONE)` 與完整 control/page 清理；cap 可配置 + VA budget，不只綁 `MaxWorkers`；reset 失敗丟棄；優先序 1 且證明 mmap 仍是瓶頸再做。
- Benchmark 必須真實 `CodeHash`；三 case；Cold 每次 iteration 重設兩層 cache；獨立於現行 consistency 的 zero-hash 行為。
- **跨頁**：bitmap 只避免跨頁走 signal；同頁硬體 + signal 後盾；1 之後才做；步驟 0 只 characterization。
- **PAGE_FAULT ExitPC 修正不提前到 1a**；1a 不加 per-op `ExitPC` store。步驟 0 用 known-failing characterization。
- **Debug tag 統一 `pvmtrace`**；gate 在遷移前是 `trace,pvmtrace` + `./PVM/...`，之後永久 `pvmtrace` + `./PVM/...`。
- **(B)** 固定接在 1a 之後。**(A)** 接在 B 之後：N syscall → 1；`pages` 仍 O(N) 直到 D。
- 1b / 1c / 2d 分開 checkpoint。

## 需先討論（未同意前不實作）

- **(2c)**：1a–1c 有數字後；入口語意依上面三種控制流。
- **(2b)**：Lock 與 FaultWindow 分開量；window 不包 omega。
- **PAGE_FAULT ExitPC 的實作策略**（優先序 1 之後）：per-op store vs signal-safe RIP→PC 表。

---

## 收尾（非優化）

- ✅ commit：eviction、native block chaining、label handle 化、gas check 融合、`.bin` 讀取——原分支 `b4def943`，本分支等價 cherry-pick `e70cce5a`。
- ✅ `cmd/fuzz/main.go` 拿掉 `Ancestry item added` debug print。
- 清掉 stray build 產物（repo 根目錄 `fuzz`、`node` binary）或加 `.gitignore`。
- （非 PVM）ring-verifier cache 修法交給 blockchain 側：`GetVerifier` cache key 加 gammaK hash、`restoreWithState` 不再 `ClearVerifierCache`。

---

## [Task: pvm-synthetic-accumulate-corpus] Synthetic Ψ_M accumulate corpus (opt measurement)

Dump 的 0.7.2 blob 在 0.8.0 ISA 下會錯位，且 `HostCallArgs` 是空的，Warm 幾乎只量到 mmap。本 Task 合成 **0.8.0** program + 已 decode 的 stub state，用來判斷 production 優化有沒有打到 native / host / setup，**不是** jam-conformance，也不是 omega 全條件正確性。

**範圍**

- Invoke：`Psi_M` @ **PC 0** + `AccumulateOmegas`。不做 Ψ_A 外層，不做 parallel / 多 service 同時 invoke。
- Storage：只預填 `StorageDict` 與 `PreimageLookup`（已 decode、可走通）。**不**建 unmatched `StateKeyVals`。
- 兩支同一骨架的 program，只改 working set：
  - **SmallData**：`grow_heap` 少頁（仍 `h < n`）；短 mem loop；小 blob。
  - **LargeData**：`grow_heap` **多頁**，但 **不得超過** heap 合法範圍：目標頁 `n` 必須 `h < n ≤ b`，其中 `h = heapStart/ZP`、`b = (stackStart − ZZ)/ZP`（`heapLimit` = stack start）。選 `n` 時遠低於 `b`（例如 +64 頁），不要貼齊 stack。mem loop 走在已 grow 的 `[heapStart, n·ZP)`。
- Host-call：Accumulate 能 dispatch 的 ID 各跑一次（含 `log` **一次**，不要放進迴圈）。資料路徑（`gas` / `grow_heap` / `fetch` / `lookup` / `read` / `write` / `info`）要求 Continue、不 panic。其餘允許 r7=`HUH`/`WHO`/`CASH`/`CORE`。Refine-only（7–14）不呼叫。
- 煙霧測試：interpreter vs recompiler 皆 Halt、gas 同號、無 Go panic。**不**比 omega 回傳碼、不比 storage 寫回。Halt 前清 r7/r8，避免 `R()` payload 不一致。
- Bench 矩陣：`{SmallData, LargeData} × {WarmNamedHash, ColdNamedHash} × {Interpreter, Recompiler}`。可另加 `UncachedZeroHash` 當 compile 參考，不當大/小 data 主對照。Cold：每次計時 invoke 前重設 `programCache` 與 `theProgramStore`。

**不做**：改寫 0.7.2 dump blob；改 dump harness 語意；2a/2b/2c；PAGE_FAULT ExitPC 修正。

- [x] 合成 blob builder（`buildBlobExact` 風格；Standard Code 包裝；同一骨架 Small/Large）。
- [x] Stub `HostCallArgs`：`StorageDict` + `PreimageLookup`；balance / items / bytes 足以讓 `write` 不 `FULL`。
- [x] 建構時斷言 Large/Small 的 `grow_heap` 目標 `h < n ≤ b`。
- [x] Dual-backend 煙霧測試。
- [x] `BenchmarkPsiMSynthetic`：上列矩陣；`b.ReportAllocs()`；現有 `BenchmarkPsiM` dump 案例不改。
