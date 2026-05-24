# ServiceAccount globalKV 重構 — Phase 1 完成 + Phase 1 收尾 + Phase 2 (Trie Store + 增量 Merklize) + Phase 3 (未來規劃)

## Phase 1 — globalKV 重構（已完成 ✅）

> 目標：把 `ServiceAccount` 的 `StorageDict`（a_s）和 `LookupDict`（a_l）合併為
> 統一的 `globalKV map[StateKey][]byte` + 增量計數器。`PreimageLookup`（a_p）保持獨立不動。

### 完成狀態

| Step | 狀態 | 摘要 |
|------|------|------|
| 1 | ✅ | `ServiceAccount` 加 `globalKV` + `totalNumberOfItems` / `totalNumberOfOctets` + `NewServiceAccount()` |
| 2a | ✅ | `NewStorageStateKey` / `NewPreimageMetaStateKey` / `NewPreimageLookupStateKey` |
| 2b | ✅ | `internal/utilities/safemath/` (Add/Sub/Mul + ErrOverflow) |
| 2c | ✅ | Get/Insert/Delete/Update + Clone + `cloneMapOfSlices` 通用 helper（含 `a00b5ba0` DeepCopy + new() StateKey 修正） |
| 3 | ✅ | `CalcKeys`/`CalcOctets` 讀計數器 O(1)；`ThresholdBalance` method；`dev_assert` build tag |
| 4a | ✅ | `host_call_general.go`: read/write/info 切到 globalKV |
| 4b | ✅ | `host_call_accumulate.go`: new/transfer/eject/query/solicit/forget/provide；`finalizeNewAccount` 解決 new() 的 StateKey 問題 |
| 4c | ✅ | PVM 結構體 `StorageKeyVal` 欄位移除（與 7.5d 合流：合流後 PVM 不再持有 `StorageKeyVal`，全部改走 globalKV 路徑） |
| 5 | ✅ | `HistoricalLookup`/`FetchCodeByHash`/`ValidatePreimageLookupDict` 簽名加 `serviceID`；新增 `AddPreimage` method |
| 6 | ✅ | `StateEncoder` 直接走 globalKV |
| 7 | ✅ | `StateKeyValsToState` 三類桶 + `IsPreimage` 回傳 hash；`IsLookup` 移除 |
| 7.5 | ✅ | `unmatchedKeyVals` 整條 pipeline 移除（`StateKeyValsToState` → `(State, error)` + ChainState 6 個 method + state-root append 全清）|
| 8a | ✅ | `ServiceAccountDerivatives` 移除 |
| 8b | ✅ | `unmarshal_json.go` 走 globalKV（dual-write legacy maps 給 jamtests wire format）|
| 8c-f | ⏳ | **未做** — jamtests `ServiceAccount.Encode` wire format 編 raw storage key，無法從 hash 反推，需重新設計 JAM test-vector binary format（→ Phase 1 收尾） |
| 9 | ✅ | 5 個 test fixture 改用 `MigrateLegacyMapsToGlobalKV` |
| **Bug fix** | ✅ | GP §9.8 `ThresholdBalance` 正確實作 `storage − a_f` |

### 12 commits on `feat/global-kv-refactor`

```
b6b0c171 docs(Todo): mark Step status, record fuzz + benchmark results, fix typo
642999b6 fix(service-account): subtract GratisStorageOffset in ThresholdBalance (GP §9.8)
d1eb0163 test: migrate legacy-map fixtures onto globalKV (Step 9 partial)
4f91eddf refactor(types): retire ServiceAccountDerivatives + JSON unmarshal via globalKV (Step 8a/8b)
c824060e refactor(state): retire unmatchedKeyVals fallback pool (Step 7.5)
c56422cb fix(accumulation): InsertPreimageMeta in preimage integration + keep delta on error
a00b5ba0 fix(service-account): DeepCopy + new() StateKey now carry globalKV / counters
07744327 refactor(merklization): sync encoder + three-bucket decoder (Steps 6 + 7)
2cb8da6d refactor(accumulation): switch preimage flows to globalKV (Step 5)
0e952ce0 refactor(pvm): switch host_call_accumulate.go to globalKV (Step 4b)
1308bca4 refactor(service-account): introduce globalKV foundation (Steps 1-3 + partial 4a)
dc7842b3 (baseline)
```

### 最終測試結果（與 baseline byte-for-byte 對齊）

| 測試集 | 結果 |
|--------|------|
| Minifuzz × 4 suites (storage / storage_light / fallback / safrole) | 102/102 pairs each, all PASS |
| Picofuzz × 4 suites | PASS, no MISMATCH/FAIL |
| jam-conformance 0.7.2 traces | **282 / 282 PASSED** |

### Picofuzz p99 latency 對比

| Suite | Baseline | Refactor | Δ |
|-------|----------|----------|---|
| storage | 220.5 ms | 208.8 ms | **−5.3%** |
| storage_light | 81.0 ms | 68.5 ms | **−15.4%** |
| fallback | 36.4 ms | 39.1 ms | +7.5% |
| safrole | 41.3 ms | 46.7 ms | +13.0% |

Storage-heavy workloads 顯著加快。

> **fallback / safrole 退步分析**：推測為 `decodePreimageMetaValue` 的 JAM
> 反序列化開銷 — 原本 `LookupDict` 直接是 Go 結構（in-memory），改為
> `globalKV[k]` 後每次讀取需 `jam.Decode` 一次 `[]byte → TimeSlotSet`。此
> 退步在 storage-heavy workload 被 globalKV 統一帶來的加速抵銷，但在
> preimage-meta-heavy workload（fallback / safrole）則顯現。Phase 2.3 增量 merklize 對 storage-heavy workload 收益最大；fallback / safrole
> 的反序列化開銷是 Phase 1 globalKV 重構的固有成本，需要另外的 leaf-level 快取
> 擴展（讓 `globalKV[k]` 的 `TimeSlotSet` 反序列化結果也快取）補回，未來再評估。

---

## Phase 1 收尾 — `StorageDict` / `LookupDict` 完全 retire

> 對應原 Phase 1 Step 8c-f（上表第 8c-f 行 ⏳）。Phase 2 啟動前先收掉，
> 讓 `ServiceAccount` struct 變乾淨，後續 dirty key tracking 設計時少兩個欄位要兼顧。

### 目標

把 `ServiceAccount` 上殘留的 `StorageDict` / `LookupDict` 完全移除。Phase 1
runtime 已經完全不讀這兩個 map，但它們仍被 `Clone()` dual-write、被 jamtests
`ServiceAccount.Encode` wire format 使用。

### 風險

- `ServiceAccount.Encode` 的 wire format 編 raw storage key，hash 過後無法反推。
  需要重新設計 JAM test-vector binary format。
- 如果 upstream jam-conformance 的 test vector format 有硬規範，可能需要先跟
  upstream 對齊再動。

### 子步驟

- [ ] 確認 upstream jam-conformance test vector 的 wire format 是否有硬規範；
  若有，需先跟 upstream 對齊再動
- [ ] **escalation**: upstream 確認超過 2 週未回覆 → 改 plan B（選項 C）：
  保留 `StorageDict`/`LookupDict` struct 欄位 + 繼續 dual-write，但 runtime
  一律走 `globalKV`（`Clone()` / `assert_consistency.go` / `ValidatePreimageLookupDict`
  等全部改為走 globalKV）。wire format 維持向後相容。待 upstream 回覆後
  再完成最後一步（拔除 struct 欄位 + 拔除 dual-write）
