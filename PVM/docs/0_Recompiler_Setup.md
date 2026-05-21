# Recompiler Setup

平台限制：`//go:build linux && amd64`。

---

## 0.1 兩塊 mmap

Recompiler 需要兩塊獨立的 mmap memory space，各自服務不同目的：

```
Host 進程
├── Sandbox mmap（JITContext）── 存放 PVM 執行狀態 + guest RAM
│   [Control 4KB] [Guest 4GB] [Guard 4KB]
│
└── Executable mmap（ExecutableMemory，預設 16MB）── 存放 JIT native code
```

**為何分開**：guest memory 需要 `PROT_READ|PROT_WRITE`，JIT code 需要 `PROT_READ|PROT_EXEC`；W^X 政策下兩者不能共存於同一 mmap。

### Sandbox Memory（Generic Sandbox，單次 mmap）

```
┌──────────────────────────────────────────────────────────────────┐
│  unix.Mmap(4KB + 4GB + 4KB, PROT_NONE, MAP_NORESERVE)           │
│                                                                  │
│  ┌──────────────┬───────────────────────────────┬────────────┐   │
│  │ Control 4KB  │     Guest Memory 4GB          │ Guard 4KB  │   │
│  │  R/W         │  預設 PROT_NONE，按需 mprotect │ PROT_NONE  │   │
│  └──────────────┴───────────────────────────────┴────────────┘   │
│  ↑               ↑                                               │
│  mmapBase        R15 = mmapBase + 4096                           │
└──────────────────────────────────────────────────────────────────┘
```

- **初始**：整塊 `PROT_NONE` + `MAP_NORESERVE`（只佔虛擬位址空間，物理 RAM 按需分配）
- **Control region**：立刻 `mprotect` 為 `R/W`
- **Guest memory 4GB**：`InitFromProgram` / sbrk 時才按 segment / page 設權限
- **Guard page**：越界存取落在合法 mmap 範圍內，由 signal handler 處理（不觸發 kernel crash）

存取方式：
- **Native JIT**：`[R15 + PVM_addr]`（32-bit PVM 位址直接作為 offset）
- **Go host call**：`ctx.GuestMem()[addr : addr+size]`（同一塊 mmap）

### Executable Memory（第二塊 mmap，Dual Mapping）

```go
const DefaultExecutableSize = 16 * 1024 * 1024 // 16MB
```

> **概念背景**：`mprotect` 是 Linux syscall，用來動態改變虛擬記憶體頁面的權限（R/W/X 組合）。
> **W^X**（Write XOR Execute）是安全策略：同一塊頁面不能同時可寫可執行，防止攻擊者注入 shellcode。
> JIT 的困難在於：compile 時需要寫入機器碼（W），執行時需要跑它（X），但 W^X 不允許同時擁有兩者。
> 傳統做法是每次切換：寫完 → `mprotect(RX)`；要再寫 → `mprotect(RW)`。這很慢（syscall + TLB flush）。

使用 **dual mapping** 實現 W^X：同一塊實體頁面透過 `memfd_create` 建立兩個虛擬位址映射：

```
memfd_create("jit-code")
  ├── rwMem = mmap(fd, PROT_READ|PROT_WRITE, MAP_SHARED)  ← 寫入 native code
  └── rxMem = mmap(fd, PROT_READ|PROT_EXEC,  MAP_SHARED)  ← 執行 native code
```

- 寫入走 `rwMem`，執行走 `rxMem`（`rxBase = &rxMem[0]`）
- 同一實體頁，寫入後**立刻可執行**，不需要任何 `mprotect` 切換
- x86-64 硬體保證 instruction cache 與 data write 一致，不需 icache flush
- 存放：entry/exit trampoline、各 block native code、djump rodata

**安全性取捨**：頁面同時在不同 VA 可寫可執行（relaxed W^X），但 guest code 無法取得 `rwMem` 位址，sandbox 內安全。

**為何這麼做（歷史）**：舊版每次 compile 一個 block 都要 `MakeWritable → Write → MakeExecutable`（2× `mprotect` 整塊 16MB arena）。實測 mprotect 佔 compile 時間的 87%、全程 ~49%（QEMU 下 `PROT_EXEC` 觸發 TB-invalidation 更嚴重）。Dual mapping 將此成本降為零。

---

## 0.2 Control Region（R15 負 offset）

R15 指向 guest memory 起點；VM 執行狀態存在 R15 **之前** 的 4KB 區域。
常用欄位排最靠近 R15，確保 disp8（-128~+127）編碼效率。

