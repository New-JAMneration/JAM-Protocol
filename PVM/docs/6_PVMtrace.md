# 6. PVMtrace — Per-Instruction Trace 系統

## 概述

PVMtrace 是 PVM 的 per-instruction binary trace 系統，記錄每條指令執行的完整狀態快照。

**目前用途**：per-instruction trace 僅用於本專案內部的 interpreter 與 recompiler 結果驗證——
兩個 backend 對相同 program 執行後，diff 兩份 trace 找出第一個語意分歧點，
快速定位 recompiler emit 的 bug。

**未來方向**：Host-call boundary trace（`host_calls.jsonl.gz`）可用於跨組別互相驗證——
不同團隊的 PVM 實作只需在 host-call 邊界輸出相同格式的 trace，就能比對
registers/gas/memory 的輸入輸出是否一致，無需每條指令都完全對齊。

**目前狀態**：PVMtrace 尚不完整。目前只記錄了 host-call 的 memory read/write（透過
`BeginHostCallMemTrace` 包裝 `GuestMemory`）。完整版需要將 accumulate / refine 等
host-call 的內部行為詳細記錄——例如每次 omega 實際做了什麼操作、service state 變更、
跨 service 呼叫結果等，都應寫入 `host_calls.jsonl.gz` 的 `details` 欄位。

---

## 1. Build Tag 控制

```
-tags=trace       → 啟用 trace（實際記錄）
無 tag（預設）     → 所有方法為 no-op stub，零運行成本
```

相關檔案：
- `PVM/PVMtrace/trace_impl.go` — `//go:build trace`（實際實作）
- `PVM/PVMtrace/trace_stub.go` — `//go:build !trace`（no-op）
- `PVM/recompiler/invoke_mode_trace.go` — `//go:build linux && amd64 && trace`

---

## 2. 環境變數配置

| 環境變數 | 預設值 | 說明 |
|---------|--------|------|
| `JAM_PVM_TRACE_DIR` | （必填） | trace 輸出根目錄，未設定則不啟用 |
| `JAM_PVM_TRACE_RUN_ID` | "" | 可選子目錄名稱（例如 commit hash） |
| `JAM_PVM_TRACE_STREAMS` | "all" | 要記錄的 streams（"all" 或逗號分隔） |
| `JAM_PVM_TRACE_BUFFER_MB` | 4 | 每個 stream 的寫入 buffer 大小 |
| `JAM_PVM_TRACE_MAX_STEPS` | 0（無限） | 最大記錄步數，超過則 truncate |
| `JAM_PVM_TRACE_TOTAL_MB` | 0（無限） | 總輸出大小上限 |
| `JAM_PVM_TRACE_GZIP_LEVEL` | 6 | gzip 壓縮等級（1-9） |

---

## 3. 輸出目錄結構

```
$JAM_PVM_TRACE_DIR/
└── [$RUN_ID/]
    └── {serviceID}_{codeHash[:16]}/
        ├── meta/
        │   ├── info.json          ← 元資料（初始/最終狀態、步數、exit reason）
        │   └── program.bin        ← 原始 program blob（可離線 deblob）
        ├── pc.gz                  ← per-step PC (u64 LE, 8 bytes/record)
        ├── opcode.gz              ← per-step opcode (u8, 1 byte/record)
        ├── gas.gz                 ← per-step gas (i64 LE, 8 bytes/record)
        ├── dst_val.gz             ← per-step destination register value after execution (u64 LE)
        ├── src1_val.gz            ← per-step source1 register value before execution (u64 LE)
        ├── src2_val.gz            ← per-step source2 register value before execution (u64 LE)
        ├── loads.gz               ← per-step memory load (u32 addr + u64 val = 12 bytes/record)
        ├── stores.gz              ← per-step memory store (u32 addr + u64 val = 12 bytes/record)
        ├── host_calls.jsonl.gz    ← host-call boundary records (NDJSON)
        └── *.gz.sha256            ← 每個 stream 的 SHA-256 checksum
```

---

## 4. Stream 格式

### 4.1 Per-Instruction Streams（固定寬度 binary）

每個 step（= 一條 PVM 指令）寫入一筆 fixed-width record：

| Stream | Record Size | 內容 | 備註 |
|--------|------------|------|------|
| `pc.gz` | 8 bytes | uint64 LE | 當前指令的 PVM PC |
| `opcode.gz` | 1 byte | uint8 | opcode number (0-230) |
| `gas.gz` | 8 bytes | int64 LE | 執行**後**的 gas 值 |
| `dst_val.gz` | 8 bytes | uint64 LE | 目標 register 執行後的值（無 dst → 0） |
| `src1_val.gz` | 8 bytes | uint64 LE | 第一個 source register 執行前的值 |
| `src2_val.gz` | 8 bytes | uint64 LE | 第二個 source register 執行前的值 |
| `loads.gz` | 12 bytes | u32 addr + u64 val | 成功讀取的 guest memory（無 load → 全零） |
| `stores.gz` | 12 bytes | u32 addr + u64 val | 成功寫入的 guest memory（無 store → 全零） |

