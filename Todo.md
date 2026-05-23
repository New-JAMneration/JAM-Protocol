# ServiceAccount globalKV 重構 — Phase 1 完成摘要 + Phase 2/3 平行化計畫

## Phase 1 — globalKV 重構（已完成 ✅）

> 目標：把 `ServiceAccount` 的 `StorageDict`（a_s）和 `LookupDict`（a_l）合併為
> 統一的 `globalKV map[StateKey][]byte` + 增量計數器。`PreimageLookup`（a_p）保持獨立不動。

### 完成狀態

| Step | 狀態 | 摘要 |
|------|------|------|
| 1 | ✅ | `ServiceAccount` 加 `globalKV` + `totalNumberOfItems` / `totalNumberOfOctets` + `NewServiceAccount()` |
| 2a | ✅ | `NewStorageStateKey` / `NewPreimageMetaStateKey` / `NewPreimageLookupStateKey` |
| 2b | ✅ | `internal/utilities/safemath/` (Add/Sub/Mul + ErrOverflow) |
| 2c | ✅ | Get/Insert/Delete/Update + Clone + `cloneMapOfSlices` 通用 helper |
| 3 | ✅ | `CalcKeys`/`CalcOctets` 讀計數器 O(1)；`ThresholdBalance` method；`dev_assert` build tag |
| 4a | ✅ | `host_call_general.go`: read/write/info 切到 globalKV |
| 4b | ✅ | `host_call_accumulate.go`: new/transfer/eject/query/solicit/forget/provide；`finalizeNewAccount` 解決 new() 的 StateKey 問題 |
| 4c | ✅ | PVM 結構體 `StorageKeyVal` 欄位移除（與 7.5d 合流）|
| 5 | ✅ | `HistoricalLookup`/`FetchCodeByHash`/`ValidatePreimageLookupDict` 簽名加 `serviceID`；新增 `AddPreimage` method |
| 6 | ✅ | `StateEncoder` 直接走 globalKV |
| 7 | ✅ | `StateKeyValsToState` 三類桶 + `IsPreimage` 回傳 hash；`IsLookup` 移除 |
| 7.5 | ✅ | `unmatchedKeyVals` 整條 pipeline 移除（`StateKeyValsToState` → `(State, error)` + ChainState 6 個 method + state-root append 全清）|
| 8a | ✅ | `ServiceAccountDerivatives` 移除 |
| 8b | ✅ | `unmarshal_json.go` 走 globalKV（dual-write legacy maps 給 jamtests wire format）|
| 8c-f | ⏳ | **未做** — jamtests `ServiceAccount.Encode` wire format 編 raw storage key，無法從 hash 反推，需重新設計 JAM test-vector binary format |
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

Storage-heavy workloads 顯著加快。Phase 2/3 預期會把優勢放大並打開平行化大門。

---

## Phase 2 — globalKV 持久化（PebbleDB backing）

### 目標

把 `globalKV` 從 `map[StateKey][]byte`（純記憶體 Go map）換成 PebbleDB-backed
KV store。同時：

- 跨 process restart 持久化 service storage / preimage meta（crash recovery）
- 支援 state 超過 RAM 大小（large state mode）
- 開放 **concurrent reads**（多 goroutine 同時 Get，PebbleDB snapshot 隔離）
- 為 Phase 3 trie node 共享提供同一個 backing store

> **狀態：規劃中。Phase 1 已穩定（282/282 conformance pass），可隨時啟動。**

### 既有資源

我們**已經有 PebbleDB 後端**（`internal/database/provider/pebble/`），且 `internal/database/database.go` 的 `Database` interface 已經提供標準 KV store 抽象：

```go
type Database interface {
    Reader       // Has / Get
    Writer       // Put / Delete
    Batcher      // NewBatch
    Iterable     // NewIterator
    io.Closer
}

type Batch interface {
    Writer
    Commit() error
    Close() error
}
```

→ Phase 2 不需要新引入 dependency，也不需要重寫 KV store wrapper。

### 設計（單一 Database 多 prefix 路線）

採 prefix-based namespacing 模式，**所有 service 共用同一個 PebbleDB
instance**（不是每個 service 一個 DB），prefix 區分用途：

