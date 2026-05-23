# ServiceAccount 統一 globalKV 重構計畫

> 目標：將 `ServiceAccount` 的 `StorageDict`（a_s）和 `LookupDict`（a_l）合併為
> 統一的 `globalKV map[StateKey][]byte` + 增量計數器。
> **`PreimageLookup`（a_p）保持獨立不動。**

## 實作狀態（2026-05-24 update）

| Step | 狀態 | 備註 |
|------|------|------|
| 1: ServiceAccount globalKV + counters + NewServiceAccount | ✅ | dual-map 過渡保留 |
| 2a: StateKey helpers | ✅ | 新增 `NewStorageStateKey` / `NewPreimageMetaStateKey` / `NewPreimageLookupStateKey` |
| 2b: safemath utilities | ✅ | `internal/utilities/safemath/` |
| 2c: globalKV 存取方法 + Clone + cloneMapOfSlices | ✅ | 16 個方法 + 通用 helper |
| 3: CalcKeys/Octets 讀計數器 + ThresholdBalance method + dev_assert | ✅ | bug fix 已套用 |
| 4a: host_call_general.go (read/write/info) | ✅ | |
| 4b: host_call_accumulate.go (new/transfer/eject/query/solicit/forget/provide) | ✅ | new() 用 `finalizeNewAccount` 解決新 service ID 問題 |
| 4c: PVM struct StorageKeyVal 欄位清理 | ✅ | 與 Step 7.5d 一起完成 |
| 5: accumulation + HistoricalLookup/FetchCodeByHash/ValidatePreimageLookupDict + AddPreimage | ✅ | 所有 signature 加 serviceID |
| 6: StateEncoder 簡化 | ✅ | 過渡期 sync 模式 |
| 7: StateKeyValsToState 三類桶 + IsPreimage(returns hash) + IsLookup 移除 | ✅ | |
| 7.5: unmatchedKeyVals pipeline 全清除 | ✅ | (State, error) 簽名 + ChainState 6 個欄位/方法 + state-root append 全清 |
| 8a: ServiceAccountDerivatives 移除 | ✅ | |
| 8b: unmarshal_json.go 走 globalKV | ✅ | 雙寫保留 legacy 給 jamtests wire format |
| 8c-f: 真正移除 StorageDict/LookupDict + ServiceInfo.Items/Bytes + codec-only struct | ⏳ 未做 | jamtests `ServiceAccount.Encode` wire format 編 raw storage key，無法從 hash 反推 |
| 9: 測試 fixtures 改用 globalKV | ✅ 主要部分 | service_account / accumulation / ce134 / jamtests preimages+accumulate 已改；新增 `MigrateLegacyMapsToGlobalKV` helper |
| **Pre-existing bug fix**: ThresholdBalance GP §9.8 | ✅ | `storage − a_f` correctly subtracted |

**11 commits on `feat/global-kv-refactor`**：