- [ ] 重新設計 jamtests `ServiceAccount.Encode` binary format（不再依賴 raw key）
- [ ] `unmarshal_json.go` 拔掉 dual-write legacy maps
- [ ] `Clone()` 拔掉 `cloneMapOfSlices(sa.StorageDict)` / `cloneLookupDict(sa.LookupDict)`
- [ ] `ServiceAccount` struct 移除 `StorageDict` / `LookupDict` 兩個欄位
- [ ] `NewServiceAccount()` 建構子移除 `LookupDict` / `StorageDict` 初始化
- [ ] `cloneLookupDict` helper 移除
- [ ] `MigrateLegacyMapsToGlobalKV` 移除（拔除後變 dead code；原本被 5 個 test
  fixture 使用，這些 fixture 需連動改為直接用 `InsertStorage` / `InsertPreimageMeta`）
- [ ] `assert_consistency.go` 的 `len(LookupDict)` / `len(StorageDict)` 交叉驗證
  — 優先砍掉（Phase 1 runtime 已不讀這兩個 map，assert 驗的是即將刪除的
  不變式，保留反而增加混淆）；除非有其他 assertion 路徑仍需保留
- [x] ~~`ValidatePreimageLookupDict`（`service_account.go:60-89`）改為走 `globalKV` lookup~~
  — 已完成（`existsInLookupDict` 已走 `account.GetPreimageMeta(stateKey)`）
- [ ] ~5 個 test 檔案中直接 access `StorageDict` / `LookupDict` 的 fixture 替換
  （`service_account_test.go` / `accumulation_test.go` / `ce134_test.go` /
  `work_package_controller_test.go` / `state_serialize_test.go`）
- [ ] **wire format 相容性驗證**：
  - 跑 jamtests 全套（不只 282 conformance），確認所有 test vector 仍能正確解析
  - 若 upstream 提供 binary fixture，用改後的 decoder 解這些 fixture，
    確認 round-trip byte-for-byte 對齊
  - 若 plan B（保留 dual-write）啟動，需文件化「runtime 走 globalKV / wire
    走舊 maps」的差異，並加 invariant test 守住兩邊一致
- [ ] 跑三套 fuzz + 282 conformance traces，確認 byte-for-byte 對齊

### 預估規模

- **修改**：~300-400 行（含測試 fixture 連動、建構子、encode/decode 改動、
  assertion 路徑、8 個 test 檔案的 fixture 替換）
- **新增**：jamtests wire format 重設計可能引入 ~50 行
- **時間**：3-5 天（含 jamtests 相容性驗證；若需跟 upstream 對齊可能更久）

---

## Phase 2 — Trie Store 持久化 + 增量 Merklize

### 目標

兩件事：

1. **Trie Store 持久化**：把 merklize 過程產生的 trie node 寫進 PebbleDB，加上
   refcount 機制讓不同 block 共享相同 subtree
2. **增量 Merklize**：每個 block 只重算被修改過的 dirty path，未動的 subtree
   直接 reuse 上一個 root 的 node hash

state 主結構**維持 Phase 1 現況**（`globalKV` 仍是 `map[StateKey][]byte`
in-memory；`StateCommit` 仍 snapshot 到 PebbleDB）。Phase 2 不改 `ServiceAccount`
的 read/write API，業務邏輯（PVM host call / accumulation）零改動。

### 既有資源

- `internal/database/database.Database` interface 已提供
  `Has/Get/Put/Delete/NewBatch/NewIterator`，Phase 2 直接使用，
  **不再引入新的 KVBackend 抽象**
- `internal/database/provider/pebble/` 已實作完整的 Pebble provider
- `internal/utilities/merklization/merklization.go` 的 64-byte node 編碼
  （Branch / Embedded leaf / Regular leaf）已 GP-conformant，Phase 2 不重寫
- `internal/blockchain/key_level_cache.go` 已有 **leaf-level 快取**
  （`KeyLevelCache`：per-StateKey 快取 `(valueHash, leafHash)`，value 沒變就
  reuse leaf hash，避免重複計算）

### 設計守則（SRP / DRY）

1. **TrieStore 職責單一**：只負責 trie node 的持久化、refcount 管理、增量
   merklize 演算法。不持有 state、不知道 block 概念。
2. **dirty key set 由 prior/current StateKeyVals diff 產生**：
   - `PersistStateForBlock` 在 StateEncoder + sort 之後，與 prior sorted
     key-vals（從 `cs.repo` 取）做 merge-scan diff → 產生 dirty key set
   - diff 成本為 O(n) byte comparison，遠比 O(n) Blake2b hash 便宜
   - 覆蓋所有 4 類 state 變更（consensus keys 1-16 / delta1 service metadata /
     delta2-4 globalKV / delta3 PreimageLookup），不依賴 STF call chain
     的 wrapper 注入
   - **前提**：diff 正確性依賴 `cs.recentStateRoots[last]` 指向真正的
     prior state root。`RestoreBlockAndState` / `RestoreStateFromSnapshot`
     後必須重置 `recentStateRoots`（Phase 2.2 子步驟）
   - diff 機制本身是 stateless 的純函式（輸入 priorKVs + currentKVs，
     輸出 dirtyEntries），跟「ChainState 是否 singleton」無關。Phase 3
     改為 context-scoped 後，diff 機制不需改
3. **`ServiceAccount` 不知道 dirty tracking**：diff 在 `PersistStateForBlock`
   層做，`ServiceAccount` API 完全不受影響
4. **StateEncoder 變更時 diff 邏輯必須同步驗證**：
   - K1 方案 A 的正確性依賴 StateEncoder 輸出穩定、確定性
   - GP 升級 / 新增 state 欄位時，必須跑 282 conformance 確認 diff
     正確覆蓋新欄位
   - CI 加 invariant test：新增 `state_encoder_invariant_test.go`：
     (a) 固定 state fixture 連跑 100 次 `StateEncoder`，sorted 後 byte-for-byte 對齊
     (b) `state1 == state2` → sorted `StateEncoder(state1) == StateEncoder(state2)`
     (c) 加進 PR CI gate
5. **全量 vs dry-run 共用核心演算法**（callback strategy 差異：nil / 寫 batch）。
   **增量 merklize 則是根本不同的 traversal 策略**（從 DB 讀舊 trie node、沿
   dirty key 的 bit path 往下走），跟全量路徑幾乎沒有共用 code。兩者的正確性
   需獨立驗證。

### `KeyLevelCache` 與 TrieStore 的整合策略

既有的 `KeyLevelCache`（`internal/blockchain/key_level_cache.go`）做的是
**leaf-level** 快取：一個 StateKey 的 value 沒變就 reuse 已算過的 leaf hash。
Phase 2 引入 TrieStore 後兩者的關係：

- **Phase 2.1 / 2.2（全量 merklize + trie node 持久化）**：`KeyLevelCache`
  **保留**，繼續發揮 leaf-level 加速。TrieStore 管的是 branch / internal node
  層級的持久化，兩者不重疊、可疊加。
- **Phase 2.3（增量 merklize）**：增量 merklize 從 dirty key set 出發，
  只重算 dirty path 上的 node。`KeyLevelCache` 的 leaf cache 可作為
  「確認某個 leaf 是否真的 dirty」的快速 check（value hash 沒變 → 跳過），
  兩者互補。
- **未來如果 TrieStore 的 branch-level refcount 已足夠取代 `KeyLevelCache`
  的收益**，可考慮移除 `KeyLevelCache` 以減少重複快取。但這不在 Phase 2
  範圍內，待 Phase 2.3 benchmark 結果出來再評估。

### 性能預期與動機釐清

- **Phase 2.1**（TrieStore API）= **對齊對照組已有的 API**。補上
  `MerklizeAndCommit` / `MerklizeOnly` / refcount / `DeleteTrie` 等基礎設施。
  不改任何 fuzz/conformance 路徑，picofuzz p99 無變化。