```
prefix 1: globalKV entry          → key = 0x01 || StateKey(31 bytes)
prefix 2: per-account counters    → key = 0x02 || serviceID(4 bytes LE)  → 12 bytes (items u32 + octets u64)
prefix 3: trie node               → key = 0x03 || nodeHash[1:32]         (Phase 3)
prefix 4: trie node value (>32B)  → key = 0x04 || valueHash[0:32]        (Phase 3)
prefix 5: trie refcount           → key = 0x05 || nodeHash[1:32]         (Phase 3)
prefix 6: legacy chain state      → 既有 chain_state.go 使用              (現狀不動)
```

Storage key 是 31 bytes（不是 32），跟 GP D.1 type-3 state key 對齊。

### 影響範圍

| 檔案 | 改動 |
|------|------|
| `internal/types/service_account_kv.go` | 把 `globalKV map[StateKey][]byte` 改成 backed by `db.KVStore` interface（新增 `kvBackend` 欄位）|
| 新增 `internal/types/kv_backend.go` | `KVBackend` interface（Phase 2 wrapper over `database.Database`）+ `mapKVBackend` (in-memory) + `pebbleKVBackend` |
| 新增 `internal/storage/service_store.go` | `ServiceKVStore` — 管理 prefix + per-service iteration + batch commit |
| `internal/blockchain/chain_state.go` | 共用 PebbleDB instance；`StateCommit` 用 batch commit |
| `cmd/fuzz/main.go` + `internal/fuzz/service.go` | 用 PebbleDB instance（或 in-memory backend）初始化 ServiceState |

業務邏輯（PVM host call、accumulation、HistoricalLookup、StateEncoder）**完全不動**——
Phase 1 已經把 `GetStorage` / `InsertStorage` / `GetPreimageMeta` 等方法封裝好了。

### 子步驟（依序）

**2.1: KVBackend 抽象 + in-memory adapter**
- [ ] 新增 `internal/types/kv_backend.go`，定義 `KVBackend` interface（Get/Put/Delete/NewBatch/Iterator）
- [ ] 提供 `NewMemoryKVBackend()`（用 `map[StateKey][]byte` 包一層，跟現在語意完全一致）
- [ ] `ServiceAccount.globalKV` 從 `map[StateKey][]byte` 改成 `KVBackend` interface
- [ ] 所有 method（Get/Insert/Delete/Update Storage/PreimageMeta、Clone、GetGlobalKVItems...）改成走 KVBackend
- [ ] **跑三套 fuzz，state root 必須完全一致**（in-memory backend 路徑）

**2.2: PebbleDB backend**
- [ ] 新增 `internal/types/kv_backend_pebble.go`，實作 `pebbleKVBackend` over `database.Database`
- [ ] 每個 service 用 prefix `0x01 || serviceID(4 LE) || StateKey(31)` = 36 bytes
- [ ] `Iterator` 用 prefix scan：列出某個 service 的所有 entry
- [ ] Batch 寫入：一個區塊的所有 storage 變動透過 PebbleDB `Batch.Commit()` 一次提交

**2.3: Counter 持久化**
- [ ] `totalNumberOfItems` / `totalNumberOfOctets` 隨 entry 一起持久化（prefix `0x02`）
- [ ] 反序列化路徑：開機時從 prefix `0x02` 讀回計數器，不必掃 globalKV 重算
- [ ] **原子性**：counter 更新必須跟 entry 寫入在同一個 batch 內提交

**2.4: Clone() 改用 snapshot**
- [ ] `ServiceAccount.Clone()` 不再 deep copy map，而是建立 PebbleDB snapshot
- [ ] snapshot 提供 immutable read view → **多 goroutine 並行 read 不會 race**
- [ ] 寫入時走 batch，commit 時 swap snapshot

**2.5: ChainState 整合**
- [ ] 共用同一個 PebbleDB instance（`ChainState` 已經有 `persistentRepo`）
- [ ] `StateCommit` 用 batch 把 block + state + serviceKV 一次提交
- [ ] 反序列化路徑（`StateKeyValsToState`）的「三類桶」改寫入 PebbleDB batch

**2.6: Memory backend fallback**
- [ ] 維持 in-memory backend 給 jamtests / unit test fixtures
- [ ] 跑 fuzz 時 default 用 PebbleDB；跑 unit test 時用 memory
- [ ] `NewServiceAccount()` 預設用 in-memory backend，避免測試需要 DB 初始化

