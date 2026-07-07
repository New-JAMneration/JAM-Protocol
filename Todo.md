# JAM Protocol Networking Todo

參考 [JAMNP-S 規格](https://github.com/zdave-parity/jam-np/blob/main/simple.md) 與 [Strawberry](https://github.com/eigerco/strawberry) 實作，列出 `internal/networking/` 目前缺少的項目。

---

## Issue #966：CE 128 STF Import（進行中）

Branch: `966-feat-ce128-stf-import`

- [x] Step 1：`ImportBlock` — parent 檢查 + `stf.RunSTF` + `StateCommit`
- [x] Step 1：`SyncManager.storeBlocks` 改走 `ImportBlock`
- [x] Step 1：單元測試（parent mismatch、失敗不發 `BlockImported`）
- [x] Step 2：head / finalized 更新（UP0 Final → peer.Finalized、ImportBlock SetCurrentHead）
- [x] Step 3：bulk sync 多輪 loop（追到 networkBest 或無進度）
- [ ] Step 4：錯誤處理 / retry policy
- [ ] Step 5：整合測試（mock CE128 responder）
- [ ] Step 6：dual-node E2E（#567）

---

- [x] 建立 `cmd/node` networking bootstrap 組裝入口（集中註冊與啟動）
- [x] 啟動 `Peer.Start` 並註冊 CE 最小 handler（CE 128 + CE 129）
- [x] 注入 `SyncManager` 依賴（blockchain + shared EventBus）並啟動訂閱
- [x] 加入 role detection stub（keystore 有 Ed25519 key 視為 validator，否則 full）
- [x] 加入 `--role full|validator` flag（省略時自動偵測 keystore）
- [x] 依 chainspec `bootnodes` 進行啟動連線嘗試（跳過自身）
- [x] 加入 graceful shutdown（signal cancel + peer close + sync manager close）

---

## 零、Testnet 最小可用清單（V=6，1 own + 5 official nodes）

> 目標：讓我們自己的一個 validator node 成功加入由 5 個官方節點組成的 testnet，能連線、同步 block、執行基本 validator 職責。

### Grid 結構（V=6）

```
W = floor(sqrt(6)) = 2

Row 0: index 0, 1      Col 0: index 0, 2, 4
Row 1: index 2, 3      Col 1: index 1, 3, 5
Row 2: index 4, 5

例：若我方 validator 是 index 0，grid 鄰居 = {1, 2, 4}（同 row 或同 col）
UP 0 stream 只需要在這三個鄰居（+cross-epoch same-index）之間開啟
```

---

### Phase 1：能連上官方節點

#### [x] TLS 憑證使用 validator 的固定 Ed25519 私鑰 ✅
- `cert.TLSConfigFromPrivateKey(sk, isServer, isBuilder)` — 從 SK 建立 stable TLS cert
- `PeerConfig.PrivateKey ed25519.PrivateKey` — 唯一 identity 入口，移除冗餘的 `PublicKey` 欄位
- `NewPeer()` 直接從 SK derive PK，不再有 ephemeral random key 路徑
- **測試**：`TestTLSConfigFromPrivateKey_leafPublicKeyMatches` 驗證 cert leaf PK 與 SK 一致

#### [x] TLS custom certificate verifier 完整化 ✅
- `verifyPeerCertificate` + `ValidateX509Certificate` 已走自訂 JAMNP 規則（Ed25519、單一 DNS SAN、無 IP SAN）
- `TestMutualAuthenticatedQUICWithFixedEd25519Keys` 兩節點 mutual auth smoke test 已存在
- builder 連線槽位：`Peer.handleConnection` 限制 `MaxBuilderConnections = 20`

#### [x] MergeValidators() ✅
- `internal/networking/validator/merge.go`：`MergeValidators` / `TransportTargets` / `GossipPeerKeys`
- `RefreshFromChain()` 於 epoch 套用時更新 grid 狀態

#### [x] `ce_request.go` 的 `Decode()` ✅
- `internal/networking/handler/ce/ce_request_decode.go`：`Decode` / `DecodePayload` + roundtrip 測試

#### [ ] 自動連線目標集合整合（prev/current/next epoch）
`NewQuicConfig()` 已加入 `MaxIdleTimeout: 30 * time.Minute`。

#### [x] 自動連線排程（Preferred Initiator 整合）✅
`Peer.StartValidatorConnections(ctx, neighbors, selfKey)` 已實作：
- `PreferredInitiator == selfKey` → 立即 dial
- 否則 → 等 5 秒後 dial（讓對方優先）
- 跳過自己（selfKey == v.Ed25519）
- ctx 取消時 goroutine 立即退出

---

### Phase 2：能同步 Block（UP 0 Block Announcement）

> UP 0 目前完全沒有實作（沒有 `handler/up/` 目錄）。`DefaultUPHandler` 只有 `EncodeMessage` 工具函數，不是真正的 UP 0 stream 協議。

**規格 Wire Format**：
```
Final = Header Hash (32B) ++ Slot (u32 LE)
Leaf  = Header Hash (32B) ++ Slot (u32 LE)
Handshake = Final ++ len++[Leaf]       // 所有已知 leaves（finalized block 的後代中無子塊者）
Announcement = Header (SCALE) ++ Final

連線建立後雙向同時進行：
--> Handshake
loop { --> Announcement }

<-- Handshake
loop { <-- Announcement }
```

#### [ ] UP 0 Handshake 實作
連線建立後雙方各送一次 Handshake：目前最終化 block (Final) + 所有已知 leaves。
**檔案**：`handler/up/up0.go`（新增）

#### [ ] UP 0 Announcement 迴圈
收到或產出新的合法 block 時，在所有 UP 0 stream 廣播 Announcement。
跳過條件：後代已 announced、block 不是 finalized 後代、對方已 announced。
**檔案**：`handler/up/up0.go`

#### [ ] UP 0 duplicate stream 處理（Acceptor 側）
同一 peer 若有多個同 kind 的 UP 0 stream，保留 stream ID 最大者，其餘 reset/stop。
**檔案**：`handler/up/up0.go`、`quic/peer.go`

#### [ ] UP 0 + Grid 整合
只有在雙方互為 grid 鄰居（`ValidatorManager.IsNeighbor()` 為 true），或至少一方非 validator 時，才開啟 UP 0 stream。
**檔案**：`quic/peer.go`、`handler/up/up0.go`

#### [ ] Block Receipt Hook → CE 128 自動 fetch
UP 0 收到 Announcement 後，透過 EventBus 觸發 CE 128 拉完整 block。
EventBus 已有 `PeerUpdated` event type，需連結到 CE 128 呼叫。
**檔案**：`quic/event_bus.go`、`quic/peer.go`

---

### Phase 3：能執行 Validator 職責

#### [ ] UpdateCoreAssignments（知道自己負責哪個 core）
依規格 §11.3 的 permutation 計算當前 epoch 的 core-to-guarantor 對應，讓節點知道要處理哪個 core 的 work-package。
**檔案**：`validator/manager.go`

#### [ ] CE 133 → CE 134 → CE 135 流程串接
目前三個 handler 各自獨立，收到訊息後無後續動作。需串接：
1. CE 133 收到 work-package → 大小驗證 → 呼叫 CE 134 sharing
2. CE 134 完成 refinement → erasure coding → 儲存所有 shards → 觸發 CE 135 distribution
**檔案**：`handler/ce/ce133.go`、`handler/ce/ce134.go`、`handler/ce/ce135.go`

#### [ ] Epoch Transition 連線管理（見第五節）
Epoch 轉換時，需等待：(1) 新 epoch 第一個 block finalized，(2) 距 epoch 開始 ≥ max(⌊E/30⌋, 1) slots，才套用 connectivity 變更並同步調整 UP 0 streams。

#### [ ] CE 146 Work-Package Bundle Submission（未實作）
詳見第一節。

#### [ ] CE 148 Segment Request（未實作）
詳見第一節。

---

### Phase 4：不影響接入但需修正的正確性問題

#### [ ] CE 134 並發安全（sync.RWMutex）
`currentAssignedCore` 需用 `sync.RWMutex` 保護，提供 `SetCurrentCore(core uint16)` 動態更新。

#### [x] PeerSet 加入 by-ValidatorIndex 索引 ✅
已完成。`internal/networking/quic/peer_set.go` 已有：
- `byEd25519Key`
- `byAddr`
- `byValidatorIdx`
- `GetByValidatorIndex(idx)`

#### [ ] MergeValidators()
Epoch 轉換時合併 prev/current/next epoch validator 列表（by Ed25519 key 去重），維持連線穩定性。

---

## 一、尚未實作的 CE Protocols

### [ ] CE 146: Work-Package Bundle Submission
**規格位置**: JAMNP-S § CE 146
**方向**: Builder → Guarantor
**說明**: Builder 在能提供 imported segments 時，應使用此協議（而非 CE 133）提交完整的 work-package bundle。
**Wire format**:
```
--> Core Index ++ Segments-Root Mappings (len++[WP Hash ++ Segments-Root])
--> Work-Package
--> [Extrinsic]
--> [Segment] (所有 imported segments)
--> [Import-Proof] (len++[Hash])
--> FIN
<-- FIN
```
**需要新增**:
- `internal/networking/handler/ce/ce146.go`
- `ce_request.go` 加入 `WorkPackageBundleSubmission CERequestID = 146`

---

### [ ] CE 148: Segment Request
**規格位置**: JAMNP-S § CE 148
**方向**: Guarantor → Guarantor（或 Builder → Guarantor）
**說明**: Guarantor 向其他 Guarantor 請求 import segments 以完成 work-package bundle。無法提供時 fallback 到 CE 139/140。
**Wire format**:
```
--> [Segments-Root ++ len++[Segment Index]]
--> FIN
<-- [Segment]
<-- [Import-Proof] (len++[Hash])
<-- FIN
```
**限制**: 單次請求 Segment 數量不超過 W_M（= 3072）
**需要新增**:
- `internal/networking/handler/ce/ce148.go`
- `ce_request.go` 加入 `SegmentRequest CERequestID = 148`

---

## 二、UP 0: Block Announcement（完全未實作）

見第零節 Phase 2，此處列出所有子任務：

### [ ] UP 0 Handshake 訊息實作
### [ ] UP 0 Announcement 迴圈
### [ ] UP 0 重複 stream 處理
### [ ] UP 0 與 Grid Structure 整合
### [ ] Block Receipt Hook → CE 128 觸發

---

## 三、Grid Structure

**規格位置**: JAMNP-S § Grid structure

### [x] Grid 鄰居計算
已完成。`internal/networking/validator/grid.go`：
- `GridMapper.NeighborIndicesInEpoch(index)` — 同 epoch 內 row/col 鄰居
- `GridMapper.IsNeighborInEpoch(a, b)` — 是否為 grid 鄰居
- `GridMapper.IsSameIndexCrossEpoch(index, key)` — 跨 epoch same-index 檢查
- `GridMapper.AllNeighborValidators(index)` — 取得所有鄰居（包含跨 epoch）

### [x] ValidatorManager + PreferredInitiator
已完成。`internal/networking/validator/manager.go`：
- `ValidatorManager.GetNeighbors()` — 取得需連線的鄰居清單
- `ValidatorManager.IsNeighbor(key)` — 判斷是否為 grid 鄰居
- `PreferredInitiator(a, b)` — P(a,b) 公式
- `PeerAddressFromMetadata(meta)` — 從 validator metadata 前 18 bytes 解碼 IPv6:port

---

## 四、Validator Connection Management（不完整）

**規格位置**: JAMNP-S § Required connectivity

### [x] PeerAddressFromMetadata
已完成。`validator/manager.go:53`：讀取 metadata 前 16 bytes 為 IPv6，17-18 bytes 為 port（LE u16）。

### [x] Preferred Initiator 邏輯
已完成（公式與排程）。`Peer.StartValidatorConnections(ctx, neighbors, selfKey)` 已整合立即 dial / 5 秒 fallback。

### [x] 自動連線目標集合整合（prev/current/next epoch）✅
`internal/networking/topology/manager.go` reconcile loop：
- `TransportTargets` / `GossipPeerKeys` 分離 transport 與 grid gossip
- `StartValidatorConnections` + Preferred Initiator 整合
- epoch transition 延遲 `max(floor(E/30), 1)` slots
- stale peer prune + builder slot policy
- validator node 啟動時 `cmd/node/network_bootstrap.go` 啟動 topology manager

### [x] Builder 連線槽位保留 ✅
為 work-package builders 保留約 20 個連線槽位（`quic.MaxBuilderConnections`）。Builder 身份驗證（有效 work-package 證明）仍待 #968。

---

## 五、Epoch Transition 連線變更（缺少）

**規格位置**: JAMNP-S § Epoch transitions

### [ ] Epoch 轉換時的連線變更延遲邏輯
等待兩個條件都滿足後才套用：
1. 新 epoch 的第一個 block 已完成 finalization
2. 距離 epoch 開始至少已過 `max(⌊E/30⌋, 1)` 個 slots

### [ ] Epoch 轉換時同步調整 UP 0 streams
套用 connectivity 變更的同時，開啟/關閉對應的 UP 0 streams。

---

## 六、現有實作待修正的問題

### [x] CE 145: Guarantee 元素數量驗證 ✅
已完成。`decodeGuaranteeBytes()` 會在配置 slice 前驗證 signatures 數量介於 `types.GuaranteeMinCount` 與 `types.GuaranteeMaxCount`（2 到 3）之間，`CE145Payload` 也會保存 invalid judgment 的 `Guarantee`。

### [ ] CE 133: 加入 CE 146 的使用指引
**問題**: 規格說明當 builder 能提供 imported segments 時應使用 CE 146 而非 CE 133。
**修正**: 在 CE 133 handler 的文件/註解中說明與 CE 146 的關係。

### [x] `ce_request.go` 的 `Decode()` ✅
已實作於 `ce_request_decode.go`（含 `DecodePayload` 與 roundtrip 測試）。

---

## 七、CE Protocol 常數補充

### [ ] 更新 `ce_request.go` 加入新 protocol IDs
```go
WorkPackageBundleSubmission CERequestID = 146
SegmentRequest              CERequestID = 148
```
（CE 147 `BundleRequest` 已加入）並在 `DefaultCERequestHandler.Encode()` 的 switch-case 中加入對應分支。

---

## 八、Strawberry 已實作但我們尚未完成的功能

### CE 133 後續流程整合（缺少）
我們的 `ce133.go` 在解析完訊息後沒有後續邏輯，缺少 CE 133 → CE 134 → CE 135 的完整流程串接。

### [ ] Work Package 大小驗證（CE 133）
### [ ] Imported Segments 批次抓取（CE 133）
### [ ] PackageBundleBuilder — Work-Package Bundle 組裝（CE 133）

### [ ] CE 134 完成後觸發 Erasure Coding 並儲存所有 Shards
CE 134 handler 完成 refinement 後需呼叫 erasure coding 並儲存所有 shards，否則後續 CE 137/138/139/140 無法正確提供 shards。

### [ ] CE 134 Core Index 並發安全（`sync.RWMutex`）
### [ ] CE 134 Core Index 一致性驗證（收到的 coreIndex 需與本節點當前分配一致）

### [ ] Ticket Store 持久化（CE 131/132）
目前只有轉發邏輯，沒有票券的儲存層。

### [ ] ValidatorService 抽象層
建立統一的 service layer interface，解耦各 handler 中的業務邏輯。

### [ ] Justification discriminator type 2（Segment Shard）
目前只支援 discriminator `0` 和 `1`，缺少 type `2`（CE 139/140 特有）。

### [ ] SegmentShardSize 常數確認（潛在 Bug）
**重要**: 依規格 `Segment Shard = [u8; 4104 / R]`（R = 342，每個 shard = 12 bytes），我們在 CE 139/140 使用 `HashSize = 32 bytes` 切割 segment shards 可能不符合規格。

### [ ] Context 取消支援（Read/Write）
`quic.Stream.ReadMessage()` / `WriteMessage()` 未全面支援 context 傳遞。

### [ ] Block Receipt Hook 機制（onBlockReceiveHooks）
EventBus 已有 `PeerUpdated` event，但尚未實作：
- 自動用 CE 128 拉取完整 block
- 收到 block 後觸發 assurance 分發（CE 141）
- 觸發 auditing 流程（CE 144/145）

### [ ] UpdateCoreAssignments — Guarantor 選取邏輯
定期更新當前 epoch 的 core-to-guarantor 對應（規格 §11.3 permutation）。

### [ ] CE 131 Proxy Validator 驗證封裝
目前只做本地 key 比對，需封裝 `IsProxyValidatorFor(hash)` 依 VRF output 後 4 bytes 計算。

### [x] PeerSet 多索引（by ValidatorIndex）✅
已完成。`PeerSet` 已支援 by Ed25519 key / address / validator index 查找。

### [ ] MergeValidators()
Epoch 轉換時合併兩個 epoch 的 validator peer 列表（by Ed25519 key 去重）。

---

## 九、UP/CE 前置作業 Roadmap（依 networking stack 分層）

> `peering-gap-analysis.md` 是較早期的差異分析；其中部分 `[ ]` 已被後續 PR 補上。這裡改用 networking stack 分層來排接下來的實作順序，避免在 UP/CE 還沒接上前先把 topology 或 validator 業務邏輯混在一起。

### 角色與 networking 關係

### [ ] Node vs Validator Node 邊界釐清
`node` 是一般節點執行體，負責 chain state、block import、networking stack、sync manager、資料庫與事件匯流；`validator node` 是啟用 validator 身分的 node，額外具備 validator key、validator index、active set membership、guarantor/auditor/assurer 職責。

Networking layer 應該先支援一般 node 的能力，再由 validator node 啟用額外 routing / protocol：
- 一般 node：QUIC/TLS identity、stream dispatch、CE 128 sync、基本 peer 管理。
- Validator node：required connectivity、Preferred Initiator、epoch transition、UP 0 grid gossip、CE 131/132/135/141/144/145。
- Guarantor role：CE 133/134/137/148 與 work-package bundle / shard storage。
- Builder role：builder ALPN、builder reserved slots、CE 133/146 submission。

### 已補上，可視為舊 plan 已完成

### [x] Listener accept loop + stream kind dispatch
`Peer.Start()` / `acceptLoop()` / `handleConnection()` / `dispatchStream()` 已實作，會 accept QUIC connection、抽出 peer Ed25519 key、讀 1-byte stream kind 並分派到註冊 handler。

### [x] EventBus Subscribe/Publish key 不一致
`EventBus` 已改為 `map[EventType][]Handler`，`Publish(ctx, eventType, event)` 直接使用 `EventType`。

### [x] ConnectionManager 併發鎖
`ConnectionManager` 已加入 `sync.RWMutex`，並提供 `Add` / `Remove` / `All` / `GetByAddr`。

### [x] Peer.Broadcast stream kind + Close
`Peer.Broadcast()` 已改為先寫 stream kind，再 `WriteMessage()`，最後 close stream。

### [x] Validator Grid / ValidatorManager / Peer metadata address
Grid 鄰居、跨 epoch same-index、`ValidatorManager.GetNeighbors()`、`IsNeighbor()`、`PreferredInitiator()`、`PeerAddressFromMetadata()` 已完成。

### Phase 1 — QUIC + TLS Identity

### 本輪實作（ALPN genesis hash + require 測試重構）
- [x] 確認 genesis header 來源路徑：`internal/chainspec` + `internal/blockchain` 既有能力可用
- [x] `cert.ALPNGen()` 改為從 chain state 的 genesis header hash 前綴產生 ALPN（保留無 genesis fallback）
- [x] 新增測試模擬 genesis block，驗證 ALPN 會使用正確 genesis hash prefix
- [x] `internal/networking/cert/*_test.go` 全面改為 `require` 斷言風格

### [ ] TLS custom certificate verifier 完整化
目前 `internal/networking/cert/` 已能產生 Ed25519 self-signed certificate，並檢查 Ed25519 signature algorithm、DNS SAN 格式、public key 與 SAN 一致；但仍需補強：
- client/server 都要明確不走 CA chain 驗證，只走自訂 verifier。
- `DNSNames` 不可為空，應確認 exactly one Ed25519 alternative name，或明確定義多 SAN 的拒絕/接受規則。
- 確認是否移除或拒絕非必要 IP SAN（目前 cert template 有 `127.0.0.1` / `::1`）。
- 加負測：非 Ed25519、SAN 空、SAN 不匹配、多 SAN、self-signed peer 可通過自訂驗證但不依賴 CA。

### [ ] ALPN genesis hash 正式化
目前 `ALPNGen()` 在找不到 genesis 時會 fallback `jamnp-s/0/00000000`。測試可接受，但正式 node 應從 chain state / chainspec 取得 genesis hash，避免不同 chain 誤連。

### [ ] 兩節點 mutual authenticated QUIC smoke test
建立兩個使用固定 Ed25519 key 的 node，驗證：
- server/client 都會驗 peer cert。
- negotiated ALPN 正確。
- extracted peer key 等於 cert public key。
- 不需要 CA root。

### Phase 2 — Stream Runtime

> **目標**：統一 stream runtime，讓 CE 和 UP protocols 可以一致地註冊和調度。

#### Phase 2 收斂結果（2026-05）

- [x] **JAMNP stream kind 對齊**
  - 第一個 byte 為 stream kind
  - UP stream kinds: `0..127`
  - CE stream kinds: `128..255`

- [x] **Peer 層統一 dispatch**
  - `quic.Peer.dispatchStream` 只做：
    1. 讀 stream kind
    2. `handlers[kind]` 查找與分派
  - Peer 不再有 CE/UP 雙路徑分派

- [x] **CE registry 統一**
  - `internal/networking/handler/ce/ce_handler.go` 為 CE registry
  - 透過 `Register(protoID, handler)` + `RegisterAll(peer, blockchain)` 註冊到 unified peer handlers

- [x] **CE handler 簽名統一（無過渡）**
  - `CEHandlerFunc` 目前統一為：
    `func(ctx context.Context, peerKey ed25519.PublicKey, blockchain blockchain.Blockchain, stream *quic.Stream) error`
  - 移除過渡式 `RegisterLegacy` / `WrapLegacyHandler`

- [x] **UP lifecycle 最小框架**
  - `internal/networking/handler/up/up_manager.go` 已提供 per-peer/per-kind 狀態、duplicate resolution、close 行為

- [x] **移除 legacy broadcast/encoder 路徑**
  - 移除 `Peer.Broadcast()` 與其通用 `MessageEncoder` 相依
  - 後續資料打包走 protocol-specific encoder + `stream.WriteMessage`

- [x] **文件更新**
  - `docs/CE_NETWORKING_ARCHITECTURE.md` 已對齊 unified dispatch 與 CE registry 註冊方式

### Phase 3 — Node Networking Bootstrap

### [ ] Node 啟動整合 networking stack
目前 `internal/node/sync_manager.go` 仍有 stub，且 node 啟動流程尚需確認是否完整建立 networking stack：
- 建立 `Peer`
- `Peer.Start(ctx)` 啟動 listener
- 註冊 CE / UP handlers
- 注入 blockchain 給 `SyncManager`
- 呼叫 `setupEventSubscriptions()`
- 依 validator manager / bootstrap peers 啟動連線

### [ ] Node / ValidatorNode role wiring
一般 node 與 validator node 不是兩套 network stack，而是同一套 node networking stack 的不同 role enablement：
- node 預設只啟用 sync / CE 128 / block import 所需能力。
- validator node 啟用 validator identity、required connectivity、UP 0 gossip、validator CE protocols。
- builder connection 使用 builder ALPN 並受 reserved slot policy 控制。

### UP 0 — Block Announcement（可獨立 assign）

### [ ] UP 0 persistent stream protocol
UP 0 應獨立成一個可分派任務，範圍停在「宣告 / 接收 header / 發出 fetch hook」，不要直接包含 CE 128 fetch 或 block import：
- Handshake encode/decode
- Announcement encode/decode
- peer finalized/leaves state
- `shouldAnnounce` 規則
- duplicate UP 0 stream handling
- per-peer UP 0 lifecycle
- valid announcement hook/event

### Phase 4 — CE 128 + Block Import MVP

### [ ] CE 128 + block import 最小同步鏈
UP 0 發出 valid announcement hook 後，這條鏈負責同步完整 block：
- CE 128 client/requestor fetch full block
- decode response
- import block / update chain
- 更新 peer best/finalized state
- 觸發後續 assurance / audit hooks

### [ ] CE 128 client/requestor
目前 CE 128 handler 已有，但 UP 0 需要主動向 peer 發 CE 128 request 的 client path，包含 peer 選擇、request framing、response decode、錯誤處理。

### [ ] Block Receipt Hook 機制
收到完整 block 並 import 後需觸發：
- assurance distribution（CE 141）
- audit announcement/judgment（CE 144/145）
- sync manager state transition

### Phase 5 — Topology Manager

### [x] Transport connectivity set vs Grid gossip set 分離 ✅
`validator.TransportTargets` vs `validator.GossipPeerKeys`。

### [x] ConnectionManager reconcile loop ✅
`topology.Manager.Run` / `ReconcileOnce`：dial missing + prune stale。

### [x] Preferred Initiator + 5 秒 fallback 整合到 reconcile ✅
透過 `Peer.StartValidatorConnections` 由 topology manager 週期呼叫。

### [x] Epoch transition connectivity ✅
延遲套用 + `RefreshFromChain` + prune/dial。

### [x] Builder 20 reserved slots policy ✅
見上。

### [x] `epochclock` 共用時鐘（Scheme B-lite）✅
- `internal/networking/epochclock/`：`TimeslotsFrom`、`ConnectivityApplied`、epoch/Safrole delay helpers
- topology 套用 connectivity 時 publish `ConnectivityApplied` event（供 CE 131/132 / UP 0 使用）
- bootstrap 時亦 publish anchor

### [x] TransportTargets 修正為 prev/current/next 全體 validator ✅
- `MergeValidators` 用於 transport reconcile；grid 鄰居僅用於 `GossipPeerKeys` / UP 0

### [x] UP 0 grid gossip reconcile ✅
- `Peer.ReconcileUP0Streams`：對 grid 鄰居開 UP 0、對非鄰居關閉

### [x] Builder slot eviction ✅
- 滿 20 時 evict 最舊 builder 連線；prune 跳過 builder role

### FG follow-up（#967 之後，branch `fg-safrole-connectivity-orchestration`）

### [x] FG-1：Topology reconcile smoke 文件 + 測試 ✅
- `docs/TOPOLOGY_RECONCILE_SMOKE.md`：多 validator 本地/testnet 驗證步驟
- `scripts/topology-smoke/`：3× Go validator harness（temp keystore + `JAM_TOPOLOGY_SMOKE_CONFIG` mock metadata）
- `validator/metadata.go`、`cmd/node/topology_smoke_config.go`：§1 自動化 smoke（2026-07-06 PASS）
- `topology/manager_test.go`：epoch transition pending 測試

### [x] FG-2：Safrole connectivity 時序編排 ✅
- `internal/networking/safrole/scheduler.go`：訂閱 `ConnectivityApplied` / `BlockImported`
- `ce131.go`：`WaitForwardStep2` 取代 wall-clock E/20 sleep
- `cmd/node/network_bootstrap.go`：validator role 啟動 scheduler + 註冊 CE 131/132 handler

### Phase 6 — Role-specific CE / Validator Duties

### [ ] CE request constants / Encode / Decode 整理
- 補 `CERequestID = 146` / `148`
- 補 `DefaultCERequestHandler.Encode()` 分支
- [x] 實作 `DefaultCERequestHandler.Decode()`（#968 前置，本輪完成）
- 順手修正 `AuditShardReqeust` 拼字或確認是否因相容性暫留

### [ ] CE 133 → CE 134 → CE 135 service layer
需要新增或抽象化 handler 依賴，避免 CE handler 只 parse wire data 後丟棄：
- Work-package size validation
- imported segments fetch / CE 148 fallback
- `PackageBundleBuilder`
- CE 134 refinement 後 erasure coding + shard storage
- CE 135 guarantee distribution client/server flow

### [ ] CE 139/140 correctness audit
`peering-gap-analysis.md` 的 CE 139 mock fallback 仍存在，且 `Todo.md` 已列出 SegmentShardSize 疑點。UP/CE 前置作業應先釐清：
- 找不到 bundle 時是否應回錯，不應回 mock bundle
- `Segment Shard = 4104 / R` 是否應為 12 bytes，而非目前多處使用 `HashSize = 32`
- justification discriminator type 2 是否需要支援

### 需要再分析，不先直接改

### [ ] EventBus WaitFor / Unsubscribe 語意
`WaitFor(ctx, eventType, timeout)` 目前接受 `timeout` 但主要依賴 ctx；`Unsubscribe(eventType)` 會移除該 event type 的所有 handlers。需確認目前使用者是否接受此語意，再決定是否改 API。

### [ ] 統一 networking error / logging
目前 networking 多處仍用 `fmt.Errorf` / `log`。這不是 UP/CE 的 blocker，但做大改前可決定是否一起導入 shared logger 與錯誤型別。

### [ ] `internal/node/network_event.go` 殘留檢查
`peering-gap-analysis.md` 指出它可能與 `quic.EventType` 重複。需先搜尋實際引用，再決定是否移除或保留。

---

## 十、CE 144/145 缺口補齊（已完成）

### [x] CE 145: Guarantee 結構補齊
新增 `CE145GuaranteeSignature`, `CE145Guarantee` 結構；`readGuaranteeMessage` 改為 `(*CE145Guarantee, error)`；`HandleJudgmentAnnouncement_Send` 在 invalid 時寫入第二段 guarantee。

### [x] CE 144: 驗證增強
`validateAuditAnnouncement` 加入 no-show 基本完整性檢查；補充錯誤路徑測試覆蓋。

### [x] 測試補強（ce144_test.go, ce145_test.go）
CE145: invalid+guarantee 正向（2/3 簽章）、數量邊界（0/1/4 應失敗）、invalid 缺 guarantee 應失敗、sender invalid 輸出第二段。
CE144: tranche/evidence 型別不一致失敗、evidence/workReports 數量不一致失敗、空 announcement 失敗。

---

## 完成狀態摘要

| 項目 | 狀態 |
|------|------|
| CE 128 Block Request | ✅ 完成 |
| CE 129 State Request | ✅ 完成 |
| CE 131/132 Safrole Ticket Distribution | ✅ 完成 |
| CE 133 Work-Package Submission | ⚠️ Wire handler 有，缺大小驗證 / imported segments / bundle / CE134 串接 |
| CE 134 Work-Package Sharing | ⚠️ Share/refine/signature 基礎有，缺並發安全 / core index 驗證 / erasure shard storage |
| CE 135 Work-Report Distribution | ⚠️ Encode/decode 基礎有，缺 guarantee 驗證 / 分發整合 / 統一註冊介面 |
| CE 136 Work-Report Request | ✅ 完成 |
| CE 137 Shard Distribution | ✅ 完成 |
| CE 138 Audit Shard Request | ✅ 完成 |
| CE 139 Segment Shard Request | ⚠️ Handler 有，mock fallback / SegmentShardSize 待修 |
| CE 140 Segment Shard Request w/ Justification | ⚠️ Handler 有，mock justification / SegmentShardSize 待確認 |
| CE 141 Assurance Distribution | ✅ 完成 |
| CE 142 Preimage Announcement | ✅ 完成 |
| CE 143 Preimage Request | ✅ 完成 |
| CE 144 Audit Announcement | ✅ 完成（補強驗證與測試）|
| CE 145 Judgment Publication | ✅ 完成（補齊 Guarantee 結構與驗證）|
| CE 146 Work-Package Bundle Submission | ❌ 未實作 |
| CE 147 Bundle Request | ✅ 完成 |
| CE 148 Segment Request | ❌ 未實作 |
| UP 0 Block Announcement | ❌ 未實作（無 handler/up/ 目錄）|
| Grid Structure（計算） | ✅ 完成（validator/grid.go）|
| PeerAddressFromMetadata | ✅ 完成（validator/manager.go）|
| PreferredInitiator 公式 | ✅ 完成（validator/manager.go）|
| Listener accept loop + stream dispatch | ✅ 完成（Peer.Start / dispatchStream）|
| EventBus Subscribe/Publish | ✅ 完成（EventType keyed）|
| PeerSet 多索引 | ✅ 完成（by key / addr / validator index）|
| TLS cert（SelfSignedCertGen + ALPN） | ⚠️ cert/SAN 生成已有，custom verifier 與 ALPN 正式化待補 |
| TLS 使用 validator 固定 Ed25519 key | ✅ 完成（TLSConfigFromPrivateKey + PeerConfig.PrivateKey）|
| Keystore（LocalKeyStore + JIP-5） | ✅ 完成（keystore/local/）|
| QUIC MaxIdleTimeout | ✅ 完成（30 分鐘）|
| 自動連線排程 | ✅ 完成（Peer.StartValidatorConnections）|
| Node networking startup integration | ❌ 未完整整合 |
| Epoch Transition 連線邏輯 | ❌ 未實作 |
| CE Request Decode | ⚠️ 未實作（TODO placeholder）|
| UpdateCoreAssignments | ❌ 未實作 |
| CE 133→134→135 流程串接 | ❌ 未實作 |

---

## PVM Host Call Cache Coherency Bug（ServiceID 0 balance 差 401258）

### 問題
`fuzz-reports/0.7.2/traces/1775225235_4202/00000257.json` 驗證失敗：
- `ServiceID 0` expected balance = `18446744073708628538`
- 實際 balance = `18446744073708227280`
- 差額 = `401258`（剛好等於被 eject 的 service `10878981` 的 balance）

### 根因
PVM `Psi_A` 會維護兩份 caller service account 表述：
1. `newPartialState.ServiceAccounts[serviceID]`（map entry，同時等同 `GeneralArgs.ServiceAccountState` 及 `ResultContextX.PartialState.ServiceAccounts`）
2. `GeneralArgs.ServiceAccount`（local copy pointer）

`eject` host call（`PVM/host_call_accumulate.go:550-555`）只更新了 map，沒有同步 local copy pointer。當後續 host call（例如 `storage`/`write`）從 stale `GeneralArgs.ServiceAccount` 讀回並寫回 map 時，會覆蓋掉 `eject` 的 balance credit，導致差額遺失。

`transfer` 在 L463-466 有做完整 3-way sync（map + ServiceAccountState + ServiceAccount pointer），`eject` 缺失同一模式。

### 步驟
- [x] 1. 定位 bug（`eject` 3-way sync 缺少 `*GeneralArgs.ServiceAccount` 更新）
- [x] 2. 套用修復：在 `eject` 成功分支同步更新 local copy pointer
- [x] 3. `make run-target`（terminal 1）啟動 target
- [x] 4. 執行 fuzz test（terminal 2）跑 `1775225235_4202` 驗證
- [x] 5. 確認 `00000257.json` 通過，整個 7-file trace 不再 FAIL（7/7 PASSED）

### 修復內容（`PVM/host_call_accumulate.go:552-557`）
在 `accountS.ServiceInfo.Balance += accountD.ServiceInfo.Balance` 之後，參照 `transfer` L463-466 的 3-way sync 模式補上：
```go
(*input.Addition.GeneralArgs.ServiceAccountState)[serviceID] = accountS // update general
*input.Addition.GeneralArgs.ServiceAccount = accountS
```

---

## PVM Host Call Cross-Service Read Side-Effect Bug（Storage key 遺失）

### 問題
`fuzz-reports/0.7.2/traces/1776702160_6218/00000640.json` 驗證失敗：
- `state_root mismatch`，差異點在 service `3791505980` 的一個 storage key（prefix `0x3c53ca4bfd66e17b...`）。
- Expected post_state：該 key 仍保留 90-byte value。
- 實際 post_state：key 不見了（value = nil）。
- 同樣症狀也影響 `1775225235_4202`、`1775225235_9287`、`1775746363_9717`、`1776696452_5244`。

### 根因
`read` host call（`PVM/host_call_general.go`）在 cross-service read 時並不是 side-effect free：

1. 當 caller `s` 讀另一個 service `s*` 的 key 時，`serviceID` 會被 reassign 成 `s*`。
2. 若 `a.StorageDict[k]` 不存在但 `UnmatchedKeyVals` pool 裡有值，就會把 value cache 回 `a.StorageDict` 並從 pool 移除，**且用的是被 reassign 後的 `s*`**。
3. `s*` 通常不是當前 accumulating 的 service，所以它那份被更新過的 `ServiceAccount`（內含新 cache 的 key）永遠不會被 `ParallelizedAccumulation` 寫回 `PosteriorStates`。
4. 但 `removeStorageFromKeyVal` 已經把那個 key 從 caller 本地的 `UnmatchedKeyVals` copy 中刪掉了；後續 intersection merge 時該 key 就從全局 pool 徹底消失。

根據 Graypaper，`ΩR` 必須是 side-effect free — 這條路徑違反了規格。

### 步驟
- [x] 1. 加入 `[TARGET_KEY]` log 追蹤該 key 在 `ImportBlock` / `RestoreBlockAndState` / `RunSTF` 前後於 Prior/Post `UnmatchedKeyVals` 中的存在情況
- [x] 2. 確認 key 於 `RunSTF` 後消失；定位至 `ParallelizedAccumulation` 的 intersection 合併邏輯
- [x] 3. 加入 `[READ_CROSS]` log 驗證 cross-service `read` 會把 target service 的 key 從 caller 的 pool copy 中刪除
- [x] 4. 套用修復：在 `read` 只於 `callerServiceID == serviceID`（讀自己的 storage）時才做 cache + pool 移除
- [x] 5. 重跑五個 trace，全部 PASSED
- [x] 6. 清除所有暫時除錯用的 log（`[TARGET_KEY]` / `[SCOPE_3791505980]` / `[READ_CROSS]`）

### 修復內容（`PVM/host_call_general.go`，`read` 函式）
1. 在 reassign `serviceID` 前先保存 `callerServiceID := serviceID`。
2. 當 fallback 到 `UnmatchedKeyVals` pool 取到值時，只有 `callerServiceID == serviceID` 才執行：
   - `a.StorageDict[k] = v`
   - 寫回 `ResultContextX.PartialState.ServiceAccounts[serviceID]`
   - `removeStorageFromKeyVal(...)`
   — 其他情況只把值讀出給 guest，保持 `ΩR` side-effect free。

### 後續潛在風險（尚未修）
`provide` host call（`PVM/host_call_accumulate.go` 約 L885+）有類似的 cross-service 路徑：`s` 可能被 reassign 為 `input.Interpreter.Registers[7]` 指定的 target service，後續 `getLookupItemFromKeyVal` / lookup pool 操作可能在 target 非 accumulating 時造成同類資料遺失。目前 trace 尚未觀察到此 bug 觸發，列為待稽核項目。

---

## Fuzz target Docker 入口標準化（jam-conformance）

目標：讓我們的 target 可以用 Docker image 提交，`cmd/fuzz` 維持 target 語意（root 預設即 serve），並可在本地完成 Docker 驗證。

### 代辦

- [x] `cmd/fuzz/main.go`：維持 root 預設 `Action: serve`（target binary 預設行為就是 fuzz server）
- [x] `cmd/fuzz/main.go`：`serve` **僅在** `JAM_FUZZ` 設置時啟動；**僅**認 `JAM_FUZZ_*` env（socket 不再走寬鬆預設／CLI）
- [x] 移除 mini redis（提交相關）：清理 `USE_MINI_REDIS`（文件 / targets.json 範例）
- [x] `docker/Dockerfile`：補上 fuzz env 預設值，便於本地與 CI 驗證
- [x] 更新文件：`READMERef/RELEASE_AND_PUBLISH.md` 的 target submission 範例改成 Docker image
- [x] 更新 `pkg/test_data/jam-conformance/scripts/targets.json` 的 `new-jamneration` 範例為 Docker image 提交格式
- [x] `cmd/fuzz`: `serve` **必須** `JAM_FUZZ` + `JAM_FUZZ_SPEC` / `JAM_FUZZ_DATA_PATH` / `JAM_FUZZ_SOCK_PATH`（tiny|full）；子命令照常
- [x] `docker/Dockerfile`: `ENV JAM_FUZZ=1`、頭部註解改為 jam-conformance 對等 build/run 範例
- [x] `Makefile`：`JAM_FUZZ_HOST_DIR`（預設 **`.jam_fuzz_docker_run`**）與 `run-target` 與 **Docker 腳本預設 host 路徑一致**；`fuzz-docker-build`、`fuzz-docker-run`，`run-target-docker` → 與 `fuzz-docker-run` 相同
- [x] `scripts/run_fuzz_target_docker.sh`：對等 upstream `docker run`；預設 host 為 **repo 底下 `.jam_fuzz_docker_run/`**（WSL+Doker 可避免 `/tmp` 權限坑），可用 **`JAM_FUZZ_HOST_DATA=/tmp/jam_fuzz`** 對齊純 Linux 用法；含 `--init`、備註
- [x] `cmd/fuzz/main.go`：移除「無參數即 `--help`」行為（否則 Docker 無 CMD 時永遠印說明、不會執行 `serve`）
- [ ] 本地 Docker 驗證：`make fuzz-docker-build` + `make fuzz-docker-run` +（可選）minifuzz smoke test
- [x] `READMERef/PR_MANUAL_TEST_FUZZ_TARGET.md`：PR 可貼的手動測試步驟（與 `.jam_fuzz_docker_run/fuzz.sock` 一致）

---

## Fuzz target：全 in-memory + unfinalized blocks 有界（基於 v0.7.2.14）

> **目標**：`JAM_FUZZ=1` 時不寫任何真實 disk（Pebble/Redis），避免他人機器上 mid-run OOM / 連線 EOF。  
> **前提（已確認）**：每條 trace 僅 **一次 SetState**；設定檔無問題，多為跑一段時間後才斷。  
> **協定**：fuzz 不模擬 finalized chain，只需在單一 server process 內維持可 `RestoreBlockAndState` / `GetState` 的近期區塊鏈。

### 背景與根因（待修）

- [x] **記錄現況**：`PruneOldData` 僅在成功 ImportBlock 後執行；protocol error 在 `AddBlock` 後 return → `unfinalizedBlocks` 可能無限成長。
- [x] **記錄現況**：`ResetInstance()` 與 singleton Pebble 累積問題（fuzz 改為每 reset 新 in-memory repo）。
- [x] **記錄現況**：`JAM_FUZZ_DATA_PATH` 與 Pebble 路徑無關（fuzz 不開 disk）。

### Step 1 — `JAM_FUZZ=1` 強制 in-memory DB（零 disk）

- [x] `newChainStateRepositories()`：fuzz 時 `repo` 與 `persistentRepo` 共用同一 in-memory repository。
- [x] `getPersistentDatabase()`：fuzz 時 safety fallback 為 memory（不開 Pebble）。
- [x] 確認 fuzz 路徑 **不再** 建立/鎖定 `./data/pebble`；本地/Docker sock 跑完 pebble mtime 無成長。

### Step 2 — Fuzz 時 `ResetInstance` 清空 persistent 層

- [x] `ResetInstance()` / `GetInstance()` 透過 `newChainState()` 建立全新 memory DB（fuzz 不共用舊 Pebble）。
- [ ] 單元／手動：連續兩次 `SetState` 後 memory key 數量不應線性累加。

### Step 3 — 有界 `unfinalizedBlocks`（核心）

- [x] 每次 `ImportBlock` 結束 `defer TrimUnfinalizedBlocksForFuzz()` → `KeepRecent(24)`（含 protocol error）。
- [x] 成功路徑維持 `PruneOldData`。
- [ ] 確認 `RestoreBlockAndState(parent)` 在 24 深度內仍可用（手動／全樹 trace）。

### Step 4 — Fuzz 只寫 in-memory `repo`（止血）

- [x] `PersistStateForBlock` / `StateCommitWithPreComputedState`：fuzz 時跳過 `persistentRepo` 寫入。
- [x] `persistBlockMapping`：fuzz 時只寫 `repo`（供 `GetBlockByHash` / restore）。
- [x] `PruneOldData`：fuzz 時只清 `repo`（不碰 disk）。

### Step 5 — 降低 mid-run「假 EOF」（連線被關）

- [x] `fuzz/server.go` `serveOneRequest()`：handler panic `recover`（僅關閉該連線，不殺 process）。
- [x] 盤點：`ImportBlock` protocol error → `ErrorMessage`；`SetState` / STF runtime / decode 失敗仍關連線（協定限制）。
- [x] `validate_fuzz.md` 補充 client `EOF` 常見原因。

### Step 6 — 驗證（fuzz sock only）

- [x] 本地：`JAM_FUZZ=1` server + `test_folder` 跑 **整棵** `pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces`（`CLIENT_QUIET=1`）→ `output.txt` **1179/1179 PASSED**，0 FAILED。
- [x] 監控：全樹跑完無 mid-run EOF；fuzz 期間 `./data/pebble` mtime **未變**（舊 838MB 為修前 node 資料）。
- [x] Docker：`make fuzz-docker-build` + container `new-jamneration-target:latest` + `test_folder` → `output_docker.txt` **1179/1179 PASSED**，約 **7m36s**，0 FAILED。
- [ ] （可選）對照修前後 wall time；功能仍須全 PASS。

### 完成定義

- [x] 長時間 conformance replay **不中斷**（本地 + Docker 全樹 1179 PASS，無 process 死亡 / client EOF）。
- [x] fuzz 全程 **無 disk I/O**（in-memory only；pebble 未在 fuzz 期間更新）。
- [ ] `unfinalizedBlocks` + persisted state 份數 **≤ 24**（或等價有界策略；程式已 `KeepRecent(24)`，缺專門單測／metrics）。

---

## Rust-VRF FFI 記憶體安全 / 效能強化（不更換 ark-vrf 依賴）

> 來源分析：`pkg/Rust-VRF/MEMORY_OWNERSHIP_ISSUES.md`。批次 A~D 全做；`unchecked → checked` 反序列化（批次 E）暫不動，之後另議。

### 批次 A — Go 端記憶體安全（低風險）

- [ ] A1. 所有 FFI 呼叫後對輸入/輸出 `[]byte` 補 `runtime.KeepAlive`（`prover.go`、`verifier.go`、`vrf.go`）
- [ ] A2. `Prover.Free()` / `Verifier.Free()` 內 `runtime.SetFinalizer(x, nil)` 清除 finalizer
- [ ] A3. `Handler.Free()` 釋放後將 `h.Prover` / `h.Verifier` 設 `nil`
- [ ] A4. `Handler.UpdateProverIndex` 改為「先建新、成功才釋放舊」避免失敗時留下半初始化狀態
- [ ] A5. 呼叫端補 `defer handler.Free()`：`sealing.go` L43/81/161、`auditing.go` L70/110/265/300

### 批次 B — Rust FFI panic-safe（需重編 .so）

- [ ] B1. `Verifier::new` 改回傳 `Result`，移除 `.ok_or(()).unwrap()` 與 `.expect(...)`
- [ ] B2. `vrf_verifier_new` / `vrf_prover_new` 建構失敗時回傳 null pointer，不跨 FFI panic
- [ ] B3. 新增 `go_slice_to_slice` helper：`data` null 或 `len<=0` 回傳空 slice，取代裸 `from_raw_parts`

### 批次 C — ring_verifier use-after-free

- [ ] C1. `GetVerifier` epoch 轉換時改為「丟棄舊參考、交給 GC finalizer 釋放」，不在共享中的 verifier 上立即呼叫 `Free()`

### 批次 D — 效能優化（沿用 ark-vrf）

- [x] D1. `Verifier` 建立時快取反序列化後的 `RingCommitment`（`commitment_obj`），省去每次 ring 驗證（含 batch）的反序列化
- [ ] D2. ~~`Prover` 快取 `prover_key`~~ → **略過**：`w3f-ring-proof` 的 `ProverKey` 未實作 `Clone`（`prover()` 以值消耗），且簽章僅由出塊者偶發執行、非熱路徑，成本/風險不划算

### JIP-5 遷移（淘汰 fake_validators.go）

- [x] 新增 `internal/safrole/validator_keys.go`：`DeriveTinyValidator` / `LoadTinyValidatorsData` / `LookupBandersnatchSecretSeed` / `CreateVRFHandler`
- [x] 刪除 `internal/safrole/fake_validators.go`
- [x] Rust FFI 統一 JIP-5 `Secret::from_seed`（`sign_ietf_vrf`、`vrf_get_public_key_from_secret`、`Prover::new_with_secret`）；`from_test_vector` 保留 `from_scalar` 供 JSON test vectors
- [x] `vrf_test.go` 改用本地 JIP-5 helper（避免 `vrf_test → keystore → vrf` import cycle）
- [x] 修正 `TestReplaceOffenderKeys`（先保存 offender Ed25519 再 in-place 清零）
- [x] 修正 auditing 簽章驗證測試（改用 `encoder.Encode` 對齊 `BuildAnnouncement` / `BuildJudgements`）
- [x] 測試通過：`go test ./internal/safrole/...` `./pkg/Rust-VRF/vrf-func-ffi/src/...` `./internal/auditing/...`

### 驗證

- [x] cargo build --release 重編 `libbandersnatch_vrfs_ffi`（通過）
- [x] `cargo test --release` 17/17 通過
- [x] `go test ./pkg/Rust-VRF/vrf-func-ffi/src/... -count=1` 通過
- [x] `go test ./internal/safrole/... ./internal/auditing/... -count=1` 通過
- [~] `go test ./internal/blockchain/...`：僅 `genesis_test.go` FAIL，原因是缺少 `genesis-tiny.bin` 測資（既有問題、與本次改動無關）

### 本輪補強 — Rust unwrap/expect 與 READMERef 使用說明

- [x] R1. `lib.rs` 移除 production path 的 `.unwrap()` / `.expect()`，改以 `Result` / `ok_or` / `map_err` 回傳錯誤
- [x] R2. `lib.rs` 移除 FFI constructor 內的 `unwrap_or` / `unwrap_or_else` 用法，改為明確 `match`
- [x] R3. 不修改 `internal/` 呼叫端；只在文件註明 handler 生命週期與 `defer Free()` 用法
- [x] R4. 在 `READMERef/` 新增 Rust-VRF FFI 使用與安全注意事項
- [x] R5. 重新跑 `cargo test --release` 與 VRF Go 測試

### 本輪補強 — FFI panic 全防護 / 輸出邊界 / verify 錯誤碼 / 私鑰清零

> 對照業界做法（Rustonomicon FFI、zcash #4652、cgo pointer rules）。只動 `pkg/Rust-VRF`，不碰 `internal/`。

- [x] S1. 新增 `ffi_guard` helper，將**所有** `extern "C"` 進入點包進 `catch_unwind`（panic 不跨 FFI，回傳 `Internal` 錯誤碼）
- [x] S2. 輸出寫入加 `output_cap` 容量參數 + null 檢查（`write_output` helper）；同步更新 Rust 簽名、`prover.go`/`verifier.go`/`vrf.go` 的 C prototypes 與呼叫端
- [x] S3. `verify_ietf_vrf` / `get_ietf_vrf_output` / `sign_ietf_vrf` 由 `Result<_, ()>` 改為 `Result<_, VrfError>`；FFI wrapper 改用 `e.to_ffi_error_code()` 回傳精確錯誤碼
- [x] S4. 私鑰位元組複本以 `zeroize` 清零（`new_with_secret` / `from_test_vector` 的 `Vec<u8>`；`Secret` 本身受 ark-vrf 限制，已於文件註明）
- [x] S5. 重編 `.so`（輸出至 `vrf-func-ffi/target/release`）+ `cargo fmt` + `cargo test --release` 17/17 + `go test ./pkg/Rust-VRF/...` 通過 + `internal/safrole` / `internal/auditing` 通過