- **Phase 2.2**（fuzz 路徑切到 `MerklizeAndCommit`）= **領先對照組的設計選擇**。
  對照組 conformance 路徑用 `MerklizeStateOnly`（不寫 trie node）；我們改為
  寫 trie node。picofuzz p99 預期維持或略慢 ~5%（多了 batch.Put + refcount）。
  **Phase 2.2 的核心動機是為 Phase 2.3 鋪路**（trie node 進了 DB 後增量
  merklize 才能從 prior root traverse）。磁碟層 sharing 在短 fuzz session
  幾乎無收益，long trace 才有觀察價值。
  **注意**：如果 Phase 2.3 PoC-1 結果 NO-GO（merklize < 30%），需評估是否
  rollback Phase 2.2 的 trie writing（改回 `MerklizeOnly`），避免只有成本
  沒有收益。
- **Phase 2.3**：增量 merklize 才是 wall-clock 性能主菜。blocks with small
  delta 的 merklize 時間從 O(n) → O(log n × delta)。

### Prefix Schema

```
prefix 0x03: trie node               key = 0x03 || nodeHash[1:32]    (31 bytes)
prefix 0x04: trie node value (>32B)  key = 0x04 || valueHash[0:32]   (32 bytes)
prefix 0x05: trie node refcount      key = 0x05 || nodeHash[1:32]    value: uint64 LE
```

跟 `chain_state` 等既有持久化資料共用同一個 PebbleDB instance
（透過 `cs.persistentRepo.Database()`），不另開 DB。

> **Prefix 命名慣例注意**：既有 prefix（`store/prefix.go`）使用 ASCII 文字
> （如 `"sr:"`, `"sd:"`, `"h:"`），新的 trie prefix 使用 raw byte
> （`0x03`/`0x04`/`0x05`）。兩套慣例在同一個 DB 裡不衝突（ASCII 最小值
> `0x62` > `0x05`），但共存需知曉。注意：上述 prefix 值跟對照實作的
> iota-based prefix 不同（對照實作 trieNode=0x04, value=0x05, refcount=0x06）。
> 正確性不受影響（兩邊 DB 獨立）。

### 子步驟

> **可與 Phase 1 收尾並行開工**：Phase 2.1 與 D1 無技術依賴。D1 第一條
> （upstream 確認）發出後，不必等回覆即可啟動 Phase 2.1。

### 開工前裁決清單

進入 Phase 2.1 前必須裁決：
- [ ] D-1: refcount 更新策略（推薦 A: batch 外逐個，理由：對齊對照實作）
- [ ] D-2: persistentRepo state data 清理（推薦 a: evict 時刪一行）
- [ ] D-3: ResetInstance trie 清理（推薦 a: DeleteRange trie prefix。~~選項 c 不可行~~）
  — 提前到 Phase 2.1 是因為 Phase 2.1 已經會生 trie node，清理策略要先定
- [ ] D-4: refcount 偏高處理（推薦：文件化接受 + DeleteTrie 收尾）

### 最小執行清單 (happy path)

```
Week 1   : Phase 2.1 (callback 改造 + TrieStore + MerklizeOnly)
Week 2   : Phase 2.2 (整合 + trie 生命週期 + restoreWithState 重置 + 消除三重計算)
Week 3   : PoC-1 (1 天) → GO/NO-GO 決策
Week 3-4 : PoC-2 (2-3 天) — 僅在 PoC-1 ≥ 30% (GO 或 Conditional GO) 時執行
Week 4-8 : Phase 2.3a (branch cache，純記憶體) → 交叉驗證
Week 8-10: Phase 2.3b (dirty path + 持久化) → 三套測試 byte-for-byte 對齊

Contingency: Conditional GO (30-40%) → 只做 2.3a，估 4 週
             8 週後 mismatch 未解 → 降級 2.3a only
             10 週後整合失敗 → Phase 2.3 延後，保留 2.1+2.2 收益
```

### 子步驟

**2.1: TrieStore + refcount + MerklizeOnly**

- [ ] **前置（獨立 commit / PR）：`merklize()` callback 改造**
  SRP 角度這是 pure refactor（signature change），跟 TrieStore feature 分開。
  改完跑三套測試 byte-for-byte 對齊 Phase 1（callback 全傳 nil 時等效原行為）。
  - 現有 `merklize(entries []StateKeyVal, depth int) OpaqueHash` 沒有 callback
    參數且不回傳 error
  - 定義 `type TrieNode [64]byte`，掛 `IsLeaf()` / `IsBranch()` /
    `GetBranchHashes()` / `GetLeafKey()` / `GetLeafValue()` / `GetLeafValueHash()`
    等方法（Phase 2.3 增量 traversal 會直接用到）。注意：`GetBranchHashes()`
    回傳的 left hash byte 0 的 MSB 已被 mask（`& 0x7F`），不影響 DB lookup
    （key 用 `hash[1:32]`），但 `left[0] ≠ 原始 child hash byte 0`
  - 改為 `merklize(entries, depth, storeNode func(OpaqueHash, TrieNode) error,
    storeValue func([]byte) error) (OpaqueHash, error)`
  - leaf 路徑需在 `encodeLeafNode` 之前先呼叫 `storeValue(value)`
    （僅 value > 32B 且 callback 非 nil）
  - **branch 路徑**需在 `encodeBranchNode` + hash 之後呼叫
    `storeNode(hash, branchNode)`（僅 callback 非 nil）— 缺了這步
    TrieStore 只存 leaf、不存 branch，增量 merklize 無法 traverse
  - 確認 leaf + branch 兩個路徑都有 `storeNode` 呼叫
  - `merklizeWithCache()` 的 branch 路徑同樣需要加 `storeNode(hash, branchNode)`
    callback（與 `merklize()` 改造保持一致，兩個函式在同一檔案、結構相同）
  - 上層 `MerklizationSerializedState` / `MerklizationSerializedStateWithCache`
    跟著改 signature + 回傳 error
  - `chain_state.go` 的 `merklizeWithKeyCache` 及所有呼叫者適配
  - `merklizeWithKeyCache` 新增 `storeNode` / `storeValue` 可選參數
    （或拆成 dry-run 版本 + 帶 callback 版本）。`ComputeStateRootWithCache`
    維持呼叫 dry-run 版本（callback = nil）
  - 預估 2-3 天（含三套測試驗證 + error propagation 連動）
- [ ] 新增 `internal/storage/trie_store.go`（或合適位置）。TrieStore 作為
  ChainState 的 field，在 `GetInstance().initOnce.Do` 內初始化
  （傳入 `persistentRepo.Database()`）。
  **Lifecycle 規範**：TrieStore 不持有 goroutine、不持有 PebbleDB snapshot、
  不開 iterator。`ResetInstance()` 完整清單（4 件事）：
  1. GC 舊 TrieStore Go object（自動）
  2. `DeleteRange` 清 DB 中的 trie node（D-3a 落地）
  3. Reset `recentStateRoots = nil`（避免新 session 第一個 block 用到舊 root 做 diff）
  4. Reset `lastCommittedStateRoot = StateRoot{}`（zero hash，等 `SetState` / `SeedGenesisToBackend` 重新填入）