**2.7: 驗證**
- [ ] **三套 fuzz suite 全綠**（state root byte-for-byte 對齊 Phase 1 結果）
- [ ] Benchmark：picofuzz p99 對比 Phase 1（預期：concurrent read 抹掉 fallback / safrole 的 wrapper overhead；storage workloads 因 batch commit 進一步加快）
- [ ] Crash recovery 測試：跑到一半 kill -9 docker container，重新啟動後 state root 仍正確
- [ ] Iterator 一致性測試：在 snapshot 上 iterate 不被同時的 writer 影響

### 預估規模

- **新增**：~250 行（`kv_backend.go` + `kv_backend_pebble.go` + counter persistence + snapshot wrapper）
- **修改**：~80 行（`service_account_kv.go` 內部從 map → KVBackend；ChainState batch 整合）
- **業務邏輯**：**0 行改動**（Phase 1 已封裝完成）
- **時間**：1.5～2 週（含 crash recovery test）

### 風險

1. **PebbleDB sync 模式選擇**：`pebble.Sync` 慢但安全；`pebble.NoSync` 快但 crash 可能丟資料。fuzz 環境用 NoSync + 區塊邊界 fsync 一次。
2. **Iterator 排序**：PebbleDB 預設按 key bytewise 排序，跟 trie merklization 用的 bit-level partition 順序不一致。merklization 仍需要 `sort.Slice`，這個保留不動。
3. **Memory backend 跟 PebbleDB backend 的 Clone 語意差異**：memory 是 deep copy（O(n)），PebbleDB 是 snapshot（O(1) 但需要 release）。要做封裝，caller 不需要 care 哪個 backend。

---

## Phase 3 — Trie Structural Sharing + 平行化

### 目標

把 merklization 從「每個區塊全量重建 trie」改為「**增量更新 + reference-counted node sharing**」。
獲得三個直接收益：

- **O(1) fork state**：建立 fork 只是 copy root hash + bump refcount，不再 deep copy
- **平行 block validation**：multiple forks 同時推進，trie node 自動共享
- **平行 accumulate**：不同 service 的 trie 子樹互不干擾，goroutine pool 並行

> **狀態：規劃中，依賴 Phase 2 PebbleDB backend 完成。**

### 既有資源

我們的 binary trie 編碼（位於 `internal/utilities/merklization/merklization.go`）已經是 GP-conformant 的 64-byte node 格式：

| 元件 | 編碼（GP §C.6 / Appendix D）|
|------|----------------------------|
| Branch node | `[64]byte`: `{left[0]&0x7F, left[1:32], right[0:32]}` |
| Embedded leaf (value ≤ 32) | `[64]byte`: `{0x80\|len, key[:31], value, pad}` |
| Regular leaf (value > 32) | `[64]byte`: `{0xC0, key[:31], blake2b(value)}` |
| Partition | `partitionByBit`（in-place）|

→ Phase 3 的 trie node 序列化直接用既有編碼，不需重新設計。Phase 3 要加的
是「**把 node 寫進 PebbleDB + refcount 管理 + 增量 merklize**」的層，core
merklize 演算法本身不動。

### 設計

**核心抽象**：`internal/storage/trie_store.go`，封裝在 Phase 2 同一個 PebbleDB instance 上：

```go
type TrieStore struct {
    db database.Database
}

// MerklizeAndCommit 計算 keyValues 的 trie root，把每個 node + leaf value 寫進 PebbleDB
// （prefix 3 / prefix 4），同時把每個新 node 的 refcount +1（prefix 5）。
// 返回 root hash。
func (t *TrieStore) MerklizeAndCommit(pairs [][2][]byte) (StateRoot, error)

// MerklizeOnly 不寫 DB，只算 root。給 verifier / dry-run 路徑用。
func MerklizeOnly(pairs []StateKeyVal) StateRoot

// GetNode / GetNodeValue / TrieExists：trie read 介面
func (t *TrieStore) GetNode(hash OpaqueHash) (Node, error)
func (t *TrieStore) GetNodeValue(node Node) ([]byte, error)
func (t *TrieStore) TrieExists(rootHash OpaqueHash) (bool, error)

// IncreaseNodeRefCount / DecreaseNodeRefCount：手動操作 refcount（少用）
func (t *TrieStore) IncreaseNodeRefCount(hash OpaqueHash) error
func (t *TrieStore) DecreaseNodeRefCount(hash OpaqueHash) (uint64, error)

// DeleteTrie：遞迴刪除某個 root 的所有 node，refcount 降為 0 才真刪
func (t *TrieStore) DeleteTrie(rootHash OpaqueHash) error
```