```
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

**測試結果（驗證跟 baseline byte-for-byte 對齊）**：

| 測試集 | 結果 |
|--------|------|
| Minifuzz × 4 suites (storage/storage_light/fallback/safrole) | 102/102 pairs each, all PASS |
| Picofuzz × 4 suites | PASS, no MISMATCH/FAIL |
| jam-conformance 0.7.2 traces | **282 / 282 PASSED** |

**速度對比（picofuzz p99）**：

| Suite | Baseline | Refactor | Δ |
|-------|----------|----------|---|
| storage | 220.5 ms | 208.8 ms | **-5.3%** |
| storage_light | 81.0 ms | 68.5 ms | **-15.4%** |
| fallback | 36.4 ms | 39.1 ms | +7.5% |
| safrole | 41.3 ms | 46.7 ms | +13.0% |

→ Storage-heavy workloads 顯著加快（Phase 1 目標達成）；非 storage workloads 多了一點 wrapper overhead。Phase 2/3 預期會把優勢放大。

---

## 重要：什麼合併、什麼不合併

| 原欄位 | 對應符號 | 是否合併進 globalKV | 原因 |
|--------|---------|-------------------|------|
| `StorageDict` (a_s) | `map[string]ByteSequence` | **是** → globalKV（delta2, prefix `0xFFFFFFFF`） | key/value 適合 StateKey 索引 |
| `LookupDict` (a_l) | `map[LookupMetaMapkey]TimeSlotSet` | **是** → globalKV（delta4, prefix `E4(length)`） | timeslot metadata 適合 StateKey 索引 |
| `PreimageLookup` (a_p) | `map[OpaqueHash]ByteSequence` | **否，保持獨立** | blob 資料可能很大，需直接 hash 查詢 |

> 設計原則：`globalKV` 統一儲存 service storage (s) 和 preimage meta (l)，以 `StateKey` 為鍵。

## ServiceInfo.Items/Bytes 與增量計數器的關係

目前 `ServiceInfo.Items`（a_i）和 `ServiceInfo.Bytes`（a_o）會在 delta1 序列化中被寫入 state。
新增的 `totalNumberOfItems` / `totalNumberOfOctets` 增量計數器是它們的 runtime 替代品。

**同步規則：**
- **寫入時**：所有 Insert/Delete 操作只更新增量計數器
- **序列化前**：從計數器同步到 `ServiceInfo.Items/Bytes`（確保 delta1 編碼正確）
- **反序列化後**：從 `ServiceInfo.Items/Bytes` 初始化計數器（O(1)，不需遍歷 globalKV）

過渡期（Step 1~7）兩者共存；Step 8 評估是否可移除 ServiceInfo 中的 Items/Bytes 欄位。

## 預期效益

- State 序列化時 storage(a_s) + lookup(a_l) 的 key 本身就是 trie key，省去轉換
- `a_i`/`a_o` footprint 追蹤從 O(n) 降為 O(1)
- 為未來持久化 KV store + trie 結構共享鋪路
  - key 固定 31 bytes、value 是 raw bytes，與 KV store 介面天然對齊（可直接接 PebbleDB / BadgerDB）

## 測試驗證

每完成一個 Step，都需要跑以下三套測試確保無 regression：

1. **Minifuzz**（4 suites: storage, storage_light, fallback, safrole）— 102 個 test pair，無 error
2. **Picofuzz**（4 suites）— 無 MISMATCH / FAIL
3. **jam-conformance traces**（282 個 JSON）— 全部 PASSED

測試流程詳見 `cursor-fuzz-testing-prompt.md`

---

## Step 1：雙寫過渡 — 新增 globalKV 欄位（保留全部舊 map）  ✅ DONE

- [ ] 在 `ServiceAccount` struct 新增 `globalKV map[StateKey][]byte`（私有欄位）
- [ ] 新增 `totalNumberOfItems uint32` 和 `totalNumberOfOctets uint64` 增量計數器（私有欄位）
- [ ] **`PreimageLookup` 保持不動**，`StorageDict` 和 `LookupDict` 暫時也保留（雙寫過渡）
- [ ] 確保 JSON/JAM 序列化不受影響（新欄位不參與序列化）
- [ ] 新增 `NewServiceAccount()` 建構函數，初始化 `PreimageLookup` 和 `globalKV` 為空 map
- [ ] 跑三套測試確認無 regression

**影響檔案**：`internal/types/state.go`

---

## Step 2：新增存取方法 + StateKey 輔助函數  ✅ DONE

### 2a: StateKey 構造輔助函數

在 `internal/utilities/merklization/state_key_constructor.go` 或新檔案中新增：
- [ ] `NewStorageStateKey(serviceID ServiceID, rawKey ByteSequence) (StateKey, error)` — 構造 delta2 用的 StateKey（prefix `0xFFFFFFFF` + rawKey → Blake2b → interleave with serviceID）
- [ ] `NewPreimageMetaStateKey(serviceID ServiceID, hash OpaqueHash, length U32) (StateKey, error)` — 構造 delta4 用的 StateKey（prefix `E4(length)` + hash → Blake2b → interleave with serviceID）
- [ ] **回傳 error 的理由**：純運算理論上不會出錯，但簽名保留 error 是為了 propagate 內部 JAM encoding 工具（如 `jam.PutUint32`）可能的 error，並符合 codebase 中其他 KV 介面的慣例

> 參考 GP Appendix D.1 State-Key-Construction 的第三種 arity

### 2b: 前置作業 — safemath 工具

- [ ] 新增 safemath 工具函數（`SafeAdd[T]`、`SafeSub[T]`、`SafeMul[T]`），回傳結果 + overflow bool
- [ ] 目前 codebase 中僅有 `PVM/argument_invocation.go` 的 `checkOverflow` 做乘法溢位檢查（PVM 私有，非通用工具）
- [ ] 需要通用的加法/減法/乘法溢位保護：
  - **Add/Sub**：供 Insert/Delete 方法的計數器更新使用
  - **Mul**：供 Step 3 的 `ThresholdBalance` 計算 `B_I*a_i` 和 `B_L*a_o` 使用
- [ ] 新增 `ErrOverflow` 常數，供 Insert/Delete/ThresholdBalance 在溢位時回傳
- [ ] 建議放在 `internal/utilities/safemath/` 或直接在 `internal/types/` 內實作

> 用於 Insert/Delete 方法 + ThresholdBalance 中的整數運算保護

### 2c: globalKV 存取方法

**定義位置**：方法必須在 `internal/types` package 內（與 `ServiceAccount` struct 同 package），
因為 `globalKV` 是私有欄位。建議放在新檔案 `internal/types/service_account_kv.go`。

針對 **Storage（a_s）** 操作：
- [ ] `(sa *ServiceAccount) GetStorage(key StateKey) ([]byte, bool)` — 從 globalKV 讀取
- [ ] `(sa *ServiceAccount) InsertStorage(key StateKey, originalKeySize uint64, value []byte) error`
  - 新增 key：`items += 1`, `octets += 34 + originalKeySize + len(value)`
  - 已存在 key（更新）：**先 `Sub(octets, len(prevValue))` 再 `Add(octets, len(newValue))`**，items 不變、keyLen 不重算
    - 實作注意：不可直接寫 `octets += len(newValue) - len(prevValue)`，uint64 下若新值較小會 underflow
  - **不接受「空值即刪除」語意**：caller 必須顯式呼叫 `DeleteStorage`（PVM `write` 的 `vz==0` 分支由 PVM 層分派）
  - 使用 safemath 進行溢位保護
  - **lazy init 防護**：`if sa.globalKV == nil { sa.globalKV = make(...) }`（反序列化路徑可能不走 NewServiceAccount）
  - **原子性實作要求**：先在區域變數副本上做 safemath 運算，所有步驟成功後才一次性賦值回 `sa.totalNumberOfItems` / `sa.totalNumberOfOctets`。任何一步 overflow 都直接 return ErrOverflow，計數器與 globalKV 保持原狀
- [ ] `(sa *ServiceAccount) DeleteStorage(key StateKey, keyLen, valueLen uint64) error`
  - key 存在時：`items -= 1`, `octets -= (34 + keyLen + valueLen)`
  - **冪等操作**：key 不存在時回傳 nil error，不更新計數器
  - 使用 safemath 進行溢位保護
  - **設計約束**：StateKey 是 Blake2b hash，無法反推 originalKey 長度，所以 caller 必須提供 `keyLen` 和 `valueLen`

針對 **PreimageMeta（a_l，即 LookupDict）** 操作：
- [ ] `(sa *ServiceAccount) GetPreimageMeta(key StateKey) (TimeSlotSet, bool)` — 從 globalKV 讀取並 **JAM Unmarshal** 還原 `TimeSlotSet`
  - Unmarshal 失敗時回傳 `nil, false`（與 key 不存在行為一致，不向 caller 暴露 error）
- [ ] `(sa *ServiceAccount) InsertPreimageMeta(key StateKey, length uint64, timeslots TimeSlotSet) error`
  - 內部先 **JAM Marshal** `timeslots` 再寫入 globalKV（value 在 globalKV 中已是編碼後的 bytes）
  - **新增 key**：`items += 2`, `octets += (81 + length)`
  - **已存在 key**：只覆蓋資料，**不更新計數器**
  - 使用 safemath 進行溢位保護
  - **lazy init 防護**：同 InsertStorage
- [ ] `(sa *ServiceAccount) UpdatePreimageMeta(key StateKey, newValue TimeSlotSet) error`
  - 更新已存在的 meta 值（內部 JAM Marshal），**不改計數器**
  - 若 `globalKV == nil` → 回傳 error（邏輯錯誤，不做 lazy init；與 Insert 的 lazy init 語意不同）
  - 若 key 不存在 → 回傳 error。**存在性判斷用 `_, exists := sa.globalKV[key]` 直接 map 查詢**（不要用 `GetPreimageMeta`，避免 Unmarshal 失敗被誤判為「key 不存在」）
- [ ] `(sa *ServiceAccount) DeletePreimageMeta(key StateKey, length uint64) error`
  - key 存在時：`items -= 2`, `octets -= (81 + length)`
  - **冪等操作**：key 不存在時回傳 nil error，不更新計數器
  - 使用 safemath 進行溢位保護

> **介面契約**：`globalKV` 中的 preimage meta value 已是 JAM 編碼後的 bytes。
> 序列化時直接 dump 即可，反序列化時直接灌入 globalKV，不需額外轉換。

Clone / DeepCopy：
- [ ] `(sa *ServiceAccount) Clone() ServiceAccount` — 實作模式：`cloned := *sa`（隱式複製所有 scalar fields 含 private 計數器）+ 對兩個 map 做 deep clone：
  ```go
  func (sa *ServiceAccount) Clone() ServiceAccount {
      cloned := *sa
      cloned.globalKV = cloneMapOfSlices(sa.globalKV)
      cloned.PreimageLookup = cloneMapOfSlices(sa.PreimageLookup)
      return cloned
  }
  ```
- [ ] `(ss ServiceAccountState) Clone() ServiceAccountState` — 遍歷呼叫每個 account 的 `Clone()`
- [ ] 新增 `cloneMapOfSlices[K, V]` 泛型 helper（`internal/utilities/` 或 `internal/types/`）

通用：
- [ ] `(sa *ServiceAccount) GetGlobalKVItems() map[StateKey][]byte` — 取得完整 globalKV map（序列化用）
- [ ] `(sa *ServiceAccount) SetGlobalKVItems(globalKV map[StateKey][]byte)` — 設定完整 globalKV map（反序列化用）
- [ ] `(sa *ServiceAccount) GetTotalNumberOfItems() uint32` / `SetTotalNumberOfItems(n uint32)`
- [ ] `(sa *ServiceAccount) GetTotalNumberOfOctets() uint64` / `SetTotalNumberOfOctets(n uint64)`

**注意**：`PreimageLookup`（a_p）的讀寫保持現有方式（直接 `account.PreimageLookup[hash]`）。
- [ ] 跑三套測試確認無 regression

**影響檔案**：`internal/types/service_account_kv.go`（新增）、`internal/utilities/merklization/state_key_constructor.go`

---

## Step 3：CalcKeys / CalcOctets 改為讀增量計數器  ✅ DONE (含 GP §9.8 bug fix)

- [ ] `CalcKeys(account)` → 回傳 `account.GetTotalNumberOfItems()`（O(1)）
- [ ] `CalcOctets(account)` → 回傳 `account.GetTotalNumberOfOctets()`（O(1)）
- [ ] **保留** `CalcStorageItemfootprint` 和 `CalcLookupItemfootprint`（純數學計算，不依賴任何 map，用於 write/solicit 的 pre-check 流程）
- [ ] `CalcThresholdBalance` 改為 method `(sa *ServiceAccount) ThresholdBalance() (uint64, error)`
  - 內部直接讀 `totalNumberOfItems`、`totalNumberOfOctets`、`DepositOffset`
  - 使用 safemath 做溢位保護
  - **注意**：現有 `CalcThresholdBalance` 有預存 bug（`storage >= aF` 時回傳 `storage` 而非 `storage - aF`，GP §9.8: `a_t = max(0, B_S+B_I*a_i+B_L*a_o - a_f)`）。**重構完成並跑完三套測試確認無 regression 後，再獨立修正此 bug**，避免重構期間引入額外變數
    - 修復後正確形式：`if sum < aF { return 0, nil } return sum - aF, nil`（搭配 safemath）
- [ ] 確認 `GetServiceAccountDerivatives` 正常運作（Step 8 將移除）
- [ ] **過渡期強制 assertion**（用 build tag `//go:build dev_assert` 控制）：
  - 每次 PVM host call / accumulation 結束時，呼叫 `assertConsistency(account)`
  - 檢查：`totalNumberOfItems == CalcKeys(舊算法)` 且 `totalNumberOfOctets == CalcOctets(舊算法)`
  - 檢查：`StorageDict` / `LookupDict` 每個 entry 都能在 `globalKV` 中找到對應 StateKey
  - 不一致時 `panic("globalKV consistency violation")`，附 diff log
  - Step 8 移除舊 map 時，assertion 一併移除