- [ ] 實作 API：
  - `MerklizeAndCommit(pairs [][2][]byte) (StateRoot, error)` — 算 root 並把
    每個 node + leaf value 寫進 PebbleDB，refcount += 1
  - `MerklizeOnly(pairs [][2][]byte) (StateRoot, error)` — dry-run，不寫 DB。
    職責邊界：`MerklizeOnly` 是 TrieStore 內部用（Phase 2.3 增量 merklize 的
    全量 fallback）+ callback 全 nil 的純計算；**不走 `KeyLevelCache`**
    （TrieStore 不該知道 cache 存在）。既有的 `ComputeStateRootWithCache`
    維持 fuzz path 用（走 `KeyLevelCache`），兩者語意一致但呼叫者不同
  - `GetNode(hash OpaqueHash) (Node, error)` / `GetNodeValue(node Node) ([]byte, error)`
  - `TrieExists(rootHash OpaqueHash) (bool, error)`
  - `IncreaseNodeRefCount(hash OpaqueHash) error` / `DecreaseNodeRefCount(hash OpaqueHash) (uint64, error)`
  - `GetNodeRefCount(hash OpaqueHash) (uint64, error)` — debug / test 用
  - `DeleteTrie(rootHash OpaqueHash) error` — 遞迴刪除，refcount 為 0 才真刪
- [ ] 跟 `internal/utilities/merklization` 用改造後的 callback pattern 對接
  （`encodeBranchNode` / `encodeLeafNode` 不重寫）
- [ ] Refcount 用 `uint64 LE` 編碼；DeleteTrie 用 root forceDelete=true /
  child forceDelete=false 的不對稱遞迴
- [ ] **refcount 更新策略決策**：
  - 選項 A（照對照實作）：refcount 在 `batch.Commit()` 後逐個更新。簡單，
    但 crash 後 refcount 可能不一致。fuzz 場景可接受（`ResetInstance` 重來）
  - 選項 B（改進版）：refcount 也放進 batch（需要 indexed batch 語意）。
    原子性更強，但 PebbleDB 支援度需確認
  - **Refcount 更新順序**：merklize 全部成功 → `batch.Commit()` 成功 →
    才開始呼叫 `IncreaseNodeRefCount`。任何前一步失敗 = 不更新 refcount。
    對齊對照實作 `Trie.MerklizeAndCommit` 的順序
  - 推薦：Phase 2.1 先用選項 A（對齊對照實作，降低風險）。**注意**：對照實作
    是 `vfs.NewMem()` in-memory FS，crash recovery 不適用。我們的 on-disk +
    NoSync 環境下，選項 A 的 refcount-mismatch 風險真實存在，但 fuzz 場景
    每次 crash 都會 `ResetInstance()` 重來而踩不到。**Phase 3 production
    node 啟動時必須升級為選項 B**
- [ ] Unit test 矩陣覆蓋：建 trie / get node / refcount 增減 / delete trie /
  多 trie 共享 node 的 refcount 正確性 /
  **共享 root guard test**：兩棵 trie 共享同一個 root hash，對其中一棵呼叫
  `DeleteTrie`，驗證另一棵的 root node 仍可正常 `GetNode`（不 dangling）
- [ ] **callback 一致性 guard test**：新增 `merklize_callback_test.go`，
  用相同 input 分別跑 `merklize` 和 `merklizeWithCache`，assert 兩者：
  (a) 產出相同 state root (b) callback 被呼叫的 (hash, node) 集合完全相同
  （順序可不同）。防止未來修改一方時另一方失聯

**2.2: 整合到 StateCommit**

目前有六條 merklize / state commit 路徑需要釐清：

| 方法 | 行為 | TrieStore 處理方式 |
|------|------|-------------------|
| `PersistStateForBlock` | serialize → merklize → save（被 `StateCommit()` 呼叫）| **改這裡**：`merklizeWithKeyCache` → `trieStore.MerklizeAndCommit()` |
| `StateCommit()` | 呼叫 `PersistStateForBlock` + 存 block + 記 ancestry | 上層流程不動，自動走 TrieStore |
| `StateCommitWithPreComputedState()` | 呼叫者已算好 stateRoot + keyVals，只做 save | **不走 TrieStore**（trie node 不寫 DB）。此路徑僅做 state snapshot 持久化。目前無 production caller（dead path，API 保留供未來使用）|
| `ComputeStateRootWithCache` (public) | 被 `fuzz/service.go` 呼叫（**消除冗餘後 L102/L129/L190 被移除**；剩餘呼叫者為 `BuildStateRootInputKeyValsAndRoot`（genesis）） | **維持現有 `merklizeWithKeyCache` 路徑**（等效 dry-run，不寫 trie node）。`KeyLevelCache` 行為不變 |
| `BuildStateRootInputKeyValsAndRoot` | fuzz 預計算路徑使用 | 同上，維持現有路徑不動 |
| `SeedGenesisToBackend` | `chain_state.go:850-886`，呼叫路徑 5 → 直接寫 `persistentRepo` | 不走 TrieStore（genesis 初始化，Phase 2.3 fallback 處理） |

- [ ] `PersistStateForBlock` 內的 `merklizeWithKeyCache` 替換為
  `trieStore.MerklizeAndCommit()`
- [ ] `StateCommit()` 和 `StateCommitWithPreComputedState()` 的上層流程相應適配
- [ ] 從第二個 block 起，未變的 trie node 自動 share（refcount += 1）
- [ ] 釐清 `KeyLevelCache` 與 TrieStore 的互動：leaf cache 保留、branch
  node 由 TrieStore 管理，兩者疊加不衝突
- [ ] **設計 trie node 生命週期**：ring buffer evict（`recentStateRoots`）
  目前只刪 `cs.repo` 的 state data。**Phase 2.2 後必須一併呼叫
  `trieStore.DeleteTrie(evictedStateRoot)`**，否則 trie node 永遠
  refcount += 1 / 永不刪 / DB 無上界膨脹。
  - **兩處 evict 位置都要加 DeleteTrie**：`PersistStateForBlock`
    (`chain_state.go:649-655`) 和 `StateCommitWithPreComputedState`
    (`chain_state.go:456-462`)
  - `StateCommitWithPreComputedState` 路徑不寫 trie node，但 evict 的
    stateRoot 可能由 `PersistStateForBlock` 寫入。`DeleteTrie` 對不存在的
    root 會 silent return（`deleteNode` 在 `ErrNotFound` 時 return nil），
    統一加不會有副作用
  - 需確認 `stateRoot` 與 trie root hash 的 type 對齊
  - DeleteTrie 邏輯天生支援共享 subtree（refcount 減到 0 才真刪），落地阻力最小
- [ ] **`restoreWithState` 重置 `recentStateRoots`**（~15 行）：
  fork handling 後 `recentStateRoots` 仍指向舊 fork 的 state root，
  導致後續 diff 取到錯誤的 prior state → 增量 merklize state root mismatch。
  修正：`restoreWithState` 在 restore 完成後取得 restored block 的 stateRoot，
  重置 `cs.recentStateRoots = []StateRoot{restoredStateRoot}`
  **+ `cs.lastCommittedStateRoot = restoredStateRoot`**。
  **stateRoot 取得 fallback chain**：
  1. 優先從 `cs.repo` 取（memory，ring buffer 未 evict 時）
  2. 不在 memory → 重置為 zero hash（`persistentRepo` 沒有 headerHash→stateRoot
     映射，見已知限制 L405-410）
  3. deep fork restore（超過 MaxLookupAge=24）：stateRoot 必為 zero hash，
     所有 fork restore 後的 block 走 fallback 全量。當前 fuzz 不踩此邊界
  同時適用於 `RestoreBlockAndState`、`RestoreStateFromSnapshot`
  和 **`ResetInstance`**（fuzz `SetState` 時呼叫，如果不 reset，
  後續 `PersistStateForBlock` 會用到舊 root 做 diff → 增量 merklize mismatch）
  - **共享 root guard**：如果 evict 的 stateRoot 跟 ring buffer 中其他 entry
    相同（兩個 block 產生相同 state，概率極低但 fuzz 空 block 可能觸發），
    需 skip `DeleteTrie` 或先 `GetNodeRefCount(evictedRoot)`，refcount > 1 時
    只 `DecreaseNodeRefCount` 不做 `DeleteTrie`。
    根因：`deleteNode(root, forceDelete=true)` 在 refcount > 0 時仍會刪除
    root node 本體（只有 children 受 `forceDelete=false` 保護），導致共享
    root 的另一個 trie 變成 dangling reference
  - **evict 來源混合**：ring buffer 裡混合了三種路徑寫入的 stateRoot。
    只有 `PersistStateForBlock` 路徑的有對應 trie node；
    `StateCommitWithPreComputedState` / `SeedGenesisToBackend` 路徑的
    evict 呼叫 `DeleteTrie` 會 silent return（`ErrNotFound`），無副作用