### 子步驟（依序）

**3.1: TrieStore + persistent node storage**
- [ ] 新增 `internal/storage/trie_store.go`，實作 `MerklizeAndCommit` / `GetNode` / `IncreaseNodeRefCount` / `DecreaseNodeRefCount` / `DeleteTrie` / `TrieExists`
- [ ] 用我們現有的 `internal/utilities/merklization` 的 `encodeBranchNode` / `encodeLeafNode`（編碼一致，不需重寫）
- [ ] **無需重新實作 trie merklize 演算法**——只是把現有 `merklize()` 加上 `storeNode` / `storeValue` callback
- [ ] Refcount key encoding：`0x05 || nodeHash[1:32]` → `uint64 LE`
- [ ] 第一個版本仍走全量 merklize（每區塊重算），但 node 進 DB 並 share

**3.2: 整合到 `StateCommit`**
- [ ] `ChainState.StateCommit` 改用 `trieStore.MerklizeAndCommit()`
- [ ] 從第二個區塊開始，沒改的 trie node 自動 share（refcount += 1）
- [ ] **三套 fuzz 全綠**（state root 完全一致）
- [ ] 觀察 DB 大小成長曲線：應該是 sub-linear（共享 node 越來越多）

**3.3: 增量 merklize（incremental update）**
- [ ] 引入「dirty key set」追蹤：每個區塊累積本區塊被改過的 StateKey
- [ ] 只重算 dirty path 上的 node，未變的 subtree 直接 reuse 舊 root hash
- [ ] 預期效果：blocks with small delta 的 merklize 時間從 O(n) → O(log n × delta)
- [ ] 這是 Phase 3 最大的性能收益，預期 picofuzz p99 大幅下降

**3.4: Snapshot-based fork**
- [ ] `ChainState` 引入 fork-aware state machine：每個 block 對應一個 root hash + PebbleDB snapshot
- [ ] Fork branch：root hash 加一個 refcount，不 deep copy state
- [ ] Fork prune：discard fork 時 `DeleteTrie(oldRoot)`，refcount 降為 0 的 node 才真刪

**3.5: 平行 block validation**
- [ ] STF (`stf.RunSTF`) 可在多個 fork branch 上 concurrent 跑
- [ ] 每個 goroutine 拿自己的 PebbleDB snapshot + dirty key buffer
- [ ] commit 時用 batch 寫回，PebbleDB serializes writes 自動 isolation
- [ ] 需要解決：`blockchain.GetInstance()` singleton → 改成 context-scoped instance

**3.6: 平行 accumulate**
- [ ] `ParallelizedAccumulation` 目前已經用 goroutine 跑 per-service `Psi_A`，但 `PartialStateSet.DeepCopy()` 是 bottleneck
- [ ] 改用 trie snapshot：每個 service 自己的 trie subtree（StateKey 前 4 bytes interleaved serviceID）
- [ ] 不同 service 的 commit 互不干擾
- [ ] Merge 階段：把 N 個 service-level 修改 merge 進 parent trie root

**3.7: 驗證**
- [ ] 三套 fuzz：state root byte-for-byte 對齊 Phase 1
- [ ] Benchmark：picofuzz p99 對比 Phase 2，預期增量 merklize 帶來大幅加速
- [ ] Stress test：跑很長的 trace（5000+ blocks）確認 DB 不會無止盡膨脹
- [ ] Refcount 不變式測試：every node 的 refcount = (number of tries referencing it)
- [ ] Fork 測試：建 100 個 fork → prune 99 個 → 確認只剩 1 個 fork 的 node 還在 DB

### 預估規模

- **新增**：~500 行（`trie_store.go` + incremental merklize logic + snapshot fork management）
- **修改**：~200 行（`ChainState` 改用 trie store；`ParallelizedAccumulation` 改 service-level snapshot）
- **業務邏輯**：**0 行改動**
- **時間**：3～4 週（增量 merklize + 平行化測試是最大時間消耗）