- [ ] 更新 `service_account_test.go` 相關測試
- [ ] 跑三套測試確認無 regression

**影響檔案**：`internal/service_account/service_account.go`、`service_account_test.go`

---

## Step 4：PVM Host Call 改用新方法（最大改動）  ✅ DONE

### 4a: host_call_general.go

- [ ] `read` (=3)：`a.StorageDict[string(key)]` → 用 `NewStorageStateKey()` 構造 key，再呼叫 `GetStorage(stateKey)`
- [ ] `read`：**StorageKeyVal fallback pattern 處理**（見下方說明）
- [ ] `write` (=4)：改用新方法，流程如下：
  1. 用 `GetStorage(stateKey)` 取得舊值（取代 `a.StorageDict[string(key)]`）
  2. **保留 `CalcStorageItemfootprint` 做 pre-check**：用舊值/新值的 footprint 差異計算新的 items/octets
  3. 用 `CalcThresholdBalance(newItems, newOctets, DepositOffset)` 檢查餘額是否足夠
  4. 若餘額不足 → 回傳 FULL，**不呼叫 InsertStorage**
  5. 若餘額足夠 → 呼叫 `InsertStorage` 或 `DeleteStorage` 執行寫入（方法自動更新計數器）
  6. 移除手動 `ServiceInfo.Items/Bytes` 更新（由方法自動處理）
- [ ] `lookup` (=2)：**保持現狀** — 仍從 `a.PreimageLookup[hash]` 讀取（PreimageLookup 不合併）
- [ ] `info` (=5)：`ServiceInfo.Items/Bytes` 改為讀計數器 `GetTotalNumberOfItems()`/`GetTotalNumberOfOctets()`
- [ ] 跑三套測試確認無 regression

### 4b: host_call_accumulate.go

- [ ] **掃描全部 `ServiceAccount{...}` literal**，改用 `NewServiceAccount()`（否則 `globalKV` 為 nil 會 panic）
- [ ] `new` (=18)：ServiceAccount 初始化改用 `NewServiceAccount()` + 新方法
- [ ] `transfer` (=20)：minBalance 計算改用計數器
- [ ] `solicit` (=23)：用 `NewPreimageMetaStateKey()` 構造 key，`account.LookupDict[...]` → `GetPreimageMeta`/`InsertPreimageMeta`
- [ ] `forget` (=24)：`account.LookupDict[...]` 刪除 → `DeletePreimageMeta`
- [ ] `provide` (=26)：preimage blob 仍寫入 `PreimageLookup`，meta 部分改用 `InsertPreimageMeta`
- [ ] 跑三套測試確認無 regression

### StorageKeyVal fallback pattern 處理

目前 `read`/`write` host call 有 fallback 機制：當 `StorageDict` 沒有某 key 時，
會從 `StorageKeyVal`（序列化的 state key-vals pool）搜尋（`getStorageFromKeyVal()`）。
找到後會 cache 回 `StorageDict`。

重構後採用 **方案 A**：初始化 state 時將所有 storage/lookup 的 StateKeyVals **全量載入 globalKV**。
`globalKV` 的核心價值是 **state == globalKV**，不該有「半載入」狀態。

待清理的 fallback 函數：
- [ ] `getStorageFromKeyVal()` → 移除
- [ ] `removeStorageFromKeyVal()` → 移除
- [ ] `getLookupItemFromKeyVal()` → 移除
- [ ] `lookupInKeyVal()` / `lookupAndRemoveKeyVal()`（`accumulation`）→ 移除

### 4c: PVM 結構體清理（方案 A 下游影響）

- [ ] `HostCallArgs.GeneralArgs.StorageKeyVal` 欄位移除（`host_call_general.go:60`）
- [ ] `ResultContext.StorageKeyVal` 欄位移除（`accumulate_invocation.go:231, 242`）
- [ ] `accumulate_invocation.go` 的 5 處 `*StorageKeyVal` 傳遞清理
- [ ] `I(...)` 函式簽名清理（`accumulate_invocation.go:108, 109`）
- [ ] `newStorageKeyVal := storageKeyVal.DeepCopy()` 等 DeepCopy 模式清理

**影響檔案**：`PVM/host_call_general.go`、`PVM/host_call_accumulate.go`、`PVM/accumulate_invocation.go`

---

## Step 5：accumulation 模組 + HistoricalLookup 改用新方法  ✅ DONE

- [ ] `HistoricalLookup()` — **本次重構保持 free function 形式**（不改為 method，改 method 影響 17+ 處呼叫端，是純風格 refactor，留作未來獨立 ticket）。簽名新增 `serviceID types.ServiceID` 參數；`account.LookupDict[lookupkey]` 改用 `GetPreimageMeta(NewPreimageMetaStateKey(serviceID, hash, length))`；`account.PreimageLookup[hash]` 保持不動
  - 業務呼叫端（3 處）：`host_call_refine.go:43`、`refine_invocation.go:54`、`work_package.go:19`
  - 測試呼叫端（14 處）：`service_account_test.go` 所有 `HistoricalLookup(...)` 呼叫
