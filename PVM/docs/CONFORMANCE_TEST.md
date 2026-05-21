# PVM Conformance 測試指南

本文件說明如何驗證 **interpreter（ground truth）** 與 **recompiler（test target）** 在 fuzz conformance 與 PVMtrace 兩個層級上的一致性。

修改 `PVM/` 後，建議依序執行：

1. **Fuzz conformance** — 端到端 state root 是否與測試向量一致
2. **PVMtrace diff** — 逐指令比對，定位第一個語意分歧

---

## 架構概覽

```text
                    ┌─────────────────────────────────────┐
                    │  jam-conformance fuzz trace JSON    │
                    │  pkg/test_data/.../traces/<id>/   │
                    └──────────────┬──────────────────────┘
                                   │ test_folder (unix socket)
              ┌────────────────────┴────────────────────┐
              ▼                                         ▼
   ┌──────────────────────┐                 ┌──────────────────────┐
   │  Fuzz Target Server  │                 │  Fuzz Client         │
   │  (先啟動，聽 socket)  │◄───────────────│  go run ./cmd/fuzz/  │
   └──────────────────────┘                 └──────────────────────┘

   Ground truth : --pvm-backend interpreter   （任意平台）
   Test target  : --pvm-backend recompiler    （僅 linux/amd64）
```

| 測試層級 | 目的 | Ground truth | Test target |
|---|---|---|---|
| Fuzz conformance | 區塊 import 後 state root / key-val 正確 | interpreter | recompiler |
| PVMtrace | 逐指令 pc/opcode/reg/mem 一致 | interpreter trace | recompiler trace |

**重要差異**

- Fuzz conformance 的 recompiler 使用 **BlockBasedInvoke**（production 路徑），映像為 `new-jamneration-target:latest`（無 `trace` build tag）。
- PVMtrace 的 recompiler 使用 **DebugSingleStepInvoke**（`trace` build tag 自動啟用），映像為 `new-jamneration-target:trace`。
- 兩層都失敗時，先用 PVMtrace 找第一個 `dst_val` 分歧，再回頭修 recompiler emit/執行邏輯。

---

## 前置需求

- Go 1.25+（本機 interpreter 測試）
- Docker Desktop + buildx（Apple Silicon 上建置 linux/amd64 recompiler 映像）
- 測試資料：`pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces/`

```bash
# 確認版本
cat VERSION_GP VERSION_TARGET
```

---

## 一、建置 Docker 映像（recompiler）

修改程式碼後 **必須重建**。在 repo 根目錄執行：

### Production fuzz target（conformance 用）

```bash
docker buildx build --platform=linux/amd64 \
  --build-arg GP_VERSION="$(cat VERSION_GP)" \
  --build-arg TARGET_VERSION="$(cat VERSION_TARGET)" \
  --build-arg OUTPUT=new-jamneration-target \
  -t new-jamneration-target:latest \
  -f docker/Dockerfile --load .
```

或使用 Makefile（建議在 arm64 Mac 上手動加上 `--platform`）：

```bash
make fuzz-docker-build
```

### Trace 映像（PVMtrace 用）

```bash
make fuzz-docker-build-trace
# 產出 new-jamneration-target:trace
```

---

## 二、Fuzz Conformance Test

流程：**先啟動 target server → 再用 client 送 `test_folder`**。

### 2.1 Interpreter（ground truth，任意平台）

**Terminal 1 — 啟動 server**

```bash
make run-target
# 等同：
# mkdir -p .jam_fuzz_docker_run
# JAM_FUZZ=1 JAM_FUZZ_SPEC=tiny \
#   JAM_FUZZ_DATA_PATH=.jam_fuzz_docker_run/ \
#   JAM_FUZZ_SOCK_PATH=.jam_fuzz_docker_run/fuzz.sock \
#   go run ./cmd/fuzz/
```

預設使用 `interpreter` backend（`cmd/fuzz/main.go` 的 `--pvm-backend` 預設值）。

**Terminal 2 — 執行測試**

```bash
# 單一 trace 資料夾
go run ./cmd/fuzz/ test_folder \
  ./.jam_fuzz_docker_run/fuzz.sock \
  ./pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces/<TRACE_ID>/

# 整個 traces 目錄（所有子資料夾）
go run ./cmd/fuzz/ test_folder \
  ./.jam_fuzz_docker_run/fuzz.sock \
  ./pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces/
```

**預期結果**：所有 JSON 顯示 `PASSED`（含預期 protocol error 的 case，例如 `author index is out of range`）。

### 2.2 Recompiler（test target，僅 linux/amd64 Docker）

使用 named volume 共享 unix socket（`jam-sock`）。