- [ ] **已知限制（文件化）**：`PersistStateForBlock` 不寫
  `persistentRepo.SaveStateRootByHeaderHash`（headerHash→stateRoot 映射不在
  disk）。ring buffer evict + DeleteTrie 後，被 evict 的 state 完全無法從
  disk 恢復。目前 fuzz 不踩此邊界（fork depth < MaxLookupAge=24）。
  如果未來需要 deep fork recovery，需先修正 `PersistStateForBlock` 也寫
  `persistentRepo.SaveStateRootByHeaderHash`
- [ ] **三套 fuzz + 282 conformance 全綠**，state root byte-for-byte 對齊 Phase 1
- [ ] 觀察 DB 大小成長曲線，應該是 sub-linear（共享 node 越來越多）。
  具體方法：跑 long trace（500+ blocks），每 100 blocks 記錄
  `db.Metrics().DiskSpaceUsage()`，畫曲線確認 sub-linear
- [ ] ~~**A7 benchmark**~~ — **延後**：`pebble.NoSync` vs `pebble.Sync` 的切換
  在 fuzz/conformance 階段完全不影響（crash 後 `ResetInstance` 重來）。
  未來做 production node 時再處理（一行改 `pebble.go:31` 或 config-driven 切換）
- [ ] **落地 D-2 裁決（persistentRepo state data 清理）**：根據開工前
  裁決的策略（推薦 a），在兩處 evict 位置加
  `cs.persistentRepo.DeleteStateData(db, evict)`。不加 = long trace fuzz
  磁碟 unbounded 成長
- [ ] **釐清 `ResetInstance()` trie 清理策略**：PebbleDB instance 透過
  `sync.Once` 建立，`ResetInstance()` 後舊 trie node 仍在 DB 中。需決定：
  (a) `ResetInstance()` 時 `DeleteRange` trie prefix（`0x03/0x04/0x05`）—
  注意：`Database` interface 目前沒有 `DeleteRange` 方法，需擴充 interface
  或用 Iterator + Batch + Delete 手刻（fuzz reset 頻率低，效率差異可接受）。
  注意：擴充 `Database` interface 會牽動 2-3 個 provider（pebble / memory / redis），
  推薦先用 Iterator + Batch + Delete 手刻，未來 production 再升級為 interface 原生 `DeleteRange`
  (b) 改為每 session 新建 PebbleDB（對齊對照組做法）— 注意：需打破
  `getPersistentDatabase()` 的 `sync.Once`（`chain_state.go:60-83`），
  改為 `sync.Mutex` + 條件重建，或在 `ResetInstance()` 時 close 舊 DB + 重建。
  此改動影響 chain/header 持久化，**推薦選 (a) 避開此問題**
  ~~(c) 文件化接受~~ — **不可行**：Phase 2.2 落地後 trie node 持久化會跟殘留
  的跨 session trie node 複合膨脹，disk 無上界成長。只能選 (a) 或 (b)
  注意：跨 session refcount 疊加 — `getPersistentDatabase` 用 `sync.Once`，
  `ResetInstance` 後舊 session 的 trie node 仍在 DB。新 session 寫入相同
  hash 的 node 時 refcount 疊加（1+1=2）。後續 `DeleteTrie` 只 decrease
  不真刪（refcount > 0）→ trie node 逐 session 膨脹。選項 (a) 的
  `DeleteRange` 會一併清理此問題
- [ ] **⚠️ 消除 `fuzz/service.go` 三重計算（重點審查項）**：
  目前 `ImportBlock` 每個 block 跑 3 次 serialize + 3 次 merklize：
  L98+L102（prior state root，dry-run）、L123+L129（posterior state root，
  dry-run）、L132 `StateCommit`（commit）。其中前 2 次是冗餘的：
  - **L102 冗餘**：prior state root = 上一個 block 的 `StateCommit` 已算過的結果。
    解法：`ChainState` 新增 `lastCommittedStateRoot` field，`StateCommit` 完成時
    寫入，`ImportBlock` 直接讀（O(1)），不再呼叫 L98 `StateEncoder` + L102
    `ComputeStateRootWithCache`
  - **L129 冗餘**：posterior state root = L132 `StateCommit` 內部會再算一次。
    解法：改 `StateCommit()` signature 為 `StateCommit() (StateRoot, error)`，
    回傳已計算的 stateRoot，`ImportBlock` 直接用回傳值
  - **改後 `ImportBlock` 流程**：
    ```
    latestStateRoot = cs.lastCommittedStateRoot           // O(1)
    stf.RunSTF()
    if err { return latestStateRoot, err }                // protocol error 回傳 prior root
    stateRoot, err = cs.StateCommit()                     // 唯一的 serialize + merklize + save
    return stateRoot, nil
    ```
  - **效果**：從 3 次 serialize + 3 次 merklize → 1 次 serialize + 1 次 merklize。
    fuzz wall-clock 預期省 ~60%。Phase 2.3 增量 merklize 的收益不再被冗餘
    計算稀釋
  - **工程量**：~25 行（`StateCommit` 改回傳值 + `lastCommittedStateRoot` field +
    `ImportBlock` 簡化 + `SetState` 簡化 + `restoreWithState` / `ResetInstance`
    同步重置 `lastCommittedStateRoot`）
  - **風險**：低。`StateCommit` 內部已算好 stateRoot（`PersistStateForBlock`
    回傳值），只是目前沒回傳給 caller
  - **效果補充**：SetState 路徑同樣由 2 次降為 1 次 serialize + merklize
    （與 ImportBlock 一致）
  - **前置條件**：
    - `SetState` 流程結尾：改為 `stateRoot, _ := cs.StateCommit(); cs.lastCommittedStateRoot = stateRoot`，
      **移除原 L186-190 的二次 `StateEncoder` + `ComputeStateRootWithCache`**
      （同樣是冗餘計算，與 ImportBlock 三重計算同源）
    - `SeedGenesisToBackend` 結尾也必須設定 `lastCommittedStateRoot`（genesis state root），否則 genesis 後第一個 `ImportBlock` 在 protocol error 時回傳 zero hash
  - **與 H1 相容**：移除的是 L102 / L129 兩條 `ComputeStateRootWithCache`
    dry-run（本來就不寫 trie node），完全不影響 L132 `cs.StateCommit()` 內
    走的 `MerklizeAndCommit`（仍為每個 block 寫一次 trie node）

**2.3: 增量 merklize（性能領先設計，高風險）**

