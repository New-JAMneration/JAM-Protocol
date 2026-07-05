# Recompiler 優化 TODO

前提:正確性已通過 conformance(0.7.2 fuzz,1179/1179)。**優先序一律以實測為準**。

量測工具(皆 env-gated,預設關閉):
- `JIT_PROFILE=1`:per-phase 計時(setup/deblob/compile/host/run)+ 計數(roundTrips、lock、cache hit/miss、host、compile、djump),每 10s 印累計,退出時印 `[JIT-PROFILE TOTAL]`。
- `JIT_CPUPROFILE=<file>` / `JIT_PPROF_ADDR=host:port`:pprof CPU。Psi_M 標 `phase=pvm`,可 `go tool pprof -tagfocus=phase:pvm`。
- `JIT_PERFMAP=1`:寫 `/tmp/perf-<pid>.map` 供 perf 依 block symbolize(WSL2 perf 受限,用 pprof)。
- `make test-timing-fuzz-trace`:fuzz target 端的 per-ImportBlock STF + Psi_A timing(現行資料集 + recompiler 路徑)。

---

## 目前狀態(2026-07-05,conformance corpus,645 invokes)

cross-invoke cache + chaining + label handle + gas 融合後的 JIT_PROFILE:

| 項目 | 量級 | 說明 |
|------|------|------|
| invoke 總計 | ~1086ms | 原 ~1897ms(chaining 前) |
| run | ~838ms | 最大宗;roundTrips 28k(原 3.3M,幾乎只剩 host call 22.7k 必要出口) |
| setup | ~195ms | `NewJITContext` per-invocation mmap → **現行最大可挖項(待辦 #1(b))** |
| host | ~125ms | 必要成本 |
| compile | ~35ms / 4.5k 次 | per-distinct-code + label handle 化(原 per-invocation 2.09M 次) |
| deblob | <1% | Program cache |

對 interpreter:tiny corpus 的 **UpdateAccumulate 1.57×+**。**full dataset(fork session)兩 backend 等速**——該資料 PVM 佔 STF <4%,Safrole ring-VRF 佔 ~51%,且其中 ~107s/run 是 fork-restore 每 block 清 ring-verifier cache 造成的 `GetVerifier` 重建(**非 PVM 問題**,修法:cache key 加 gammaK hash、restore 不清 cache,潛在 ~2× STF)。

---

## ✅ 已完成(已 commit)

- **lazy block linking**:`compileForLink` 只連已編好的 block(`block_link.go`)。
- **跨 invocation code cache**(`compiled_program.go`):per-CodeHash 共用編譯產物,single-flight、`cp.mu` 序列化編譯。
  - Phase 1:cache deblob 的 `*Program`(`PVM/program_cache.go`,兩 backend 共用)。
  - Phase 2:cache `ExecutableMemory`/`CodeCache`/djump dispatch table(recompiler-only)。
  - 結果:recompiler accumulate 反超 interpreter;fuzz 1179/1179、`-race` 0 races。
- **backend 選擇修正**(`cmd/fuzz`):`--pvm-backend` flag 預設曾蓋掉 `JAM_PVM_BACKEND` env(bench 兩邊都跑 interpreter);已修。
  - 註:`cmd/node` 的 `--pvm-backend` switch 仍為**本地未 commit**。
- **JIT_PROFILE 工具**:counters/pprof/perf + `FlushProfile` + `jit_perfmap.go`;退役 `exec(est)` 減法桶、`execBlocks`→`roundTrips`、加 `cacheHits/Misses`。
- **fuzz-target STF timing**:`internal/fuzz` 累計 per-ImportBlock STF timing,`make test-timing-fuzz-trace` 驅動。
- **eviction(cross-invoke cache Phase 3 v1)**——**已完成、未 commit**(`compiled_program.go` + `compiled_program_test.go`):
  - 觸發:reactive——只在「編了新 distinct CodeHash、insert 進 cache」且超過 cap 時。
  - victim:`refCount==0` 中 `lastUsed` 最小者(LRU;map + 單調序號,淘汰是罕見的 O(n) 掃描)。
  - 保護:`refCount>0`(執行中 arena)絕不 Munmap;全在用 → soft cap。acquire/release 配對(release 接 `Psi_M_recompiler` defer)。
  - cap:`JIT_CACHE_MAX_PROGRAMS`(預設 256,count-based)。
  - 驗收:單元測試(LRU 淘汰/Munmap/重建/soft-cap);cap=1 conformance **1179/1179**;node fuzzy `-race` **0 races**。
- **native block chaining(原待辦 #1(a))**——**已完成、未 commit**(`block_link.go` `emitChainOrExit`):
  - static exit(fallthrough / branch 兩路 / jump)改為查 PC→native dispatch table:hit 直接 `jmp` 進目標 block,miss 才 `emitExitToPC` 回 Go;miss 由 dispatcher 編譯目標 + `registerDispatch` 填表後自癒 → 迴圈收斂成全 native(back-edge 不再每圈回 Go)。
  - dispatch table 從 djump-only 擴為所有非空 program 皆建(`len(Bitmasks)>0`);slot 原子寫、native 端 aligned 8B 讀(同 djump hit path)。compile-time 直接 link 保留(省 lookup)。
  - single-step(trace tag)設 `c.singleStep` 關 chaining,維持每指令回 Go。
  - 驗收(fuzz conformance 0.7.2,645 invokes):roundTrips **3,310,512 → 28,302(117×↓)**、roundTrips/lock 142.7 → 1.22、invoke **1897ms → 1143ms(-40%)**、run 1652ms → 883ms;conformance **1179/1179**;traces 全 modes 800/800;fuzzy_light `-race` 0 races。
  - 註:全 native 迴圈 Go runtime 無法 preempt,靠每 block 的 gas check 保證有界。

---

## 待辦(尚未實作)

### 1. 減少執行階段開銷
- ~~(a) native block chaining~~ ✅ 已完成(見上)。
- **(b) per-invocation mmap / syscall**:`NewJITContext` 每次 mmap/munmap 大 guest region → 評估 **pooling/重用 JITContext**;`SetFaultWindow` 從 per-block 移到 per-invocation。chaining 後 setup(207ms)相對佔比升高,此項效益上升。

### 2.(低)codegen / emitted-code 品質
compile 已小,純 codegen 速度優先序低。經解剖,子項的判定:
- ~~gas check 融合(4→2 條)~~ ✅ 已完成、未 commit(`gas.go`):`load+test+jcc+sub` 融成 `sub [R15-48],1; js oog`(語意等價:pre<1 ⟺ post-sub<0;OOG 冷路徑 `sub -1` 補回,回報 gas 與 interpreter 一致)。run 883→838ms(-5%);驗收:traces 800/800、conformance 1179/1179、`-race` 0 races。
- ~~label `fmt.Sprintf` → int handle~~ ✅ 已完成、未 commit(`asm` + 全 emit 檔;設計/實測見 `2_x86_Assembler.md` §3.3.1)。compile 70ms→37ms(-47%),emitted bytes 不變;驗收:traces 800/800、conformance 1179/1179、`-race` 0 races。
- 常數折疊:緩,要 profile 證明才動。連續 memory access 合併:**不做**(page-fault 語意 per-access,正確性風險)。

### 3.(低)eviction 後續
proactive 淘汰(從 state 訊號主動刪「已知不會再用」的 CodeHash);bytes-based cap;arena-full 優雅處理(目前滿了 → compile error,16MB/service 通常夠);`*Program` cache 也加界限。

### 4.(暫緩)block-based gas(GP 0.8.0)
0.7.2 仍 per-instruction,改了**沒測資料可驗**。`gas.go` 已備好 `emitBlockGasCheck` / `emitBlockOutOfGasExit`,等 0.8 向量再開。

---

## 收尾(非優化,commit 前處理)
- commit(累積中的未 commit 工作):eviction(`compiled_program.go`/`_test.go`)、native block chaining(`block_link.go` 等)、label handle 化(`asm` + emit 檔)、gas check 融合(`gas.go`)、`.bin` 讀取(`cmd/fuzz/main.go`,含 `--format`/`--skip`)。
- `cmd/fuzz/main.go` 拿掉 `Ancestry item added` debug print、更新過時的 ancestry NOTE 註解。
- `cmd/node` 的 backend switch + profiling 檔案仍為本地未 commit(上次刻意排除)。
- 清掉 stray build 產物(repo 根目錄 `fuzz`、`node` binary)或加 `.gitignore`。
- (非 PVM)ring-verifier cache 修法交給 blockchain 側:`GetVerifier` cache key 加 gammaK hash、`restoreWithState` 不再 `ClearVerifierCache`(full-dataset 實測 `GetVerifier` 107s/run、45%)。
