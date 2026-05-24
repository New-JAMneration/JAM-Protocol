# 實作進度摘要（供新對話 context）

## 已完成

### Phase 1 收尾（D1）— committed `52a15351`
- 拔除 `ServiceAccount.StorageDict` / `LookupDict` struct 欄位
- 拔除 dual-write（unmarshal_json / Clone）、MigrateLegacyMapsToGlobalKV、cloneLookupDict
- Export `BuildStorageStateKey` / `BuildPreimageMetaStateKey`
- 改寫 jamtests Validate() + ParseAccountToServiceAccountState 走 globalKV
- 所有 test fixture 改用 InsertStorage / InsertPreimageMeta
- 關鍵發現：`types.ServiceAccount.Encode/Decode` 是 transitively dead code，不需要 wire format 重設計或 upstream 確認
- **600/600 conformance traces PASSED**

### Phase 2.1 callback 改造 — committed `56090903`
- 新增 `TrieNode [64]byte` type + `IsLeaf/IsBranch/GetBranchHashes/GetLeafKey/GetLeafValue/GetLeafValueHash` 方法
- 新增 `StoreNodeFunc` / `StoreValueFunc` callback types
- `merklize()` / `merklizeWithCache()` 加 storeNode/storeValue callback + 回傳 error
- 所有 caller 已適配（Phase 2.1 階段 storeNode/storeValue 全傳 nil）
- **600/600 conformance traces PASSED**

### Phase 2.2 StateCommit 整合 — committed `de85dcbe`
- ChainState 加 `trieStore *store.Trie` field + GetInstance/ResetInstance 初始化
- `PersistStateForBlock` 走 `trieStore.MerklizeAndCommit`
- Ring buffer evict 時 `trieStore.DeleteTrie` + `persistentRepo.DeleteStateData`
- `ResetInstance` 呼叫 `trieStore.DeleteAll()` 清跨 session trie data
- 消除三重計算：`lastCommittedStateRoot` field + `StateCommit() (StateRoot, error)` + ImportBlock/SetState 簡化（3x→1x serialize+merklize）
- `restoreWithState` 重置 `recentStateRoots` + `lastCommittedStateRoot`
- `SeedGenesisToBackend` 設定 `lastCommittedStateRoot`
- **600/600 conformance traces PASSED**

### Phase 2.1-2.3 TrieStore + incremental merklize — committed `8b9c6292`
- **TrieStore** (`internal/store/trie_store.go`): MerklizeAndCommit / MerklizeOnly / GetNode / GetNodeValue / TrieExists / IncreaseNodeRefCount / DecreaseNodeRefCount / GetNodeRefCount / DeleteTrie / DeleteAll
- **DiffSortedKeyVals** (`internal/store/diff.go`): O(n) merge-scan 產生 inserted/deleted/modified entries
- **IncrementalMerklize** (`internal/store/incremental_merklize.go`): 從 prior root 出發，按 bit[depth] 遞迴分組 dirty entries，reuse unchanged subtrees
- `persistStateForBlockMerklize`: 增量為主路徑，fallback 全量
- `incrementalMerklizeAndCommit`: batch.Put + refcount 持久化
- `KeyLevelCache.Invalidate(key)` on delete 防止 stale cache hits
- Bug fix: `buildSubtree` 必須在 len>1 時總是建 branch（即使一側為空）— 對齊全量 merklize 行為
- Bug fix: `collapseBranch` → `tryCollapse` 僅 leaf child 才 collapse；re-hash 修復 MSB mask
- Unit tests: 24 個 TrieStore tests + 7 個 diff tests + 7 個 incremental tests + 2 個 StateEncoder invariant tests
- **600/600 conformance traces PASSED**
- **Minifuzz 4 suites: 0 errors**
- **Picofuzz 4 suites: 0 MISMATCH/FAIL**

## Benchmark 結果

### PoC-1 耗時分佈（PersistStateForBlock 內部）
- encode: ~20-22%
- sort: <0.1%
- **merklize: ~58%**（加權平均 4 suites，穩態）
- save: ~7-8%
- 結論：**GO**（遠超 40% 門檻）

### Minifuzz 效能對比（main vs incremental）
| Suite | main | incr | Δ |
|-------|------|------|---|
| **storage** | 6019ms | 1027ms | **−82.9%** |
| storage_light | 2683ms | 2750ms | +2.5% |
| fallback | 1325ms | 1384ms | +4.5% |
| safrole | 1570ms | 1549ms | −1.3% |

### Picofuzz 效能對比
| Suite | main | incr | Δ |
|-------|------|------|---|
| storage | 863ms | 820ms | −5.0% |
| storage_light | 808ms | 869ms | +7.5% |
| fallback | 871ms | 763ms | −12.4% |
| safrole | 784ms | 761ms | −2.9% |

### Conformance Traces 效能對比（per-mode）
| Mode | main | incr | Δ |
|------|------|------|---|
| storage | 5794ms | 5906ms | +1.9% |
| storage_light | 4173ms | 4282ms | +2.6% |
| fallback | 3329ms | 3406ms | +2.3% |
| safrole | 3514ms | 3554ms | +1.1% |
| preimages | 4108ms | 4212ms | +2.5% |
| preimages_light | 4088ms | 4180ms | +2.3% |

### 結論
- storage-heavy workload（Minifuzz storage）: **−83% 加速**
- 其他 workload: ±5% 範圍（dirty entries 少時增量路徑與全量等效）
- conformance traces (tiny mode, dirty=0 居多): +2% 微小開銷（diff + trie lookup）

## 已知問題
- `StateEncoder` 非確定性：globalKV map iteration order 影響 delta1 value encoding。不影響 production（每 block 只呼叫一次 + sort），但 minimal test fixture 會觸發。
- Picofuzz stf-data/storage 目錄為空（需要 repopulate）

## 測試方式
- WSL: `export PATH=$HOME/go-sdk/go/bin:$HOME/go/bin:$PATH && CGO_ENABLED=1 go build ./...`
- Docker build: `docker build --build-arg GP_VERSION=0.7.2 --build-arg TARGET_VERSION=0.3.1 -t new-jamneration-target:TAG -f docker/Dockerfile .`
- Docker trace test: `docker run --rm -v "${PWD}/pkg/test_data:/app/pkg/test_data" jam-gobuilder:TAG sh -c "CGO_ENABLED=1 go build -o /tmp/jam-node ./cmd/node 2>/dev/null && /tmp/jam-node test --type trace --mode MODE"`
- Minifuzz/Picofuzz: 見 `cursor-fuzz-testing-prompt.md`