### 平行化收穫

| 場景 | Phase 1 | Phase 2 | Phase 3 |
|------|---------|---------|---------|
| 不同 service 平行 accumulate | ❌（Go map race）| ⚠️（PebbleDB single-writer）| ✅（trie subtree 隔離 + batch commit）|
| 平行 block validation | ❌（state singleton）| ⚠️（DeepCopy 太慢）| ✅（O(1) snapshot fork）|
| Read-side concurrency | ❌ | ✅（PebbleDB snapshot）| ✅ |
| 增量 trie 更新 | ❌（每次全量）| ❌ | ✅（dirty path only）|
| Fork prune cost | O(n) deep copy | O(n) DB scan | O(改過的 node 數）|

### 風險

1. **Incremental merklize 邊界**：必須正確處理 leaf delete → branch collapse 的情況，否則 trie 結構錯誤
2. **Refcount overflow**：`uint64` 上限 1.8e19，正常 fork tree 不會踩到，但邏輯 bug 可能造成 refcount 漏減 → 浪費 DB 空間
3. **Snapshot 生命週期**：必須記得 release，否則 PebbleDB 的 LSM tree 無法 compact 老資料
4. **平行 accumulate 的 conflict resolution**：兩個 service 同時 read 對方的 storage 時，目前的 partial state 模型可能不適合直接 parallelize；可能需要 transaction-style retry

---

## 附錄：Phase 1 設計事實（不能改）

> 這些是 Phase 2/3 設計時必須遵守的不變量。違反就 state root 不一致。

1. **`PreimageLookup`（a_p）保持獨立** — 不進 globalKV。
   理由：blob 可能很大，且 key 是 raw hash，不適合塞進 31-byte StateKey
   space；查詢 pattern 是 `PreimageLookup[hash]`，不經過 StateKey 轉換。

2. **`StorageDict` / `LookupDict` 在 Phase 1 仍存在** — 但只供 jamtests
   `ServiceAccount.Encode` wire format 使用，runtime 完全不讀。Phase 2/3 不應該
   依賴這兩個 map。

3. **`ServiceInfo.Items` / `Bytes`** — 仍存在於 struct 作為 wire format mirror，
   每次序列化前從計數器同步。Phase 3 引入 codec-only struct 時可一併移除。

4. **`encodePreimageMetaValue`** — globalKV 中 preimage meta 的 value 是 JAM
   Marshal 後的 bytes，不是原始 `TimeSlotSet`。Phase 2/3 把 globalKV 換 backing
   時這個編碼不能變。

5. **delta1 序列化欄位順序**（GP §9.3）：
   `Version → CodeHash → Balance → MinItemGas → MinMemoGas → Bytes(a_o) → DepositOffset(a_f) → Items(a_i) → CreationSlot → LastAccumulationSlot → ParentService`
   — 順序錯一個 = state root 全錯。

6. **StateKey 構造**（GP D.1/D.2）：
   - storage: `prefix=0xFFFFFFFF || rawKey → Blake2b → interleave with serviceID`
   - preimage meta: `prefix=E4(length) || hash → Blake2b → interleave with serviceID`
   - preimage lookup: `prefix=0xFFFFFFFE || hash → Blake2b → interleave with serviceID`

7. **GP §9.8 `a_t` 公式**：`max(0, B_S + B_I*a_i + B_L*a_o − a_f)`
   — 已修復；Phase 2/3 不能 regress 回到 `return storage`。

8. **`unmatchedKeyVals` fallback pool 已徹底移除** — Phase 2/3 不應引入任何
   類似的「state 半載入 + 補洞 pool」概念。state == globalKV，single source of
   truth。

---

## 測試驗證流程（沿用 Phase 1）

每個 Phase 2/3 子步驟完成後，跑以下三套確認無 regression：

1. **Minifuzz**（4 suites: storage, storage_light, fallback, safrole）— 102 個 test pair，無 error
2. **Picofuzz**（4 suites）— 無 MISMATCH / FAIL
3. **jam-conformance traces**（282 個 JSON）— 全部 PASSED

State root 必須跟 Phase 1 byte-for-byte 對齊。

測試流程詳見 `cursor-fuzz-testing-prompt.md`。