> **前置要求**：Phase 2.3 開工前須先完成以下 2 個 PoC，確認技術可行性後再正式估工時：
>
> | # | PoC | 目的 | 預估 |
> |---|-----|------|------|
> | PoC-1 | 用 5 個手工構造的 dirty key 場景，在紙上畫出 before/after trie 結構 + **用 picofuzz 4 個 suite（storage / storage_light / fallback / safrole）各跑 benchmark 分離 StateEncoder / sort / GetStateData(prior) / diffSortedKeyVals / merklize / SaveStateData 共 6 段耗時佔比**。**前提：必須在「消除三重計算 + trie writing 全部落地」（Phase 2.2 完成）後執行**。
  量到的 merklize 佔比 = 增量 merklize 替換 `MerklizeAndCommit` 後能省的成本上限
  （刻意包含 trie writing disk I/O，因為 Phase 2.3 增量也會省掉 unchanged subtree 的 batch.Put）（含 diff 成本，避免高估增量收益） | 驗證 branch collapse / split 邏輯 + **GO/NO-GO 決策**：4 個 suite 各自報告 merklize 佔比，以加權平均為準。
  加權平均 ≥ 40% → GO；30%-40% → conditional GO（只做 2.3a 評估後決定）；
  < 30% → 降級延後。即使加權 GO，若個別 suite < 20%（如 safrole），需在
  Phase 2.3 文件中註明「該 workload 增量收益有限」。
  **注意 KeyLevelCache 影響**：量測需獨立記錄 warm cache（跑 200+ blocks
  讓 cache 飽和）和 cold cache（每 block 前 `ClearKeyLevelCache()`）兩種
  情境。若 warm < 30% 但 cold > 50%，走 conditional GO | 1 天 |
> | PoC-2 | 實作 standalone `IncrementalMerklize` prototype（不接 TrieStore，用 in-memory map 模擬 node DB）+ **實作 `diffSortedKeyVals` prototype** | 驗證增量 traversal + collapse/split 端到端正確性 + diff 產生 dirty key 的正確性。**必須按遞迴分組（每層 branch 把 dirty keys 按 bit 分左右）而非逐 key 獨立處理，否則多 dirty key 場景會 O(k²) 重複讀寫**。測試案例必須涵蓋全部 4 類 state 變更：consensus keys 變更（Tau/Beta 等每 block 必變）、delta1 變更（service account metadata）、delta2/4 變更（globalKV insert/delete/modify）、delta3 變更（PreimageLookup blob）、mixed（同一 block 涵蓋多類）、leaf delete 後 branch collapse remaining leaf 的 StateKey 各 byte 非零驗證 re-encode 後 hash 不變（key 還原精度）、`DeferredTransfers` 造成多個 service 的 delta1 同時變更（確認 diff 覆蓋到所有受影響的 serviceID）| 2-3 天 |

- [ ] **實作 `diffSortedKeyVals(priorKVs, currentKVs) → []DirtyEntry`**：
  merge-scan 兩個已排序 `StateKeyVals` slice，產出 inserted / deleted /
  modified 三類 entries。O(n) byte comparison（遠比 Blake2b 便宜）。
  `DirtyEntry` 定義：`type DirtyEntry struct { Key StateKey; NewValue []byte; IsDelete bool }`
  （inserted/modified 攜帶新 value；deleted 時 NewValue=nil, IsDelete=true）。
  覆蓋所有 4 類 state 變更（consensus keys 1-16 / delta1 service metadata
  含 `DeferredTransfers` 造成的 balance/gas 變更 / delta2-4 globalKV /
  delta3 PreimageLookup），不依賴 STF call chain wrapper 注入
- [ ] **`PersistStateForBlock` 取 priorKVs**：
  - `priorStateRoot` 來源：`cs.recentStateRoots[len(cs.recentStateRoots)-1]`
    （前提：L1 的 `restoreWithState` recentStateRoots 重置已落地）
  - 如果 `recentStateRoots` 為空（genesis / SetState 後第一個 block）→
    fallback 到全量 merklize
  - 從 `cs.repo.GetStateData(priorStateRoot)` 取得上一次 commit 的 sorted
    key-vals（已排序、已存在，O(1) lookup + O(n) 讀取）
  - 如果 `cs.repo.GetStateData(priorStateRoot)` 找不到（prior state data
    已被 ring buffer evict 出 memory repo）→ **直接 fallback 全量 merklize**
    （不主動讀 `persistentRepo`；ring buffer evict 後資料已是冷 path，
    disk I/O + 反序列化成本可能抵銷增量收益）
  - Stress test 長 trace 時要量測 ring buffer evict 後的 fallback 比例
- [ ] **（可選）保留 globalKV dirty tracking 作為 debug assert**：
  在 STF host_call 層用 wrapper 追蹤 `InsertStorage` / `DeleteStorage` /
  `InsertPreimageMeta` / `DeletePreimageMeta` 的呼叫，debug 期間 assert
  wrapper 追蹤的 key ⊆ diff 產生的 key（不是正確性依賴，是 debug 工具）
- [ ] 增量 merklize 演算法：
  - 入口：`trieStore.IncrementalMerklize(priorRoot OpaqueHash, dirtyEntries []DirtyEntry) (newRoot, error)`
    — 不再需要 `currentKVs` 參數（dirty entries 已自包含 value + operation type）
  - 從 `priorRoot` 開始，**在每層 branch node 把 dirtyKeys 按 bit[depth]
    分組為 leftDirtyKeys 和 rightDirtyKeys，遞迴處理各自的 subtree**
  - 沒有 dirty key 的 subtree → reuse 舊 node hash（refcount += 1）
  - 有 dirty key 的 subtree → 遞迴往下走，直到 leaf 層再 re-encode
  - 此遞迴分組策略與全量 merklize 的 `partitionByBit` 語意等價，但
    操作對象是 trie node（從 DB 讀）而非 entries slice（在記憶體 partition）
  - 處理 **leaf delete → branch collapse** 邊界（最容易踩雷）：
    collapse 時需 `GetNode(childHash)` → `GetNodeValue(node)` 取得
    remaining leaf 的原始 key+value（regular leaf 情況），用
    `encodeLeafNode(key, value)` re-encode 並 `storeNode`。
    確認 re-encoded leaf hash = 原始 leaf hash（不變式）
  - 處理 **leaf insert → branch split**：single leaf 變 branch 時需
    在新 depth 重新 partition，確認 bit extraction 邏輯跟全量的
    `partitionByBit` 語意一致
  - 增量 merklize 重算 leaf hash 時，**同步更新 `KeyLevelCache`**
    （維持路徑 4/5 `ComputeStateRootWithCache` 的 cache hit rate）。
    增量路徑替換 `merklizeWithKeyCache` 後，如果不更新 cache，
    路徑 4/5 的 leaf cache 會逐漸失效
  - **`KeyLevelCache` invalidation on delete**：處理 dirty entries 時，
    若 `entry.IsDelete=true`，呼叫 `cs.keyLevelCache.Invalidate(entry.Key)`
    （需新增此方法）。避免後續同 key 重新 insert 時 cache 誤命中 stale
    leaf hash（目前「巧合正確」但 GP 升級後可能失效）
  - 注意：增量路徑是獨立的 traversal 策略（從 DB 讀舊 node、沿 bit path 走下去），
    **不能像全量那樣 in-place partition 整個 `StateKeyVals`**
- [ ] 建議分兩階段**落地**（在 PoC-1 / PoC-2 完成後）：
  - **2.3a**: branch hash cache（純記憶體，接 chain_state，先驗證 traversal 邏輯）
  - **2.3b**: dirty path reuse + 持久化結構（完整落地）
  - 注意：2.3a/b 是落地拆分，跟 PoC-1/2 的「設計驗證」是不同階段
- [ ] `PersistStateForBlock` 在 diff 產出 dirty keys 後走增量路徑，
  否則 fallback 走 2.1 / 2.2 的全量 `MerklizeAndCommit`。
  **明確的 fallback cases**：
  - `SetState`（genesis 初始化，無 prior key-vals 可 diff）
  - test runner 直接呼叫 `StateCommit()` 的路徑（如果 prior state data 不在 repo）
  - `RestoreBlockAndState` 後的重新 commit
  - **prior root 對應的 trie 不存在於 TrieStore**（已被 ring buffer evict
    或未持久化）— 此 case 在 trie 生命週期處置後可能更頻繁
  - **IncrementalMerklize traversal 過程中 `GetNode` 或 `GetNodeValue` 回傳
    `ErrNotFound`**（child node 缺失，trie 結構不完整，可能因 refcount
    不一致或跨 session 殘留）→ 中斷增量路徑，fallback 全量 + warning log
  - **fork restore 後第一個 block**：`restoreWithState` 重置 `recentStateRoots`
    後，第一個 block 走 fallback（全量 merklize），因為 prior trie node 可能
    已被前一個 fork 的 `DeleteTrie` 清掉。第二個 block 起恢復增量路徑