- [ ] `ValidatePreimageLookupDict()` / `existsInLookupDict()` — 簽名新增 `serviceID` 參數；`account.LookupDict[key]` 改用 `GetPreimageMeta`
  - `FetchCodeByHash()` 因呼叫 `ValidatePreimageLookupDict` 也需傳遞 serviceID
  - 影響：`PVM/accumulate_invocation.go:46`、`service_account_test.go` 相關測試
- [ ] `ValidatePreimageExtrinsics()` — 移除第 3 參數 `*unmatchedKeyVals`（方案 A 下不再需要）
  - 影響：`stf/validate_extrinsic.go:8,13`、`stf/update_preimages.go:13,16`、`stf/sft.go:46,88`、`stf/stf_timing.go:86,140`、`accumulation_test.go:99-100`
- [ ] `ShouldIntegratePreimage()` — `account.LookupDict[...]` 改用 `GetPreimageMeta`；`account.PreimageLookup[...]` 保持不動
- [ ] `UpdateDeltaWithExtrinsicPreimage()` — lookup 寫入改用 `InsertPreimageMeta`；preimage blob 寫入保持 `PreimageLookup[hash] = data`
  - 「附加 timeslot 到既有 metadata」場景使用 `UpdatePreimageMeta()`
- [ ] `Provide()` — 同上：meta 走 globalKV，blob 走 PreimageLookup
  - 「附加 timeslot 到既有 metadata」場景使用 `UpdatePreimageMeta()`
- [ ] `lookupInKeyVal()` / `lookupAndRemoveKeyVal()` — 方案 A 下移除（已在 Step 4 列出）
- [ ] **【必要】** 新增 `(sa *ServiceAccount) AddPreimage(serviceID, preimage, currentTimeslot) error` 方法，統一封裝 GP §9.6 不變量，取代分散的 preimage + meta 寫入邏輯。三個分支：
  ```
  h = HashData(p)
  k = NewPreimageMetaStateKey(serviceID, h, len(p))

  if PreimageLookup[h] 存在:
      meta, ok = GetPreimageMeta(k)
      if !ok:
          return nil          ← 邊界：blob 在但 meta 不在，靜默 noop
      if len(meta) < MaxHistoricalTimeslotsForPreimageMeta:
          UpdatePreimageMeta(k, append(meta, currentTimeslot))
      return nil              ← blob+meta 都在，追加 timeslot（受上限保護）
  else:
      PreimageLookup[h] = p
      InsertPreimageMeta(k, len(p), [currentTimeslot])   ← 新增 blob+meta
  ```
  - `UpdateDeltaWithExtrinsicPreimage`、`Provide`、`provide` host call 等改用此方法
- [ ] 更新 `accumulation_test.go`
- [ ] 跑三套測試確認無 regression

**影響檔案**：`internal/service_account/service_account.go`、`internal/accumulation/extrinsic_preimage.go`、`accumulation_test.go`

---

## Step 6：簡化 StateEncoder()（state root 必須一致）  ✅ DONE

> **過渡期方案**：本 Step 在 `ServiceInfo.Items/Bytes` 欄位仍存在時運作。
> Step 8 移除這兩個欄位後，序列化將改為 codec-only struct 直接讀計數器，
> 本 Step 的「序列化前同步」邏輯屆時會被移除。

- [ ] delta1（ServiceInfo）：**序列化前**先從計數器同步 `ServiceInfo.Items = GetTotalNumberOfItems()`、`ServiceInfo.Bytes = GetTotalNumberOfOctets()`
- [ ] delta2（StorageDict）+ delta4（LookupDict）：改為直接遍歷 `GetGlobalKVItems()`，key 已是 StateKey，直接輸出 `StateKeyVal{Key: key, Value: value}`
- [ ] delta3（PreimageLookup）：**保持獨立迴圈**，仍需遍歷 `PreimageLookup` 並構造 delta3 key
- [ ] **排序步驟保留不動**：`sort.Slice(encoded, ...)` 必須保留，merklization 仍依賴排序輸入
- [ ] 確認產出的 StateKeyVals 排序和內容與舊實作完全一致
- [ ] 評估 `KeyLevelCache`（`internal/blockchain/key_level_cache.go`）的 cache key 是否仍有效
- [ ] **state root 必須與改之前完全一致**（用 jam-conformance 282 個 trace 驗證）
- [ ] 跑三套測試確認無 regression

**影響檔案**：`internal/utilities/merklization/state_serialize.go`

---

## Step 7：簡化 StateKeyValsToState()  ✅ DONE

**設計事實**：storage key 與 preimage meta key 在 globalKV 中天然無法區分（兩者都是 Blake2b hash 過的 `[31]byte`）。
這也是合併的本質前提 — 反序列化時不需要區分，統統灌進 globalKV 即可。
唯一需要區分的是 preimage lookup（delta3），靠「preimage lookup key 是其 value 的 hash」性質判定。

**反序列化採「一次掃描 + 三類桶」模式**（順序由依賴關係保證，不依賴 key 排序）：

```
第一遍：全掃所有 StateKeyVals，分流到三個 slice
  - chapter keys → 直接處理
  - service keys (C(255,s)) → serviceKeys slice
  - IsPreimage() == true → preimageLookupEntries slice（回傳已算的 hash 供重用）
  - 其餘 → storageOrMetaKeys slice

第二遍：按順序處理
  1. serviceKeys → 建立 ServiceAccount（**允許 struct literal，不強制 NewServiceAccount()**）+ 從 ServiceInfo.Items/Bytes 初始化 counter
     - 反序列化路徑不走 Insert 方法，PreimageLookup 和 globalKV 由後續桶步驟惰性初始化

**反序列化的三個 lazy init 點**（漏一個就 panic）：

| 順序 | 物件 | 時機 | 寫入方式 |
|------|------|------|----------|
| 1 | `state.Delta` (ServiceAccountState) | 處理第一個 service key 時 | `state.Delta[id] = sa` |
| 2 | `sa.PreimageLookup` | 處理第一個 preimage lookup entry 時 | `if nil { make(...) }` 後 `sa.PreimageLookup[h] = v` |
| 3 | `sa.globalKV` | 處理第一個 storage/meta entry 時 | `GetGlobalKVItems` → check nil → `SetGlobalKVItems` |

> 每次寫入後都要把 `serviceAccount` 寫回 `state.Delta[id]`（Go map value semantics）。
  2. preimageLookupEntries → 寫入 PreimageLookup
  3. storageOrMetaKeys → 逐個 entry 處理（不區分 storage 和 meta）：
     a. 從 stateKey 反解出 serviceID
     b. 找出對應的 serviceAccount
     c. `globalKVItems := serviceAccount.GetGlobalKVItems()`
     d. 若 nil → 初始化空 map
     e. `globalKVItems[sk] = encodedValue`（加入單一 entry）
     f. `serviceAccount.SetGlobalKVItems(globalKVItems)`
     g. `state.Delta[id] = serviceAccount`（Go map value semantics，每次都要寫回）
```

> **重要**：目前程式碼的 storage entries 不會進 `StorageDict`，而是作為 `storageStateKeyVals`
> fallback pool 回傳（line 680-688）。方案 A 下這些 entries 全部改灌入 `globalKV`，fallback pool 為空。