| Offset (R15 - N) | 欄位 | 用途 |
|-------------------|------|------|
| 8 | ReturnStack | 存 host RSP，exit trampoline / signal handler 還原 |
| 16 | ReturnAddress | 存 host 返回位址 |
| 24 | HeapPointer | sbrk 維護的 heap 頂端 |
| 32 | ExitPC | JIT 退出時的 PVM PC |
| 40 | ExitReason | HALT / PANIC / OOG / HOST_CALL / PAGE_FAULT |
| 48 | Gas | 每條指令 inline `sub [R15-48], 1`（disp8 範圍內） |
| 49–152 | Registers[13] | 13 × 8 bytes = 104B，trampoline 邊界 save/restore |
| 160 | MemAccessAddr | debug trace 用 |
| 168 | MemAccessVal | debug trace 用 |
| 176/184/192 | Djump Table/Bitmask/Dispatch | djump native dispatch 用 |

**重點**：block 執行期間 PVM registers **全程在 x86 register**；只有 trampoline 邊界（進入/離開 native）才會與 control region 同步。

---

## 0.3 Register 對應（13 PVM → x86-64，零 spill）

| PVM idx | 名稱 | x86-64 | 備註 |
|---------|------|--------|------|
| 0 | RA | RAX | 高頻，非 REX |
| 1 | SP | RDX | |
| 2 | T0 | RBX | |
| 3 | T1 | RSI | |
| 4 | T2 | RDI | |
| 5 | S0 | R8 | |
| 6 | S1 | R9 | |
| 7 | A0 | R10 | |
| 8 | A1 | R11 | |
| 9 | A2 | R12 | |
| 10 | A3 | R13 | |
| 11 | A4 | R14 | |
| 12 | A5 | RBP | Generic Sandbox 設計的關鍵 — 釋放 RBP 給 A5 |

**保留（不映射 PVM register）**：

| x86 | 角色 |
|-----|------|
| R15 | Guest base + control region 定址 |
| RCX | Scratch（DIV/shift/位址計算/djump） |
| RSP | Native stack pointer |

**設計決策**：舊設計用 R15 = control region + RBP = memory base，只能映射 12 個 PVM register（A5 必須 spill）。新設計 R15 同時承擔 memory base（正向）和 control region（負向），釋放 RBP 給 A5，實現零 spill。

---

## 0.4 Guest Memory 邊界（InitFromProgram）

從 program blob 解出 segment 資料後，依 Graypaper 佈局 `mapSegment`：

| Segment | 位置（32-bit VA） | 權限 | 內容 |
|---------|-------------------|------|------|
| Read-only | `ZZ`(64KB) 起 | `PROT_READ` | code blob `o` |
| Read-write | `2*ZZ + Z(o)` 起 | `R/W` | init data `w` + heap pages |
| Stack | 高址往下 | `R/W` | 空 |
| Argument | `2^32 - ZZ - ZI` 起 | `PROT_READ` | invoke argument |

初始暫存器：RA=`0xFFFF0000`、SP=stack top、A0=argument 起始、A1=argument 長度。

另有 **Layer 1** 軟體權限表 `ctx.pages`（Go 端 `GuestMemory.IsReadable/IsWriteable`），避免 Go host call 直接碰 `PROT_NONE` page 觸發無法 recover 的 SIGSEGV。

---

## 0.5 JIT Entry/Exit Trampoline（一次性 lazy emit）

Trampoline 是 Go ↔ native 的橋：

**Entry trampoline**（Go → native）：
1. 保存 host callee-saved（RBX, RBP, R12–R15）
2. `R15 = guestBase`（從 RAX 傳入）
3. 寫 ReturnAddr / ReturnStack 到 control region
4. 從 control region 載入 13 個 PVM register 到對應 x86 register
5. `JMP` block native code

**Exit trampoline**（native → Go，共用）：
1. 13 個 PVM register 寫回 control region
2. 還原 RSP（ReturnStack）
3. `JMP` ReturnAddr → 回到 entry 的 `return_label` → `RET` 到 Go

觸發 exit 的情境：`ecalli` / HALT / OOG / PANIC / PAGE_FAULT。

```
Go callNative() → CALL entry_trampoline
  → [native block] → JMP exit_trampoline
  → return_label → RET → Go
```

---

## 相關檔案索引

| 檔案 | 職責 |
|------|------|
| `context.go` | JITContext：sandbox mmap + control region accessor |
| `guest_memory.go` | InitFromProgram + mapSegment + GuestMemory 實作 |
| `executable.go` | ExecutableMemory：dual-mapping arena（memfd + RW/RX views） |
| `register_map.go` | PVM → x86 靜態映射表 |
| `trampoline.go` | Entry/Exit trampoline emit |
| `call_native_amd64.s` | Go → trampoline 的 assembly bridge |