- [ ] **增量 vs 全量交叉驗證**：開發期間，每個 block 同時跑增量和全量路徑：
  - assert 兩者產出相同 state root
  - assert 增量寫入的 (hash, node content) 集合是全量寫入集合的子集
  - assert 增量 reuse 的舊 node 仍在 TrieStore 中且 hash 正確
  - **關閉條件**：連續 3 輪 282 conformance traces + 4 個 picofuzz suite +
    1 輪 5000-block stress test 全綠無 mismatch
  - 關閉後保留 flag（`crossValidateIncrementalMerklize bool`），預設：
    dev=on / CI=on / prod=off
  - 重大 STF / merklize / TrieStore 改動後必須重新打開跑一輪
- [ ] **三套 fuzz + 282 conformance byte-for-byte 對齊**
- [ ] Benchmark：picofuzz p99 對比
  - vs Phase 1（含 KeyLevelCache）— 預期增量 merklize 帶來大幅加速
  - vs Phase 2.2（全量 merklize + trie store，含雙重 merklize）— 演算法本身收益
  - 確認 KeyLevelCache 跟增量 merklize 疊加的實際收益（兩者是 leaf/branch
    不同層的快取）
  - 優先採信 Phase 1 vs Phase 2.3 的對比（消除雙重 merklize 後的真實收益）；
    Phase 2.2 vs Phase 2.3 僅供演算法分析參考（受 H1 雙重 merklize 影響）
  - 分析 picofuzz 4 suites 的 per-block delta size distribution（每個 block
    改了多少 key），確認增量 merklize 的理論加速倍數
  - **分離 benchmark**：分別測量 StateEncoder / sort / merklize / SaveStateData
    各自耗時佔比。確認增量 merklize 的加速沒有被其他 O(n) 步驟稀釋
  - **fuzz wall-clock benchmark**：除了測 `PersistStateForBlock` 內部改善，
    也要測 `ImportBlock` 整體 wall-clock。fuzz 每個 block 跑 3 次
    serialize + 2 次 dry-run merklize + 1 次 commit merklize，增量只省
    第 3 次 → 整體改善約為 `PersistStateForBlock` 改善的 ~1/3
- [ ] Stress test：長 trace（5000+ blocks）確認 DB 沒有 leak / 無止盡膨脹

### 預估規模

- **新增**：~600-800 行（`merklize()` callback 改造 + trie_store.go ~250 +
  `diffSortedKeyVals` ~100 + 增量 merklize 邏輯 ~200 + 測試 ~100）
- **修改**：~150-200 行（`merklize()` signature 改 → 所有呼叫者適配 +
  ChainState commit 路徑整合 + `PersistStateForBlock` 加 diff + priorKVs 取得）
- **STF 業務邏輯（PVM host call / accumulation / safrole / disputes 等）**：
  **0 行改動**（diff-based 方案不需修改 STF call chain signature）。
  但 ChainState commit path / fork restore path 約有 ~150-200 行改動，
  已列入修改規模估算
- **時間**：8-10 週（2.1: 1 週含 callback 改造 + TrieNode type / 2.2: 1 週
  含路徑釐清 + trie 生命週期 / 2.3: 6-8 週含 2 個 PoC + diffSortedKeyVals
  實作 + GetStateData 反序列化 + 增量演算法 + branch collapse/split 邊界
  debug + 交叉驗證）

### 風險

1. **增量 merklize 邊界（高風險）**：必須正確處理 leaf delete → branch collapse
   的情況，否則 trie 結構錯誤。建議拆 2.3a / 2.3b 兩階段驗證。
   具體技術難點：
   - JAM trie 的 partition-by-bit 在 leaf 數量改變時可能整個 subtree 重新分割
     （1 個 leaf 的 subtree 變 2 個、或變回 0 個時，branch 會 collapse / split）
   - 64-byte branch node 格式不含 path 資訊（`{left[0]&0x7F, left[1:32],
     right[0:32]}`），無法只看 branch node 得知它涵蓋哪些 bit prefix
   - branch node left child hash MSB 被 mask（`& 0x7F`），但 DB lookup key
     用 `hash[1:32]`（跳過 byte 0），所以 MSB mask **不影響** lookup 正確性。
     增量 traversal 時需注意：從 branch node 還原 child hash 後用 `[1:32]` 查 DB
   - 因此增量路徑需要額外的「prior root → dirty path」traversal 邏輯，
     而不是簡單的 hashmap lookup。此 traversal 邏輯的正確性是整個 2.3 的核心
   - 增量路徑是獨立的 traversal 策略，跟全量的 in-place partition 幾乎沒有
     共用 code
   - `ExecutionContext` 的 dirty set 是 per-STF-run transient 資料，
     **不構成 state 的一部分**（不違反附錄不變量 #8），
     commit 完即丟棄
2. **Refcount leak**：邏輯 bug 可能造成 refcount 漏減 → 浪費 DB 空間。
   stress test 必跑。
3. **PebbleDB NoSync 寫入丟資料（已知風險，現在不處理）**：既有 provider
   預設 NoSync (`internal/database/provider/pebble/pebble.go:31`)。Phase 2
   之後 trie node 持久化範圍變大，crash 可能丟一些 node。當前 fuzz 不踩這個。
   **啟動條件**：fuzz / production 出現 NoSync 丟資料案例 → 屆時統一處理
   （改 Sync / config-driven 切換）。參見 Phase 3.5。
4. **PebbleDB LSM compaction stall**：trie node 持續寫入會觸發 LSM
   compaction，可能在 compaction 期間造成 write stall，使 picofuzz p99
   反而變差。建議 Phase 2.2 整合後觀察 p99 tail latency，若有異常
   需調整 Pebble compaction 參數（`MaxConcurrentCompactions` /
   `MemTableSize` / `L0CompactionThreshold`）。
5. **`merklize()` callback 改造波及範圍**：改 signature 後，
   `MerklizationSerializedState*` / `merklizeWithKeyCache` / `chain_state.go`
   呼叫者都要適配。影響約 5-8 個函式。
6. ~~`ExecutionContext` 注入改 STF call chain~~ — **已消除**（K1 方案 A）：
   改用 diff-based dirty key 產生（設計守則 #2），不需要穿透 STF call chain。
   保留編號供歷史追溯。
7. **六條 merklize / StateCommit 路徑不一致風險**：`PersistStateForBlock` /
   `StateCommit` / `StateCommitWithPreComputedState` / `ComputeStateRootWithCache` /
   `BuildStateRootInputKeyValsAndRoot` / `SeedGenesisToBackend` 必須明確哪條走 TrieStore（MerklizeAndCommit）、
   哪條走 dry-run（MerklizeOnly）、哪條不走，否則 refcount 和 trie node
   持久化會不一致。
8. **KeyLevelCache 清空後 refcount 偏高**：leaf cache 在
   `MaxKeyLevelCacheSize` 滿時整個清掉（`chain_state.go:671-674`），之後同一
   leaf hash 重算時 `MerklizeAndCommit` 會 batch.Put 同 node + refcount += 1，
   refcount 偏高、加劇 trie node 膨脹。處理策略：commit 前 `TrieExists` 檢查，
   或文件明寫接受多算、由 DeleteTrie 收尾。