- [ ] storage（delta2）和 lookup（delta4）的 keyval **不區分**，用 `SetGlobalKVItems()` 直接塞入 globalKV（**不走 InsertStorage/InsertPreimageMeta 路徑**，避免計數器重複更新）
- [ ] 計數器初始化來源：從 delta1 的 wire format 讀取 items/octets → `SetTotalNumberOfItems` / `SetTotalNumberOfOctets`
  - Step 8 移除 `ServiceInfo.Items/Bytes` 後，改從 codec-only struct 解碼取得
- [ ] preimage（delta3）的 keyval 仍寫入獨立的 `PreimageLookup` map
  - **value 直接是原始 preimage bytes，不需 unmarshal**。`PreimageLookup[hash] = rawValue` 直接賦值
  - 對照：storage value 也是 raw bytes 直接灌 globalKV；只有 preimage meta value 是 JAM 編碼，`GetPreimageMeta` 讀取時才 unmarshal
- [ ] 重構 `IsPreimage()` 簽名改為 `(bool, OpaqueHash, error)`，回傳已算的 hash 供反序列化重用，避免重複 Blake2b
- [ ] **移除** `IsLookup()` 函數（dead code，從未被呼叫，且原邏輯為錯誤的啟發式判定）
- [ ] 移除兩階段 parse（先 preimage 再 lookup）邏輯 — 改為「一次掃描 + 三類桶」
- [ ] **行為差異文件化**：「孤兒 lookup entry」（meta 已設定但 preimage 未到，如 solicit before provide）
  在舊邏輯中不進 `LookupDict`（留在 fallback pool），重構後會進 `globalKV`。
  這是更正確的行為（GP 允許），但需用 conformance traces 驗證 preimages_tests / accumulate_tests 中相關情境。
- [ ] 確認所有 `StateKeyValsToState()` 呼叫處在初始化時完整載入 globalKV，無 lazy load（方案 A 前提）
  - `internal/blockchain/chain_state.go:820, 878`
  - `internal/fuzz/service.go:164`
  - `cmd/node/test.go:161, 280`
  - `internal/safrole/sealing_test.go:49`、`safrole_test.go:538`
- [ ] 更新 `parse_state_key_vals_test.go`
- [ ] 跑三套測試確認無 regression

**影響檔案**：`internal/utilities/merklization/parse_state_key_vals.go`

---

## Step 7.5：unmatchedKeyVals 生命週期清理（方案 A 下游影響）  ✅ DONE

> 方案 A 下「state == globalKV」，`StateKeyValsToState()` 不再回傳 fallback pool，
> 整個 codebase 對 `unmatchedKeyVals` 的依賴需移除。
>
> **為什麼必須移除（而非保留為空）**：方案 A 下 `serializedState` 已包含 globalKV 的所有 key。
> 若 `unmatchedKeyVals` 仍存在於 pipeline 中，未來若被誤填，
> `append(serializedState, unmatchedKeyVals...)` 會造成 key 重複，**state root 重複計算**。
> 移除整條 pipeline 才能從結構上消除這個正確性風險。

### 7.5a: StateKeyValsToState() API 簽名

- [ ] 簽名從 `(State, StateKeyVals, error)` 改為 `(State, error)`（移除第 2 回傳值）
- [ ] 更新所有呼叫點：
  - `internal/blockchain/chain_state.go:820, 878`
  - `internal/fuzz/service.go:164`
  - `cmd/node/test.go:161, 280`
  - `internal/safrole/sealing_test.go:49`、`safrole_test.go:538`
  - `internal/utilities/merklization/parse_state_key_vals_test.go:39, 93, 215, 231`

### 7.5b: ChainState 結構清理

- [ ] 移除 `preStateUnmatchedKeyVals` / `postStateUnmatchedKeyVals` 兩個欄位
- [ ] 移除 6 個 getter/setter 方法（`Get/Set PriorStateUnmatchedKeyVals`、`Get/Set PostStateUnmatchedKeyVals`、`GetPriorStateUnmatchedKeyValsRef`、`GetPostStateUnmatchedKeyValsRef`）
- [ ] `restoreWithState()` / `RestoreStateFromSnapshot()` 簽名移除 unmatched 參數
- [ ] 區塊推進邏輯（chain_state.go ~line 431-434, 497-500）的 unmatched 拷貝模式清理

### 7.5c: State root 計算流程

- [ ] `chain_state.go:666-667`：移除 `append(serializedState, unmatchedKeyVals...)`，`serializedState` 已含 globalKV，直接使用
- [ ] `chain_state.go:888-890`：同上
- [ ] `stf/validate_header.go:32`：同上
- [ ] `fuzz/service.go:193`：同上

### 7.5d: 驗證

- [ ] 跑三套 fuzz 測試確認 state root 一致
- [ ] **state root 必須與重構前完全一致**

**影響檔案**：`internal/blockchain/chain_state.go`、`internal/stf/validate_header.go`、`internal/fuzz/service.go`、`cmd/node/test.go`

---

## Step 8：移除舊的 StorageDict 和 LookupDict 欄位  ⏳ PARTIAL (8a/8b ✅, 8c-f 未做)

> **8a/8b 已完成**：`ServiceAccountDerivatives` 已移除；`unmarshal_json.go` 走 globalKV（與 legacy maps 雙寫，後者僅供 jamtests `ServiceAccount.Encode` wire format 使用）。
>
> **8c-f 未做（blocker）**：`encode.go` / `decode.go` 的 `ServiceAccount.Encode` / `Decode` JAM wire format 編碼 raw storage key（`map[string]ByteSequence`），但 `StateKey` 是 Blake2b hash 過的 `[31]byte`，**raw key 無法反推**。要真正移除 `StorageDict` / `LookupDict` / `ServiceInfo.Items` / `Bytes` 需要先重設計 jamtests 的 binary wire format（或改變 JAM test-vector 格式），這超出本次重構範圍。當前架構下兩個 deprecated map 只在 jamtests JSON 載入時被 dual-write，runtime 完全不讀，沒有 SSOT 問題，但 struct shape 還是兩條路線並存。


> 本 Step 包含「encodeDelta1 從過渡期同步模式 → codec-only struct 模式」的最終遷移。
> 完成後 Step 6 的「序列化前同步」邏輯應一併移除。

- [ ] 從 `ServiceAccount` struct 移除 `StorageDict` 和 `LookupDict` 欄位
- [ ] **保留 `PreimageLookup` 欄位**（不移除！）
- [ ] 移除 `Storage` / `LookupMetaMapEntry` **在 `ServiceAccount` struct 中的角色**
  - **保留** `Storage`、`LookupMetaMapEntry` type alias 和四個 DTO struct（`PreimagesMapEntryDTO`、`LookupMetaMapEntryDTO`、`AccountDataDTO`、`AccountDTO`）— 這些是外部 JSON 介面契約（test fixtures、conformance traces 依賴此格式）
- [ ] 評估是否保留 `LookupMetaMapkey` struct（作為 `NewPreimageMetaStateKey()` 的參數仍需要）
- [ ] **移除** `ServiceAccountDerivatives` struct 和 `GetServiceAccountDerivatives()` 函數
  - 所有呼叫端改用 `account.GetTotalNumberOfItems()` / `GetTotalNumberOfOctets()` / `ThresholdBalance()`