所有 stream 的 record 數量相同（= `TotalSteps`），可用 step index 對齊。

### 4.2 Host-Call Records（NDJSON）

`host_calls.jsonl.gz` 每行一個 JSON object：

```json
{
  "step": 12345,
  "pc": 1024,
  "op": 7,
  "op_name": "write",
  "regs_in": [0,0,0,...],
  "gas_in": 99000,
  "regs_out": [0,0,0,...],
  "gas_out": 98500,
  "exit_reason": "Continue",
  "details": {
    "memreads": [{"addr": "0x10000", "len": 32, "ok": true, "data": "0xabcd..."}],
    "memwrites": [{"addr": "0x20000", "len": 8, "ok": true}],
    "setregs": {"7": "0x42"},
    "setgas": 98500
  }
}
```

---

## 5. meta/info.json Schema

```json
{
  "format_version": 1,
  "graypaper_version": "0.7.2",
  "backend": "interpreter" | "recompiler",
  "trace_mode": "normal" | "debug-single-step",
  "service_id": 1000,
  "codehash": "0xabcdef...",
  "timeslot": 12345,
  "run_id": "abc123",
  "invocation_type": "accumulate",
  "initial_pc": 65536,
  "initial_gas": 1000000,
  "initial_regs": [4294901760, 4294836224, 0, ...],
  "final_pc": 65600,
  "final_gas": 999900,
  "final_exit_reason": "Halt",
  "final_regs": [0, 4294836224, 42, ...],
  "total_steps": 100,
  "truncated": false,
  "streams": ["pc.gz", "opcode.gz", "gas.gz", ...]
}
```

---

## 6. Recompiler Trace Mode — Debug Single-Step

當 `trace` build tag 啟用且 `Trace != nil` 時，recompiler 自動切換為
**Debug Single-Step** 模式（`DebugSingleStepInvoke`）：

```
MachineInvoke(pc):
    if Trace != nil → DebugSingleStepInvoke(pc)
    else            → BlockBasedInvoke(pc)
```

### 單步執行流程

```
for each instruction at pc:
    1. CompileSingleInstruction(instr) → 只編譯一條指令的 native code
    2. 記錄 src1_val, src2_val（執行前）
    3. ClearMemAccess()
    4. executeBlockLocked(block)        → 執行這一條指令
    5. 處理 sbrk / djump sentinel exits
    6. 記錄 dst_val（執行後）
    7. 讀取 MemAccess（load/store addr+val）
    8. trace.RecordStep(...)
    9. 根據 ExitReason → continue / return
```

### 為何需要 Single-Step？

Block-based 執行一次跑完整個 basic block，無法取得每條指令的中間狀態。
Single-step 犧牲效能，但產出的 trace 與 interpreter 完全 step-by-step 對齊。

### Memory Access 記錄

recompiler trace 使用 control region 的 `MemAccessAddr`（R15-160）和 `MemAccessVal`（R15-168）：
- emit 時在 load/store 指令後額外寫入 access address 到 control region
- 執行後 `HasMemAccess()` 檢查 → `traceRecordedMemAccess()` 讀取 guest memory 取值
- 只記錄成功的 access（與 interpreter 行為對齊）

---

## 7. Diff 工具（`PVM/PVMtrace/diff/`）

### FindFirstDivergence

比對兩份 trace（通常是 interpreter vs recompiler）的每個 step：

```go
result, err := diff.FindFirstDivergence(leftDir, rightDir)
```

比對優先順序：
1. PC + Opcode 是否相同
2. dst_val（destination register 值）
3. src1_val / src2_val
4. gas
5. loads（addr + val）
6. stores（addr + val）
7. trace 長度

### PrintDivergence

輸出人類可讀的分歧報告：

```
first divergence: step=42 stream=dst_val

step=42 pc=0x1a3 opcode=200 (add_64)
  left (interpreter):
    pc=0x1a3 opcode=200 gas=999958
    dst_val=0x000000000000002a src1_val=0x0000000000000015 src2_val=0x0000000000000015
  right (recompiler):
    pc=0x1a3 opcode=200 gas=999958
    dst_val=0x0000000000000029 src1_val=0x0000000000000015 src2_val=0x0000000000000015
```