**Terminal 1 — 啟動 recompiler server**

```bash
docker volume rm jam-sock 2>/dev/null
docker volume create jam-sock

docker run --rm --platform=linux/amd64 --name jam-server \
  -v jam-sock:/sockdir \
  new-jamneration-target:latest \
  /sockdir/jam_target.sock -pvm-backend recompiler
```

**Terminal 2 — 執行測試**

```bash
docker run --rm \
  -v jam-sock:/sockdir \
  -v "$(pwd)/pkg/test_data":/data:ro \
  new-jamneration-target:latest \
  test_folder /sockdir/jam_target.sock \
  /data/jam-conformance/fuzz-reports/0.7.2/traces/
```

或只測單一資料夾：

```bash
docker run --rm \
  -v jam-sock:/sockdir \
  -v "$(pwd)/pkg/test_data":/data:ro \
  new-jamneration-target:latest \
  test_folder /sockdir/jam_target.sock \
  /data/jam-conformance/fuzz-reports/0.7.2/traces/<TRACE_ID>/
```

**判定**

- `FAILED!!` + `state_root mismatch` / `mismatch count` → recompiler 語意錯誤
- interpreter 通過、recompiler 失敗 → 優先進入 PVMtrace 除錯

---

## 三、PVMtrace 逐指令比對

腳本 `scripts/run_pvmtrace_fuzz_capture.sh` 會自動：

1. 建置 `new-jamneration-target:trace`（可用 `SKIP_DOCKER_BUILD=1` 跳過）
2. 分別用 interpreter / recompiler 重放同一 trace 資料夾並錄製 trace
3. 執行 `pvm-diff find-diff`

### 3.1 一鍵擷取 + diff

```bash
# 預設 trace 資料夾
make pvmtrace-fuzz-capture

# 指定失敗 case
make pvmtrace-fuzz-capture \
  TRACE_FOLDER=pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces/1776775377 \
  DEBLOB_JSON=pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces/1776775377/00000412.json
```

輸出目錄：

- `./pvmtrace-out/interp/<serviceID>_<codehash>/` — interpreter ground truth
- `./pvmtrace-out/recomp/<serviceID>_<codehash>/` — recompiler candidate

### 3.2 手動 diff 指令

```bash
# 找第一個分歧
go run -tags trace ./cmd/pvmtrace/ pvm-diff find-diff \
  --left  ./pvmtrace-out/interp/<service>_<hash> \
  --right ./pvmtrace-out/recomp/<service>_<hash>

# 檢視分歧附近步驟
go run -tags trace ./cmd/pvmtrace/ pvm-diff show \
  --left ./pvmtrace-out/interp/<service>_<hash> \
  --right ./pvmtrace-out/recomp/<service>_<hash> \
  --from 3415 --limit 30

# 單步詳情
go run -tags trace ./cmd/pvmtrace/ pvm-diff detail \
  --left ./pvmtrace-out/interp/<service>_<hash> \
  --right ./pvmtrace-out/recomp/<service>_<hash> \
  --step 3443
```

### 3.3 解讀 pvm-diff 輸出

| stream | 意義 | 優先級 |
|---|---|---|
| `dst_val` | 指令執行後目的暫存器值不同 | **最高** — 真實語意錯誤 |
| `gas` | gas 計數不同 | 高 |
| `pc` / `opcode` | 控制流分歧 | 高 |
| `loads` / `stores` | 記憶體存取紀錄值不同 | 中 — 見下方注意 |

**`loads`/`stores` 假陽性（已知）**

interpreter trace 記錄的是 **記憶體原始值 / 截斷後寫入值**；recompiler debug trace 目前記錄的是 **sign-extend 後的暫存器值**。因此 signed load（如 `load_ind_i32`）可能出現 `loads` 分歧但 `dst_val` 一致。

**除錯時請以 `dst_val` 分歧為準**，不要只看 `loads` stream。

規格細節見 `PVM/plan/PVM_Modular_Trace.md`。

---

## 四、建議除錯流程

```text
1. 確認 interpreter 通過目標 trace 資料夾
        ↓ 若失敗 → 先修 interpreter / 測試資料，不是 recompiler 問題
2. 確認 recompiler fuzz conformance 失敗的 JSON 檔名（如 00000412.json）
        ↓
3. make pvmtrace-fuzz-capture TRACE_FOLDER=... DEBLOB_JSON=<失敗 json>
        ↓
4. pvm-diff find-diff → 記下 step、pc、opcode、stream
        ↓
5. 若 stream=loads 但 dst_val 相同 → 查更後面的 dst_val 分歧
        ↓
6. pvm-diff show/detail 檢視分歧前後 store/load 鏈
        ↓
7. 對照 PVM/instructions.go（interpreter）與 PVM/recompiler/emit_*.go（JIT emit）
        ↓
8. 修復後：重建 Docker → 重跑步驟 2 和 3
```