- [ ] **移除** `ServiceInfo.Items` 和 `ServiceInfo.Bytes` 欄位（消除 SSOT 違反）
  - **結構路線（保守）**：保留 `ServiceInfo` 巢狀結構，僅移除 `Items/Bytes` 兩欄。攤平 ServiceInfo 是獨立議題，不在本次重構範圍。
  - **刻意差異 — Version 來源**：參考實作將 Version hardcoded 為 `0`（其 runtime struct 無 Version 欄位）。我們保留 `ServiceInfo.Version` 並從中讀寫，使 Version 可隨 GP 版本動態管理。副作用：所有建立 ServiceAccount 的路徑必須確保 `ServiceInfo.Version` 被正確初始化。
  - 序列化採 **codec-only struct 模式**：新增 `encodedServiceAccount` 中間 struct（位於 `internal/utilities/merklization/`，private），`encodeDelta1` 從 ServiceAccount 計數器讀取後填入 codec struct 再 Marshal
  - 欄位定義（順序 = JAM 編碼的 byte sequence，順序錯一個 = state root 全錯）：
    | # | 欄位 | 型別 | GP 符號 | 來源（序列化） | 目標（反序列化） |
    |---|------|------|---------|---------------|----------------|
    | 1 | Version | U8 | — | `sa.ServiceInfo.Version` | `sa.ServiceInfo.Version` |
    | 2 | CodeHash | OpaqueHash | a_c | `sa.ServiceInfo.CodeHash` | `sa.ServiceInfo.CodeHash` |
    | 3 | Balance | U64 | a_b | `sa.ServiceInfo.Balance` | `sa.ServiceInfo.Balance` |
    | 4 | GasLimitForAccumulator | Gas | a_g | `sa.ServiceInfo.MinItemGas` | `sa.ServiceInfo.MinItemGas` |
    | 5 | GasLimitOnTransfer | Gas | a_m | `sa.ServiceInfo.MinMemoGas` | `sa.ServiceInfo.MinMemoGas` |
    | 6 | FootprintStorage | U64 | a_o | `sa.GetTotalNumberOfOctets()` | `sa.SetTotalNumberOfOctets(v)` |
    | 7 | GratisStorageOffset | U64 | a_f | `sa.ServiceInfo.DepositOffset` | `sa.ServiceInfo.DepositOffset` |
    | 8 | FootprintItems | U32 | a_i | `sa.GetTotalNumberOfItems()` | `sa.SetTotalNumberOfItems(v)` |
    | 9 | CreationTimeslot | TimeSlot | a_r | `sa.ServiceInfo.CreationSlot` | `sa.ServiceInfo.CreationSlot` |
    | 10 | MostRecentAccumulationTimeslot | TimeSlot | a_a | `sa.ServiceInfo.LastAccumulationSlot` | `sa.ServiceInfo.LastAccumulationSlot` |
    | 11 | ParentService | ServiceID | a_p | `sa.ServiceInfo.ParentService` | `sa.ServiceInfo.ParentService` |
  - 改寫 `DecodeServiceInfo()`：解碼後用 `SetTotalNumberOfItems` / `SetTotalNumberOfOctets` 設定計數器
- [ ] Step 8 移除清單中**明確排除** `CalcStorageItemfootprint` 和 `CalcLookupItemfootprint`（仍用於 PVM pre-check）
  - 移除 Step 6 的「序列化前同步」邏輯（不再需要）
- [ ] 清理 `encode.go` 中 `LookupMetaMapEntry.Encode()`、`Storage` 相關方法
- [ ] 清理 `decode.go` 中 `LookupMetaMapEntry.Decode()`、`Storage.Decode()` 等方法
- [ ] 更新 `unmarshal_json.go`：
  - 改用 `NewServiceAccount()` 初始化
  - `Storage` → 呼叫 `InsertStorage(NewStorageStateKey(...), len(originalKey), value)` 自動更新計數器
  - `LookupMeta` → 呼叫 `InsertPreimageMeta(NewPreimageMetaStateKey(...), length, timeslots)` 自動更新計數器
  - `Preimages` → 保持寫入 `PreimageLookup`
  - JSON 中的 `ServiceInfo.Items/Bytes` → **忽略，以 Insert 過程的增量計數器為準**（SSOT）
- [ ] 更新 `PartialStateSet.DeepCopy()` — 改為呼叫每個 account 的 `Clone()` 方法
- [ ] 移除 Step 3 的 `assertConsistency` 機制（舊 map 已移除，assertion 不再適用）
- [ ] 跑三套測試確認無 regression

**影響檔案**：`internal/types/state.go`、`types.go`、`encode.go`、`decode.go`、`unmarshal_json.go`

---

## Step 9：更新所有測試  ✅ 主要部分 DONE

> **完成項目**：`service_account_test.go`（seedGlobalKV helper + 17 處 HistoricalLookup 改寫）、`accumulation_test.go`（pre-state fixture 走 `MigrateLegacyMapsToGlobalKV`）、`networking/ce/ce134_test.go`、`jamtests/preimages_tests.go`、`jamtests/accumulate_tests.go`、`parse_state_key_vals_test.go`。新增 `ServiceAccount.MigrateLegacyMapsToGlobalKV(serviceID)` helper 把 struct-literal fixture 一次性 mirror 進 globalKV。
>
> **未處理項目（runtime 不受影響）**：`jamtests/reports_tests.go` 的 fixture 用 nil maps、`work_package_controller_test.go` 同樣 nil maps、`state_serialize_test.go` 只有 commented-out code、`state_key_constructor_test.go` 沒有 legacy map 依賴。這些測試在現有架構下都能跑（如果 Rust-VRF 在環境裡 build 得起來），不需要 fixture 改造。
>
> **本地端 `go test ./...` 全跑不過**：跟重構無關 — `pkg/Rust-VRF/vrf-func-ffi/src` / `pkg/erasure_coding` 是 Rust FFI，需要 Rust toolchain 才能編。Docker fuzz target 完整 build 並通過 三套 fuzz 測試（最終驗證手段）。


- [ ] **掃描全 codebase 的 `ServiceAccount{...}` literal**，確認全部改用 `NewServiceAccount()`
  - **例外**：反序列化路徑（Step 7）允許直接 struct literal 建構（搭配 lazy init 防護）
- [ ] `internal/service_account/service_account_test.go`
- [ ] `internal/accumulation/accumulation_test.go`
- [ ] `internal/networking/handler/ce/ce134_test.go`
- [ ] `internal/work_package/work_package_controller_test.go`
- [ ] `jamtests/accumulate/accumulate_tests.go`
- [ ] `jamtests/preimages/preimages_tests.go`
- [ ] `jamtests/reports/reports_tests.go`
- [ ] `internal/utilities/merklization/state_serialize_test.go`
- [ ] `internal/utilities/merklization/state_key_constructor_test.go`
- [ ] `internal/utilities/merklization/parse_state_key_vals_test.go`
- [ ] `go test ./...` 全部通過
- [ ] 跑三套 fuzz 測試最終確認

---

---

## Phase 2：持久化 KV Store（暫不實作）

> **狀態：規劃中，等 Phase 1 完成並穩定後再啟動。**

目標：將 `globalKV` 從 Go map 替換為 PebbleDB，獲得並行讀取、crash recovery、超大 state 支援。