### PrintStepTable

輸出前後多步的 tabular 比對：

```go
diff.PrintStepTable(os.Stdout, leftDir, rightDir, step-5, 10)
```

---

## 8. Deblob 靜態分析（`PVM/PVMtrace/deblob/`）

將 `meta/program.bin` 解碼並輸出結構化 metadata：

- `instr_meta.json.gz` — 每條指令的 PC、opcode、operands、所屬 block
- `blocks.json.gz` — 每個 basic block 的範圍和 gas cost

可用於離線分析指令分佈、hotspot 定位等。

---

## 9. 使用流程

### 9.1 產生 Interpreter Trace

```bash
cd JAM-Protocol
JAM_PVM_TRACE_DIR=/tmp/traces \
  go test -tags=trace -run TestXxx ./PVM/...
```

### 9.2 產生 Recompiler Trace

```bash
JAM_PVM_TRACE_DIR=/tmp/traces \
  go test -tags=trace -run TestXxx ./PVM/recompiler/...
```

### 9.3 比對兩份 Trace

```go
import "github.com/New-JAMneration/JAM-Protocol/PVM/PVMtrace/diff"

result, _ := diff.FindFirstDivergence(
    "/tmp/traces/1000_abcdef01/",  // interpreter
    "/tmp/traces/1000_abcdef01/",  // recompiler (different run_id)
)
diff.PrintDivergence(os.Stdout, result)
```

### 9.4 讀取單一 Trace

```go
import "github.com/New-JAMneration/JAM-Protocol/PVM/PVMtrace"

reader, _ := PVMtrace.OpenTrace("/tmp/traces/run1/1000_abcdef01/")
fmt.Println(reader.Info.TotalSteps, reader.Info.FinalExitReason)

pcReader, _ := reader.OpenStream(PVMtrace.StreamPC, PVMtrace.PCWidth)
defer pcReader.Close()
for {
    pc, err := pcReader.ReadU64()
    if err != nil { break }
    fmt.Printf("PC: 0x%x\n", pc)
}
```

---

## 10. 已知限制與未來改進

### 目前限制

| 限制 | 說明 |
|------|------|
| Single-step 效能差 | 每條指令獨立 compile + execute，比 block-based 慢 100x+ |
| Memory access 只記錄最後一次 | 一條指令多次 memory access（store_imm_ind 等）只記第一次 |
| Host-call details 不完整 | 某些 omega 的 memory read/write 未全部包裝 |
| 無 register dump 全集 | 只記錄 dst/src，不記錄所有 13 個 register（用 initial + delta 推導） |
| Diff 工具效能 | `readStepAt` 每次從頭掃描 gzip（O(n) per step），大量 random access 很慢 |

### 未來可改進

1. **Indexed access** — 加入 per-N-records 的 gzip block index，支援 O(1) random seek
2. **Register full dump stream** — 新增 `regs.gz`（13×8=104 bytes/record）stream
3. **Block-level trace mode** — 在 block-based 模式下記錄 block entry/exit 狀態（而非 per-instruction）
4. **Memory access 多筆** — `loads.gz` / `stores.gz` 支援 variable-length record
5. **CLI diff 工具** — 獨立 binary，支援 `pvmtrace diff <left> <right>` 命令列介面

---

## 11. 模組關係

```
PVMtrace/
├── trace.go         ← 常數、InitialState/FinalState、package doc
├── config.go        ← 環境變數讀取 → TraceConfig
├── info_schema.go   ← TraceInfo、HostCallRecord、MemAccess 結構
├── stream.go        ← stream 名稱常數、GzipRecordReader（讀取）
├── writer.go        ← streamWriter（gzip + SHA-256 sidecar）
├── trace_impl.go    ← Trace struct + RecordStep/RecordHostCall（trace build）
├── trace_stub.go    ← no-op stub（非 trace build）
├── reader.go        ← TraceReader: OpenTrace、ReadHostCalls
├── diff/
│   ├── compare.go   ← FindFirstDivergence、readStepAt
│   └── display.go   ← PrintDivergence、PrintStepTable
└── deblob/
    └── metadata.go  ← ProgramMetadata、WriteMetadata（靜態分析輸出）
```

Recompiler 側整合：
- `recompiler/invoke_mode_trace.go` — MachineInvoke 在 trace mode 使用 DebugSingleStepInvoke
- `recompiler/debug_single_step.go` — 單步執行 + trace 記錄邏輯
- `recompiler/trace_mem_trace.go` — traceRecordedMemAccess（讀取 control region memory access）
- `recompiler/host_call_trace.go` — wrapGuestMemoryForHostCallTrace（host-call memory 記錄）