9. **同 batch 內重複 hash 的 refcount 多算**：`MerklizeAndCommit` 的 callback
   把 newNodes 存 slice 而非 set，重複 hash 會多次 `IncreaseNodeRefCount`。
   需決定：用 `map[Hash]int` 去重後一次性增量，或保持照抄但接受 refcount
   偏高。
10. **增量 merklize 加速被 O(n) 序列化稀釋**：`PersistStateForBlock` 的
    StateEncoder / sort / SaveStateData 仍為 O(n)，增量 merklize 只省
    merklize 步驟的開銷。如果 merklize 佔整個 pipeline 的時間不到 40%，
    實際 wall-clock 加速可能顯著低於預期。PoC-1 的耗時分佈分析是
    Phase 2.3 的 GO/NO-GO 門檻。
11. **對照實作為 in-memory FS，refcount 非 batch 安全**：對照實作用
    `vfs.NewMem()` 無真實 disk，refcount 在 batch 外更新的風險不存在。
    我們的 on-disk + NoSync 環境下，crash 後 refcount 不一致是真實風險。
    fuzz 場景 `ResetInstance()` 規避，production 啟用時必須處理（Phase 3）。
12. **`ResetInstance` 必須 reset `recentStateRoots` + `lastCommittedStateRoot`**：
    `restoreWithState` 的重置已在子步驟中，但 `ResetInstance`（fuzz `SetState`
    呼叫）也必須做完整 4 件事（見 Phase 2.1 TrieStore lifecycle 規範），
    否則新 session 第一個 block 的 diff / protocol error 回傳會用到舊值。
13. **`Database.DeleteRange` 不存在**：D-3(a) 的 trie prefix 清理需要
    擴充 `Database` interface 或用 Iterator + Batch + Delete 手刻。
14. **memory repo miss → disk fallback 的反序列化成本**：如果 prior state
    data 被 ring buffer evict 出 memory repo 但仍在 persistentRepo，
    `GetStateData` 走 disk 路徑的反序列化 + I/O 可能抵銷增量 merklize
    的收益。long trace stress test 必須量測 disk-fallback 比例及其對
    wall-clock 的影響。stress test 時用 `runtime/pprof` 或自製 timer
    獨立記錄 disk Get（含 I/O）與反序列化（CPU）兩段。

---

## Phase 3 — 未來規劃（不啟動）

> **狀態：規劃但不啟動。** 寫進文件僅作為未來路線圖記錄。啟動條件見下方。

### 內容

| 項目 | 描述 |
|------|------|
| Phase 3.1 | ChainState singleton → context-scoped（拔 200+ callsite，預估 3-4 週） |
| Phase 3.2 | Snapshot-based fork（O(1) fork branch，PebbleDB snapshot 整合） |
| Phase 3.3 | 平行 block validation（多 fork concurrent advance） |
| Phase 3.4 | 平行 accumulate（per-service trie subtree 隔離） |
| ~~Phase 3.5~~ | ~~PebbleDB write mode~~ — **已移出 Phase 3**（見 Phase 2.2 後的獨立 benchmark 任務）|

### 啟動條件（滿足任一即可重新評估）

- 需要做 production node：state crash recovery 變成硬需求
- 需要支援 state > RAM：service state 超過記憶體大小
- 需要多 fork concurrent validation：fork choice 性能變瓶頸

### 不啟動 Phase 3 的長期維護成本

- ChainState 200+ callsite 永遠是 singleton-bound，每多寫一行 callsite
  都是未來 Phase 3 要改的 case
- `KeyLevelCache` / `recentStateRoots` / `priorStates` 等多個 ChainState 欄位
  在 Phase 3 平行化時都需要 per-context，不只是 dirty set
- 對齊 GP 升級時，singleton 的隱式依賴會讓重構更難

上述成本已知且可接受，不構成提前啟動 Phase 3 的理由。Phase 3 啟動條件見上方。

### 預估規模

- **時間**：3.5-5 個月（Phase 3.1: 3-4 週；3.2: 1-2 週；3.3: 2-3 週；
  3.4: 2-3 週；3.5: 0.5 天）
- **影響**：跨 67+ 檔案，~200+ callsites；Phase 2 的 diff-based dirty key
  產生機制**完全 forward-compatible**（不依賴 singleton / context-scoped），
  不需回頭重寫

### 已知設計準則（為未來啟動預留）

- Phase 2 的 diff-based dirty key 產生機制不依賴 singleton / context-scoped，
  Phase 3 平行化時直接複用
- TrieStore 在 Phase 2 已設計為單純的 KV abstraction（無 singleton 假設），
  Phase 3 可直接多 ChainState 共用同一個 TrieStore instance
- Snapshot lifecycle 需配合 `Close()` / RAII 模式，避免 PebbleDB LSM compaction 被擋

---

## 附錄：Phase 1 設計事實（不能改）

> 這些是後續所有 phase 設計時必須遵守的不變量。違反就 state root 不一致。

1. **`PreimageLookup`（a_p）保持獨立** — 不進 globalKV。
   理由：blob 可能很大，且 key 是 raw hash，不適合塞進 31-byte StateKey
   space；查詢 pattern 是 `PreimageLookup[hash]`，不經過 StateKey 轉換。

2. **`StorageDict` / `LookupDict` 在 Phase 1 仍存在** — 但只供 jamtests
   `ServiceAccount.Encode` wire format 使用，runtime 完全不讀。後續所有 phase
   不應該依賴這兩個 map。**Phase 1 收尾完成後此條失效**。

3. **`ServiceInfo.Items` / `Bytes`** — 仍存在於 struct 作為 wire format mirror，
   每次序列化前從計數器同步。Phase 1 收尾或未來 codec-only struct 重構時
   可一併移除。

4. **`encodePreimageMetaValue`** — globalKV 中 preimage meta 的 value 是 JAM
   Marshal 後的 bytes，不是原始 `TimeSlotSet`。後續所有 phase 把 globalKV
   換 backing 時這個編碼不能變。

5. **delta1 序列化欄位順序**（GP §9.3）：
   `Version → CodeHash → Balance → MinItemGas → MinMemoGas → Bytes(a_o) → DepositOffset(a_f) → Items(a_i) → CreationSlot → LastAccumulationSlot → ParentService`
   — 順序錯一個 = state root 全錯。

6. **StateKey 構造**（GP D.1/D.2）：
   - storage: `prefix=0xFFFFFFFF || rawKey → Blake2b → interleave with serviceID`
   - preimage meta: `prefix=E4(length) || hash → Blake2b → interleave with serviceID`
   - preimage lookup: `prefix=0xFFFFFFFE || hash → Blake2b → interleave with serviceID`
     **(此 prefix 僅用於 wire-format state-key list / 不在 `globalKV` 內)**

7. **GP §9.8 `a_t` 公式**：`max(0, B_S + B_I*a_i + B_L*a_o − a_f)`
   — 已修復；後續所有 phase 不能 regress 回到 `return storage`。

8. **`unmatchedKeyVals` fallback pool 已徹底移除** — 後續所有 phase 不應引入
   任何類似的「state 半載入 + 補洞 pool」概念。state == globalKV，single
   source of truth。

---

## 測試驗證流程（沿用 Phase 1）

每個後續 phase 的子步驟完成後，跑以下三套確認無 regression：

1. **Minifuzz**（4 suites: storage, storage_light, fallback, safrole）— 102 個 test pair，無 error
2. **Picofuzz**（4 suites）— 無 MISMATCH / FAIL
3. **jam-conformance traces**（282 個 JSON）— 全部 PASSED

State root 必須跟 Phase 1 byte-for-byte 對齊。

測試流程詳見 `cursor-fuzz-testing-prompt.md`。