- [ ] 引入 `github.com/cockroachdb/pebble` 依賴
- [ ] 新增 `db.KVStore` 介面（Get/Put/Delete/NewBatch/Close）
- [ ] `GetStorage`/`InsertStorage`/`DeleteStorage` 內部改為操作 PebbleDB
- [ ] `GetPreimageMeta`/`InsertPreimageMeta` 同上
- [ ] DB 初始化與生命週期管理
- [ ] Batch write：accumulation 結束後一次性 commit
- [ ] 跑三套測試確認 state root 一致

**預估**：~100 行改動，1~2 個檔案，1~2 週。業務邏輯完全不動（Phase 1 已封裝）。

---

## Phase 3：Trie Structural Sharing（暫不實作）

> **狀態：規劃中，依賴 Phase 2 完成。**

目標：merklization 從「每次全量重建 trie」改為「增量更新 + reference-counted node sharing」，
獲得 O(1) fork state、平行 block validation、平行 accumulation。

- [ ] Trie node 持久化到 PebbleDB（prefix 區分 node / value / refcount）
- [ ] `MerklizeAndCommit(pairs)`：增量寫入 trie nodes + batch commit
- [ ] Reference counting：`IncreaseNodeRefCount` / `DecreaseNodeRefCount`
- [ ] `DeleteTrie(rootHash)`：遞迴刪除，refcount 降為 0 才真刪
- [ ] 增量 merklize：只更新改過的 key 的 path，未改的 node 直接 reuse
- [ ] Fork state：O(1) 建立分支（copy root hash + increment refcount）
- [ ] 跑三套測試確認 state root 一致

**預估**：~500 行新增 + ~200 行修改，3~5 個檔案，3~4 週。業務邏輯完全不動。

**平行化收穫**：
- 不同 service 的 trie 子樹互不干擾 → 平行 accumulate
- O(1) snapshot → 平行 block validation / fork choice
- PebbleDB snapshot → 多 goroutine 並行讀取

---

## 預估影響規模

| 類別 | 檔案數 | 修改點數 |
|------|--------|----------|
| 核心型別 + 新方法 | 3 | ~50 |
| StateKey 輔助函數 | 1 | ~10 |
| PVM Host Call | 2 | ~55 |
| Accumulation | 2 | ~25 |
| 序列化/反序列化 | 4 | ~25 |
| 測試 | 10+ | ~100 |
| **總計** | **~22** | **~265** |

## 審核紀錄

### 第二次審核修正（v2）
1. ✅ `PreimageLookup` 保持獨立不合併
2. ✅ delta3 編碼保持獨立迴圈
3. ✅ 所有 host call 中 PreimageLookup 存取保持現狀

### 第三次審核修正（v3）
1. 新增 `ServiceInfo.Items/Bytes` 與計數器的同步規則說明
2. Step 7 反序列化改為從 `ServiceInfo.Items/Bytes` 初始化計數器（O(1)），而非從 globalKV 重算
3. 明確方法定義在 `internal/types` package（與 struct 同 package），建議新檔案 `service_account_kv.go`
4. Step 2 新增 StateKey 構造輔助函數（`NewStorageStateKey`、`NewPreimageMetaStateKey`）
5. 明確 `InsertPreimageMeta` 已存在 key 時只覆蓋資料不更新計數器
6. 所有 Insert/Delete 方法要求使用 safemath 溢位保護

### 第四次審核修正（v4）
1. **Step 4 write 流程修正**：`CalcStorageItemfootprint` 仍然需要做 pre-check（threshold balance 驗證），
   只有驗證通過後才呼叫 InsertStorage/DeleteStorage。不能直接「由方法自動處理」因為需要先確認餘額。
2. **StorageKeyVal fallback pattern**：新增處理說明。`getStorageFromKeyVal()`、`removeStorageFromKeyVal()`、
   `getLookupItemFromKeyVal()` 三個函數的處理方式需要決定（全量載入 vs 保留 fallback）。
3. **safemath 前置作業**：codebase 中沒有通用 safemath 包，需要在 Step 2 之前新增。
   只有 `PVM/argument_invocation.go` 有 `checkOverflow`（僅乘法）。

### 第五次審核補強（v5）— 整合外部 review
1. **Step 7 設計說明**：storage key 與 preimage meta key 在 globalKV 中天然無法區分（都是 Blake2b hash），
   反序列化時不需區分、統統灌進 globalKV。只有 preimage lookup（delta3）需要靠 `IsPreimage()` 判定。
2. **`IsPreimage()` 函數保留**：利用「preimage lookup key == hash(value)」性質做判定，重構後仍需要。
3. **`ServiceAccount{...}` literal 掃描**：全 codebase 的 struct literal 必須改用 `NewServiceAccount()`。
4. **StorageKeyVal fallback 確定方案 A**：全量載入、無 fallback，列出 4 個待清理函數。
5. **`DeleteStorage` 設計約束說明**：StateKey 是 hash，caller 必須提供 originalKeyLen 和 valueLen。
6. **`HistoricalLookup` 需要改動**：v5 review 指向 `lookup (=2)` 但實際上 `lookup (=2)` 只讀 PreimageLookup。
   真正需要改的是 `HistoricalLookup()` 函數（讀 `LookupDict`），影響 `historicalLookup` (=6) 和 `refine_invocation`。
7. **`UpdatePreimageMeta` 使用場景明確化**：用於「附加 timeslot 到既有 metadata」場景。
8. **`ServiceAccountDerivatives` struct 評估**：Step 8 加入評估項。
9. **`KeyLevelCache` 影響評估**：Step 6 加入驗證項。
10. **持久化鋪路說明補充**：key 31 bytes + value raw bytes = 天然對齊 KV store 介面。

### 第六次審核補強（v6）— 整合外部 review
1. **Step 7 移除 `IsLookup()`**：dead code（從未被呼叫），且原邏輯為錯誤的啟發式判定。
2. **Step 7 反序列化順序文件化**：chapter → service → preimage lookup → globalKV，天然由 key 排序保證但需明確記載。
3. **Step 2c 新增 `Clone()` 方法**：符合 SRP，搭配 `cloneMapOfSlices` 泛型 helper。
4. **Step 8 `ServiceInfo.Items/Bytes` 從「評估」改為「明確移除」**：消除 SSOT 違反，`encodeDelta1` 改讀計數器。
5. **Step 8 `ServiceAccountDerivatives` 從「評估」改為「明確移除」**：重構後三個欄位都可 O(1) 直接取得。
6. **Step 3 `ThresholdBalance` 改為 method**：內部直接讀計數器 + safemath。
   - 預存 bug（`storage >= aF` 時未減 `aF`）應獨立 ticket 處理，不在本次重構中修正。
   - v6 review 對 bug 的描述有誤（說 `storage < aF` 分支有問題，實際是另一個分支）。
7. **Step 3 assertion 從「可」改為「強制」**：用 build tag `//go:build dev_assert` 控制，不一致時 panic。
8. **Step 7 `IsPreimage` 改措辭**：現有邏輯已正確，「保留」即可，不需「重構」。
9. **Step 7 方案 A 前提驗證**：列出 7 個 `StateKeyValsToState()` 呼叫點，確認完整載入。