### 快速掃描 dst_val 分歧（本機）

```bash
cat > /tmp/trace_dst_scan.go <<'GO'
package main
import (
  "fmt"
  "github.com/New-JAMneration/JAM-Protocol/PVM/PVMtrace/diff"
)
func main() {
  left, right := "./pvmtrace-out/interp/<service>_<hash>", "./pvmtrace-out/recomp/<service>_<hash>"
  for step := int64(0); ; step++ {
    l, e1 := diff.ReadStepAtPublic(left, step)
    r, e2 := diff.ReadStepAtPublic(right, step)
    if e1 != nil && e2 != nil { fmt.Println("total steps", step); return }
    if l.DstVal != r.DstVal {
      fmt.Printf("dst_val diff step=%d pc=0x%x op=%d L=%016x R=%016x\n", step, l.PC, l.Opcode, l.DstVal, r.DstVal)
      return
    }
  }
}
GO
go run -tags trace /tmp/trace_dst_scan.go
```

---

## 五、其他相關測試

```bash
# Recompiler 單元測試（linux/amd64 Docker）
make build-recompiler-test-env
make run-recompiler-test

# ASM 測試
make run-asm-test
```

---

## 六、常見問題

### Q: recompiler server 啟動失敗 `pvm-backend "recompiler" is not available`

映像必須是 **linux/amd64** 且含 CGO JIT。請用 `docker buildx build --platform=linux/amd64 ...` 重建。

### Q: `operation not permitted` / buildx 失敗

確認 Docker Desktop 已啟動，必要時在 Cursor 外層終端執行（需完整 Docker 權限）。

### Q: fuzz socket 連不上

1. 確認 server 先於 client 啟動
2. interpreter：socket 在 `.jam_fuzz_docker_run/fuzz.sock`
3. recompiler：socket 在 volume `jam-sock` 的 `/sockdir/jam_target.sock`
4. 兩個 container 必須掛載同一 volume

### Q: 改了 code 但結果不變

Production 與 trace 是 **兩個映像**。conformance 改動後需重建 `new-jamneration-target:latest`；PVMtrace 需重建 `new-jamneration-target:trace`。

### Q: `pvm-diff` 出現 `platform (linux/amd64) does not match host (linux/arm64)` warning

`new-jamneration-target:trace` 映像是以 `--platform=linux/amd64` 建置的。在 Apple Silicon 上若執行 `docker run` 時**未加** `--platform=linux/amd64`，Docker 會用 arm64 相容層跑 amd64 映像並印出此 warning。

解法：對所有使用該映像的 `docker run` 加上 `--platform=linux/amd64`。`scripts/run_pvmtrace_fuzz_capture.sh` 的 `pvm-diff` 步驟已包含此參數；本機可直接用 `go run -tags trace ./cmd/pvmtrace/` 避免跨架構 emulation。

### Q: `deblob metadata not found`

不影響 `pvm-diff find-diff` 主流程；若要 opcode 名稱對照，確認 `DEBLOB_JSON` 路徑正確且對應區塊含 PVM program blob。

---

## 附錄：現有 Makefile 捷徑

| 目標 | 說明 |
|---|---|
| `make run-target` | 本機 interpreter fuzz server |
| `make fuzz-docker-build` | 建置 production 映像 |
| `make fuzz-docker-build-trace` | 建置 trace 映像 |
| `make pvmtrace-fuzz-capture` | 錄製雙 backend trace + pvm-diff |
| `make run-recompiler-test` | recompiler 單元測試 |

---

## 附錄：2026-06 除錯紀錄（trace `1776775377`）

| 項目 | 結果 |
|---|---|
| Interpreter fuzz | 7/7 PASSED |
| Recompiler fuzz | `00000412.json`、`00000414.json`、`00000416.json` FAILED（state_root mismatch） |
| pvm-diff 首個 `loads` 分歧 | step=3422，`load_ind_i32`，dst_val 相同（trace 記錄格式差異） |
| pvm-diff 首個 `dst_val` 分歧 | step=3443，`load_ind_u64` pc=0x17731，addr=0xfefdf9e0 |
| 分歧值 | interpreter `0x00e4ac3eace643d9` vs recompiler `0x0019ac3eace643d9`（高位 byte 不同） |

後續修復方向：檢查 step 3420–3441 對 stack 區域 `0xfefdf9e0` 的 store 鏈，以及 `PVM/recompiler/emit_memory.go` 中 `store_ind` / `load_ind` 的截斷與 emit 邏輯。
