# 本地重播（client only）

從 **repository root** 執行。只寫 **重播 client**（`test_trace_folder` / `test_folder`），不含 `polkajam-fuzz`、`fuzz-workflow.py`。

Target 仍需另開（`go run ./cmd/fuzz serve` 或已發佈的 Docker image）；下面只列 **重播命令**。

---

## 建議放進 repo 的資料夾

把 session 裡**可本地重播**的檔案整理成固定路徑，同事 `git pull` 後直接跑：

```
cmd/fuzz/fixtures/
└── l2b/                                    # session 26deee86，L2b full config
    ├── README.session.md                   # 選填：session id、失敗 step、target 映像 tag
    ├── trace/                              # 約最後 1k 步（fuzzer trace ring）
    │   ├── 00079596.bin                    # ring 起點（見下方「步數對照」）
    │   ├── ...
    │   └── 00080578.bin                    # 失敗步
    └── report-mini/                        # 最小復現（僅 2 步 JSON）
        ├── 00080577.json
        └── 00080578.json
```

**來源 session（我們對話裡那輪 L2b）：**

| 項目 | 值 |
|------|-----|
| Session ID | `26deee86e705508988baca0f952db5da` |
| Preset | **L2b**，`jam_spec` / `profile` / `fuzzy_profile` = **full** |
| Target 映像 | `ghcr.io/new-jamneration/new-jamneration-target:v0.7.2.17` |
| 失敗 step | **80578**（80577 PASSED） |
| Fuzzy | `#a44b1af9` → `WriteHostCall`，θ / PVM mismatch |

**trace ring 步數（與 L2a 那包同邏輯）：**  
fuzzer 只留約 **983** 個連續 `.bin`。L2a（ecf65b09）實測為 `13200`…`14182`；L2b 失敗在 `80578`，推定 ring 為 **`79596`…`80578`**（`80578 − 982 = 79596`）。  
從 session 的 `trace/` 整包複製到 `fixtures/l2b/trace/` 即可。

**參考：L2a 長跑包（非 L2b，但同一套 replay 參數已驗過）**

```
cmd/fuzz/fixtures/l2a/trace/   # session ecf65b09，13200.bin … 14182.bin（983 檔）
```

---

## 前置：target（簡短）

L2b 必須 **`JAM_FUZZ_SPEC=full`**。

```bash
SOCK=/tmp/jam_repro/fuzz.sock
DATA=/tmp/jam_repro
mkdir -p "$DATA"

JAM_FUZZ=1 JAM_FUZZ_SPEC=full \
  JAM_FUZZ_DATA_PATH="$DATA" \
  JAM_FUZZ_SOCK_PATH="$SOCK" \
  go run ./cmd/fuzz/ serve
```

診斷 PVM 時可加：`JAM_FUZZ_LOG_LEVEL=debug JAM_FUZZ_PVM_LOG=1`（我們重播 80578 時用過）。

---

## L2b：最小復現（2 步 report）— 已驗證可重現 mismatch

**資料：** `fixtures/l2b/report-mini/`（`00080577.json`、`00080578.json`）

**重點：** full JSON 的 bitfield 長度 43，client 必須 **`--mode full`**，否則解析失敗。

```bash
SOCK=/tmp/jam_repro_regs_80578/fuzz.sock
REPORT=./cmd/fuzz/fixtures/l2b/report-mini

# Terminal A（同上 serve，full spec）
JAM_FUZZ=1 JAM_FUZZ_SPEC=full \
  JAM_FUZZ_DATA_PATH=/tmp/jam_repro_regs_80578 \
  JAM_FUZZ_SOCK_PATH="$SOCK" \
  JAM_FUZZ_LOG_LEVEL=debug JAM_FUZZ_PVM_LOG=1 \
  go run ./cmd/fuzz/ --mode full serve

# Terminal B
go run ./cmd/fuzz/ --mode full test_folder "$SOCK" "$REPORT"
```

**預期：** `00080577.json` PASSED；`00080578.json` FAILED（θ 多寫、`0xf953…` / `0xfff9…` 等 key 與 session 一致）。

> `test_folder` 每目錄第一個檔會 `SetState`，**不是** live fuzz 的「只 init 一次」語意；用來**快速看 STF/PVM 分歧**夠用。

---

## L2b：最後 ~1k 步 live 重播（`test_trace_folder`）

**資料：** `fixtures/l2b/trace/`（`00079596.bin` … `00080578.bin`，或 session 裡實際最小/最大 step）

**語意：** 一次 `SetState` bootstrap → 之後每步只 `ImportBlock`（[`trace_folder.go`](./trace_folder.go)）。

### 從 ring 邊界開始（必讀）

Trace ring **最前 2 步**常因 parent 不在 ring 內而無法 bootstrap。  
**L2a 已驗證：** 目錄含 `13200`…`14182` 時，用 **`--from-step 13202`**（跳過 `13200`、`13201`）→ **981 passed / 0 failed**。

**L2b 推定（同規則）：** 若 ring 從 `79596` 起，先用：

```bash
go run ./cmd/fuzz/ test_trace_folder \
  "$SOCK" \
  ./cmd/fuzz/fixtures/l2b/trace \
  --from-step 79598
```

若 bootstrap 仍失敗，把 `--from-step` 往後調到 log 裡 `ring edge — SetState pre-state before step N` 印出的那個 **N**。

### 常用命令（對話裡實際跑過的參數組合）

```bash
# 1) 整段 live replay（L2a 實測參數；L2b 把路徑與 from-step 換成上表）
go run ./cmd/fuzz/ test_trace_folder \
  "$SOCK" \
  ./cmd/fuzz/fixtures/l2a/trace \
  --from-step 13202

# 2) 只跑尾端 3 步 smoke（L2a：14180–14182）
go run ./cmd/fuzz/ test_trace_folder \
  "$SOCK" \
  ./cmd/fuzz/fixtures/l2a/trace \
  --from-step 14180 --to-step 14182

# 3) 只跑 L2b 失敗步
go run ./cmd/fuzz/ test_trace_folder \
  "$SOCK" \
  ./cmd/fuzz/fixtures/l2b/trace \
  --from-step 80578 --to-step 80578

# 4) soak 第 2 輪起：同一 target 已 bootstrap，跳過 SetState
go run ./cmd/fuzz/ test_trace_folder \
  "$SOCK" \
  ./cmd/fuzz/fixtures/l2a/trace \
  --from-step 13202 \
  --skip-bootstrap
```

（L2a soak：**5 輪 × 981 ImportBlock**，第 1 輪無 `--skip-bootstrap`，第 2–5 輪加 `--skip-bootstrap`。）

### Flags

| Flag | 用途 |
|------|------|
| `--from-step N` | 從 step N 開始（含）；**ring 邊界通常要跳過最前 2 步** |
| `--to-step N` | 到 step N 結束（含） |
| `--skip-bootstrap` | 跳過開局 `SetState`（soak 第 2 輪起） |

---

## 兩種 client 怎麼選

| 目的 | 命令 | 資料夾 |
|------|------|--------|
| 快速重現 **80578 mismatch** | `test_folder --mode full` | `l2b/report-mini/` |
| 重現 **~1k 步累積狀態**（live 語意） | `test_trace_folder` + `--from-step` | `l2b/trace/` |
| 長跑 / EOF / 資源（L2a 參考） | `test_trace_folder --from-step 13202` | `l2a/trace/` |

---

## 實作

| 檔案 | 說明 |
|------|------|
| [`trace_folder.go`](./trace_folder.go) | `test_trace_folder` |
| [`main.go`](./main.go) | `test_folder`、`serve` |