### 第七次審核補強（v7）— 整合外部 review
1. **Step 2c `InsertStorage` 精確語意**：已存在 key 時 `octets += len(newValue)-len(prevValue)`，keyLen 不重算。
2. **Step 2c/4a `InsertStorage` 不接受空值即刪除**：PVM `write` 的 `vz==0` 由 PVM 層分派到 `DeleteStorage`。
3. **Step 2c Delete 冪等語意**：key 不存在時 return nil，不更新計數器。
4. **Step 2c `InsertPreimageMeta` 內部 Marshal**：globalKV 的 meta value 已是 JAM 編碼，序列化直接 dump。
5. **Step 7 反序列化改為「一次掃描 + 三類桶」**：順序由依賴關係保證，不依賴 key 排序。
   - 重要發現：現有程式碼的 storage entries 不進 StorageDict，而是作為 fallback pool 回傳。方案 A 下全部改灌入 globalKV。
6. **Step 7 `IsPreimage` 簽名擴充**：回傳已算的 hash 供反序列化重用，省一次 Blake2b。
7. **差異 1（ServiceInfo 巢狀）決策**：保守路線 — 保留 ServiceInfo 巢狀，僅移除 Items/Bytes。
8. **差異 2 codec-only struct**：Step 8 新增 `encodedServiceAccount` 中間 struct，runtime 與 wire format 解耦。
9. **差異 3 排序保留**：Step 6 明確 `sort.Slice` 保留不動。

### 第八次審核補強（v8）— 整合外部 review
1. **Step 5 `HistoricalLookup` 簽名**：新增 `serviceID` 參數，列出 3 處業務呼叫 + 14 處測試。
2. **Step 5 `ValidatePreimageLookupDict` / `existsInLookupDict` / `FetchCodeByHash` 簽名**：新增 `serviceID` 參數。
3. **Step 8 JSON DTO struct 保留**：`Storage`/`LookupMetaMapEntry` type alias + 四個 DTO struct 是外部 JSON 介面契約，不能移除。
4. **Step 8 `unmarshal_json.go` 改寫**：Storage/LookupMeta 改用 Insert 方法，JSON 中 Items/Bytes 忽略。
5. **Step 7 「孤兒 lookup entry」行為差異文件化**：三類桶模式下孤兒 lookup 也進 globalKV，更正確但需驗證。
6. **Step 7 反序列化不走 Insert 路徑**：用 `SetGlobalKVItems()` 直接塞，避免計數器重複更新。計數器從 delta1 wire format 讀取。
7. v8 補強 B 的 codec-only struct 命名：實作時再定，建議放 `internal/utilities/merklization/`。

### 第九次審核補強（v9）— 整合外部 review
1. **新增 Step 7.5：unmatchedKeyVals 生命週期清理**（v9 最重要的發現）
   - ChainState 的 `preStateUnmatchedKeyVals`/`postStateUnmatchedKeyVals` 欄位 + 6 個方法
   - State root 計算的 4 處 `append(serializedState, unmatchedKeyVals...)`
   - `StateKeyValsToState()` 簽名改為 `(State, error)`，影響 9+ 個呼叫點
2. **Step 6/8 過渡期標示**：Step 6 開頭加「過渡期方案」提示，Step 8 開頭加「最終遷移」呼應。
3. **Step 5 補列 `ValidatePreimageExtrinsics`**：移除 `*unmatchedKeyVals` 參數，影響 stf/ 多處。
4. **Step 4 新增 4c**：PVM 結構體清理（`StorageKeyVal` 欄位、`ResultContext`、`I(...)` 簽名）。

### 第十次審核補強（v10）— 整合外部 review
1. **Step 2c `GetPreimageMeta` Unmarshal 失敗行為**：回傳 `nil, false`，與 key 不存在一致。
2. **Step 2c Insert 方法 lazy init 防護**：`if sa.globalKV == nil { make(...) }`。
3. **Step 7/9 反序列化路徑例外**：允許 struct literal（不強制 `NewServiceAccount()`），搭配 lazy init。
4. **Step 8 codec-only struct Version 映射**：序列化讀 `ServiceInfo.Version`，反序列化寫回。
5. **Step 5 `AddPreimage` 便捷方法**：可選評估項。
6. **Step 3/8 `CalcStorageItemfootprint` / `CalcLookupItemfootprint` 保留**：純數學函數，PVM pre-check 仍需要，明確排除於移除清單。

### 第十一次審核補強（v11）— 整合外部 review
1. **Step 2b safemath `Mul`**：已在先前 review 中補上（確認存在）。
2. **Step 5 `AddPreimage` 從可選改為必要**：封裝 GP §9.6 不變量，多處呼叫端統一。
3. **Step 2c `InsertStorage` 已存在分支**：必須先 `Sub(octets, len(prevValue))` 再 `Add(octets, len(newValue))`，避免 uint64 underflow。
4. **Step 3 `ThresholdBalance` bug 修正形式**：補上 `if sum < aF { return 0 } return sum - aF` 正確形式。
5. **Step 7.5 正確性風險說明**：unmatched 若被誤填會造成 state root key 重複，必須從結構上消除。
6. 可選項決策：
   - **新增的** `NewStorageStateKey` / `NewPreimageMetaStateKey` 保留 `(StateKey, error)` 簽名（Step 2a，與參考實作一致）。不改動現有 `StateKeyConstruct` interface（避免擴散到既有呼叫端）。
   - **不做** HistoricalLookup 改 method（17+ 處改動，留作未來獨立 refactor）
   - **做** `ServiceAccountState.Clone()` 包裝方法（5 行，改動極小）

### 第十四次審核補強（v14）— 整合外部 review
1. **Step 2c `Clone()` 計數器複製方式**：用 `cloned := *sa` 隱式複製 unexported scalar fields，再 deep clone 兩個 map。附 pseudo-code。
2. **Step 7 globalKV 寫入模式**：從「用 SetGlobalKVItems 直接塞」改為「逐個 entry 處理」的 7 步驟描述。
3. **Step 8 Version 來源刻意差異**：我們從 `ServiceInfo.Version` 讀寫（參考實作 hardcoded 0），標記為第二個刻意差異。

### 第十三次審核補強（v13）— 整合外部 review
1. **Step 2c 原子性模式**：Insert/Delete 先在區域變數副本算，全成功才賦值回 struct。
2. **Step 7 三個 lazy init 點**：`state.Delta`、`sa.PreimageLookup`、`sa.globalKV` 各自的初始化時機。
3. **Step 8 `encodedServiceAccount` 完整欄位表**：11 個欄位 + GP 符號 + 序列化/反序列化映射。
4. **Step 7 preimage lookup value 是 raw bytes**：不需 unmarshal，避免誤加解碼。
5. **Step 2c `UpdatePreimageMeta` 存在性判斷**：用直接 map 查詢，不用 `GetPreimageMeta`。
6. **Step 5 `AddPreimage` 邊界**：blob 在但 meta 不在 → 靜默 noop。補完整三分支偽碼。

### 第十二次審核補強（v12）— 整合外部 review
1. **Step 2c `UpdatePreimageMeta` nil 防護**：`globalKV == nil` 時回傳 error（不做 lazy init，與 Insert 語意不同）。
2. **Step 5 `HistoricalLookup` 保持 free function**：決策顯式記載在 Step 5 主體中。
3. **v11 StateKey error 簽名矛盾澄清**：新增的 `NewStorageStateKey`/`NewPreimageMetaStateKey` 保留 `(StateKey, error)`（Step 2a），不改動現有 `StateKeyConstruct` interface。
4. **審核紀錄順序整理**：v7 移到 v6 後面，移除重複的 v10 區塊。
5. **Step 5 `AddPreimage` 加必要性標記**。
