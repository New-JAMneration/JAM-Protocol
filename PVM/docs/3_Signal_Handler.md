# Signal Handler & 硬體記憶體保護

本章說明 recompiler 如何利用 OS signal 機制，將硬體層面的記憶體違規轉換為 PVM 語意的 ExitReason。

前置閱讀：`0_Recompiler_Setup.md`（Guest Memory 與 Control Region）、`1_Recompiler_Workflow.md`（Execute 階段）。

---

## 1. 為什麼需要 Signal Handler

PVM sandbox 的 guest memory 是 4GB 虛擬位址空間，初始大部分是 `PROT_NONE`（不可讀不可寫）。當 native code 存取一個非法位址時：

1. CPU 硬體偵測到 page fault
2. OS kernel 發送 `SIGSEGV`（或 `SIGBUS`）到 process
3. Signal handler 攔截，判斷是否為「合法的 JIT fault」
4. 如果是 → 修改 ucontext 讓執行回到 Go → 回報 ExitPageFault 或 ExitPanic
5. 如果不是 → chain to previous handler（或 crash）

這種設計的優勢：
- **零成本 bounds check**：不需要在每個 load/store 前插入比較指令
- **硬體速度**：正常存取完全不受影響（no branch overhead）
- **最小 code size**：相比 software bounds check 少了 3-5 條 x86 指令 / memory op

---

## 2. 背景知識：RIP、RSP 與 Signal

### 2.1 RIP 與 RSP 是什麼

CPU 有兩個關鍵暫存器決定「程式現在在做什麼」：

- **RIP**（Instruction Pointer）：指向**下一條要執行的機器碼位址**。改 RIP = 改執行流程（跳到別處）。
- **RSP**（Stack Pointer）：指向**當前 call stack 的頂端**。函式的區域變數、return address 都在 stack 上。每個 thread 有自己的 stack。

簡單說：RIP 決定「跑哪段 code」，RSP 決定「用誰的 stack」。兩者合在一起 = 完整的執行 context。

### 2.2 進入 JIT 後的 RIP/RSP 變化

```
Go 側（callNative 前）         Native 側（JIT code 執行中）
─────────────────────          ──────────────────────────
RIP = Go code 的某行            RIP = JIT emit 的 native 指令
RSP = Go goroutine stack       RSP = 同一個 stack（但已 PUSH 了 callee-saved）
```

進入 JIT 前，entry trampoline 會把「回家的路」存進 control region：
- **ReturnAddr**：`callNative` 的 return address（Go 側接著要跑的那行）
- **ReturnStack**：當時的 RSP 值（Go 的 stack 位置）

這兩個值在正常 exit（exit trampoline）時用來回到 Go。Signal handler 也需要用到它們。

### 2.3 Signal 是什麼

當 CPU 遇到無法處理的狀況（page fault、除以零等），它不會 crash，而是把控制權交給 OS kernel。Kernel 會：

1. 暫停該 thread
2. 保存當時所有 register 到一個結構（`ucontext_t`）
3. 呼叫 process 註冊的 signal handler，把 `ucontext_t` 傳進去
4. Handler 可以**修改** `ucontext_t` 中的 register 值
5. Handler return 後，kernel 用（可能被修改過的）`ucontext_t` 恢復 CPU 執行

這代表 signal handler 可以改 RIP 和 RSP → 控制 CPU 接下來跳到哪裡、用誰的 stack。

### 2.4 為什麼 Signal Handler 要改 RIP/RSP

問題情境：native code 碰到 `PROT_NONE` 頁面 → SIGSEGV。此時：
- RIP 指向觸發 fault 的那條 JIT 指令（**不能繼續**）
- RSP 在 native side

我們要「回到 Go」，但不能讓 CPU 繼續跑 fault 那條指令。解法：

```
修改 ucontext：
  RIP → ReturnAddr（Go 側 callNative 之後的那行）
  RSP → ReturnStack（Go 的 stack）
```

Handler return 後，kernel 用修改後的值恢復 CPU → CPU 現在跑在 Go code 上、用 Go stack → **等於正常從 callNative return 了**。

Go 側只需 `ReadExitReason()` 就能拿到結果，完全不知道是 signal 觸發的。

Signal handler 本質上是**偽造了一次 exit trampoline return**。

---

## 3. 雙層記憶體保護模型

### Layer 1：Go-side Permission Table（soft check）

```go
// context.go
pages map[uint32]pageAccess  // pageInaccessible / pageReadOnly / pageReadWrite
```

用途：host call（`omega`）和 Go-side 的 memory access 必須先查 Layer 1。
原因：Go 無法用 `recover()` 捕獲 SIGSEGV（只有 Go panic 可以 recover）。

### Layer 2：mprotect（hardware enforcement）

```go
unix.Mprotect(slice, unix.PROT_READ|unix.PROT_WRITE)  // 啟用頁面
unix.Mprotect(slice, unix.PROT_NONE)                    // 關閉頁面
```

用途：native code 的 memory access 直接走硬體 MMU 檢查。
如果碰到 `PROT_NONE` 頁面 → 硬體觸發 SIGSEGV → signal handler 接手。

### 兩層的關係

```
Layer 1 == Layer 2（兩者完全同步）
```

- `mapSegment()` 和 `SetPageAccess()` 每次呼叫 `mprotect` 後都同步更新 `ctx.pages`
- 因此 Layer 1 與 Layer 2 的狀態**永遠一致**
- Native code 走 Layer 2（硬體 MMU 強制）
- Go code 走 Layer 1（查 `ctx.pages` map），因為 Go 無法 recover SIGSEGV
- 分成兩層的目的不是「允許不一致」，而是「相同的保護語意，不同的檢查機制」

---

## 4. 處理的 Signal 種類

| Signal | 觸發原因 | PVM 語意 |
|--------|---------|---------|
| `SIGSEGV` | 存取 PROT_NONE 頁面 | ExitPageFault 或 ExitPanic |
| `SIGBUS` | 非法記憶體對齊 / mmap 超出 | 同 SIGSEGV 邏輯 |
| `SIGFPE` | 除以零（`IDIV` / `DIV`）| ExitPanic |

三者共用同一個 handler function，以 `sig` 參數區分邏輯。

---

## 5. Signal Handler 實作（C 層）

位於 `PVM/recompiler/x86signal/x86_signal_linux.c`。

### 4.1 為什麼用 C？

- `sigaction` + `ucontext_t` 的操作需要 C-level 的結構存取
- Go 的 signal handling 有自己的 runtime 限制，無法直接修改 `ucontext`
- TLS（Thread-Local Storage）用 `__thread` 變數實現 per-goroutine state

### 4.2 TLS 變數

```c
static __thread uint8_t *guest_base_ptr;   // 當前 goroutine 的 R15 值
static __thread uintptr_t jit_code_start;  // ExecutableMemory rxMem 起始
static __thread uintptr_t jit_code_end;    // ExecutableMemory rxMem 結束
```

每次進入 native code 前，Go 側呼叫 `SetFaultWindow` 設定這三個值。

### 4.3 Handler 安裝

```c
void setup_signal_handler(void) {
    struct sigaction sa;
    sa.sa_sigaction = signal_handler;
    sa.sa_flags = SA_SIGINFO | SA_NODEFER | SA_ONSTACK;

    sigaction(SIGSEGV, &sa, &old_sigsegv);
    sigaction(SIGBUS,  &sa, &old_sigbus);
    sigaction(SIGFPE,  &sa, &old_sigfpe);
}
```

| Flag | 意義 |
|------|------|
| `SA_SIGINFO` | 使用 3-argument handler（帶 `siginfo_t` + `ucontext_t`） |
| `SA_NODEFER` | handler 執行期間不阻擋同類 signal |
| `SA_ONSTACK` | 使用 alternate signal stack（避免 stack overflow 時無法 handle） |

舊 handler 保存在 `old_sigsegv` / `old_sigbus` / `old_sigfpe`，用於 chain。

### 4.4 Handler 主體邏輯

```c
static void signal_handler(int sig, siginfo_t *info, void *uctx_void) {
    ucontext_t *uc = (ucontext_t *)uctx_void;
    greg_t *gregs = uc->uc_mcontext.gregs;
    uint8_t *base = guest_base_ptr;           // TLS: 當前 R15

    // --- SIGFPE: 除以零 ---
    if (sig == SIGFPE) {
        if (!is_jit_code(RIP) || base == NULL)
            chain_to_old(...);                 // 不是 JIT code → 交還
        jit_exit_panic(base, gregs);           // → ExitPanic
        return;
    }

    // --- SIGSEGV / SIGBUS ---
    uintptr_t fault_addr = info->si_addr;
    if (!is_jit_fault(base, fault_addr))
        chain_to_old(...);                     // 不在 guest 4GB 範圍 → 交還

    uint32_t guest_addr = fault_addr - base;   // 算出 PVM guest address
    if (guest_addr < 0x10000u)
        jit_exit_panic(base, gregs);           // 低 64KB = 保護區 → Panic
    else {
        store_pvm_regs(base, gregs);           // 回存所有 PVM regs
        ExitReason = (PAGE_FAULT << 56) | guest_addr;
        // 修改 RIP/RSP → return to Go
    }
}
```

### 4.5 Fault 判定流程圖

```
Signal 到達
    │
    ├─ SIGFPE?
    │   ├─ RIP 在 JIT code 範圍內? → ExitPanic
    │   └─ 否 → chain_to_old
    │
    ├─ fault_addr 在 [guest_base, guest_base + 4GB + 4KB) 內?
    │   ├─ 否 → chain_to_old（不是我們的 fault）
    │   │
    │   ├─ guest_addr < 0x10000（低 64KB 保護區）?
    │   │   └─ ExitPanic
    │   │
    │   └─ 否 → ExitPageFault(guest_addr)
    │
    └─ chain_to_old
```

### 4.6 `is_jit_fault` — 判斷是否為 guest memory fault

```c
static int is_jit_fault(uint8_t *base, uintptr_t fault_addr) {
    if (base == NULL) return 0;
    uintptr_t lo = (uintptr_t)base;
    uintptr_t hi = lo + (4ULL * 1024 * 1024 * 1024) + 4096;  // 4GB + guard
    return fault_addr >= lo && fault_addr < hi;
}
```

### 4.7 `is_jit_code` — 判斷 RIP 是否在 JIT code 內

```c
static int is_jit_code(uintptr_t rip) {
    return jit_code_start != 0 && rip >= jit_code_start && rip < jit_code_end;
}
```

用於 SIGFPE：只有 JIT emit 的 `IDIV` 才走 ExitPanic，其他 DIV-by-zero 交還原 handler。

---

## 6. 從 Signal Handler 回到 Go

### 5.1 `jit_exit_panic`（ExitPanic 路徑）

```c
static void jit_exit_panic(uint8_t *base, greg_t *gregs) {
    store_pvm_regs(base, gregs);                              // 1. 回存 13 個 PVM regs
    *(uint64_t *)(base - OFF_EXIT_REASON) = PANIC << 56;      // 2. 寫 ExitReason
    gregs[REG_RIP] = *(uintptr_t *)(base - OFF_RETURN_ADDR);  // 3. 改 RIP → Go return
    gregs[REG_RSP] = *(uintptr_t *)(base - OFF_RETURN_STACK); // 4. 改 RSP → Go stack
}
```

### 5.2 ExitPageFault 路徑

同上，但 ExitReason 帶 fault address payload：

```c
uint64_t exit_reason = (EXIT_PAGE_FAULT_TYPE << 56) | (uint64_t)guest_addr;
```

### 5.3 回到 Go 的機制

Signal handler 直接修改 `ucontext_t` 中的 `RIP` 和 `RSP`：

```
修改前：RIP = 觸發 fault 的 JIT native code 位址
修改後：RIP = ctx.ReturnAddr（entry trampoline 的 restore 段）
        RSP = ctx.ReturnStack（Go 的 native stack）
```

當 handler `return` 後，kernel 用修改後的 ucontext 恢復執行 → 落在 Go 的 stack 上 → 回到 `callNative` 之後 → 讀取 ExitReason。

### 5.4 Register 回存（`store_pvm_regs`）

```c
static const int pvm_to_greg[13] = {
    REG_RAX, REG_RDX, REG_RBX, REG_RSI, REG_RDI,
    REG_R8, REG_R9, REG_R10, REG_R11, REG_R12, REG_R13, REG_R14, REG_RBP
};

static void store_pvm_regs(uint8_t *base, greg_t *gregs) {
    for (int i = 0; i < 13; i++) {
        uint64_t *slot = (uint64_t *)(base - OFF_REGISTERS + pvm_reg_slot[i] * 8);
        *slot = (uint64_t)gregs[pvm_to_greg[i]];
    }
}
```

因為 fault 發生時 register 狀態是在 CPU 中（ucontext.gregs），必須透過 signal handler 回寫到 control region。
正常 exit（OOG、host call）則是由 exit trampoline 自己回存（不經過 signal handler）。

---

## 7. Go 側整合

### 6.1 為什麼需要 SetFaultWindow

Signal handler 是 **process 全域** 的——整個 process 裡不管哪個 thread 收到 SIGSEGV / SIGFPE，都會跑同一個 handler。但不是所有 fault 都跟我們的 JIT 有關：

- Go runtime 自己也可能 SIGSEGV（例如 nil pointer、stack growth probe）
- Go 的整數除法也可能觸發 SIGFPE
- 其他 cgo library 也可能有 memory fault

如果 signal handler **無條件**攔截所有 signal → 會把不屬於 JIT 的 fault 也吞掉 → Go runtime 行為被破壞。

**SetFaultWindow 的作用**：告訴 signal handler「現在這個 thread 正在跑 JIT，只有符合以下條件的 fault 才是我的」：

| TLS 變數 | 判斷用途 |
|----------|---------|
| `guest_base_ptr` | fault 地址是否落在 `[base, base+4GB+4KB)` → 是 guest memory fault？ |
| `jit_code_start` / `jit_code_end` | RIP 是否在 JIT emit 的 code 範圍內 → 是 JIT 觸發的 SIGFPE？ |

**ClearFaultWindow 的作用**：離開 JIT 後清除 TLS → 同一個 thread 之後收到的 signal 不再被攔截 → chain to old handler。

```go
// execute.go — executeBlockLocked
func executeBlockLocked(ctx *JITContext, block *CompiledBlock) PVM.ExitReason {
    x86_signal_linux.SetFaultWindow(ctx.GuestBase(), codeStart, codeEnd)
    defer x86_signal_linux.ClearFaultWindow()

    callNative(ctx.GuestBase(), block.NativeAddr, trampolineAddr)
    return ctx.ReadExitReason()
}
```

時序上形成一個「窗口」：

```
SetFaultWindow    ──── 窗口打開：handler 會攔截 JIT fault ────    ClearFaultWindow
                       ↑ callNative 在這段期間執行 ↑
```

窗口外的 signal → `guest_base_ptr == NULL` → handler 立刻 chain to old → 不影響 Go runtime。

### 6.2 LockOSThread

```go
runtime.LockOSThread()
defer runtime.UnlockOSThread()
```

**必須** 在 `SetFaultWindow` 之前鎖定 OS thread，因為：
- TLS 是 per-OS-thread 的
- Go scheduler 可能把 goroutine 搬到另一個 thread
- 如果搬了，TLS 就失效 → signal handler 拿不到正確的 `guest_base_ptr`

### 6.3 Handler 安裝時機

```go
// execute.go
func ExecuteBlock(ctx *JITContext, block *CompiledBlock) PVM.ExitReason {
    x86_signal_linux.SetupSignalHandler()  // sync.Once: 只裝一次
    runtime.LockOSThread()
    ...
}
```

用 `sync.Once` 保證整個 process 只安裝一次（多 goroutine 安全）。

---

## 8. Fault Window vs. 舊版 SetVmCtx

| | SetVmCtx（舊） | SetFaultWindow（現） |
|---|---|---|
| 設定 | 只設 `guest_base_ptr` | 設 `guest_base_ptr` + `jit_code_start` + `jit_code_end` |
| SIGFPE 判斷 | 無法判斷 RIP 是否在 JIT code 內 | 可精確判斷 |
| 誤判風險 | Go runtime 的 DIV 可能被吃掉 | 只攔 JIT 內的 DIV |

`SetVmCtx` 仍保留作為 API，但目前 `executeBlockLocked` 只使用 `SetFaultWindow`。

---

## 9. 特殊情況

### 8.1 低 64KB 保護區（guest_addr < 0x10000）

PVM 的 address 0x0000–0xFFFF 是 null pointer trap zone。碰到 → 直接 ExitPanic（不是 PageFault）。

### 8.2 Guard Page（4GB + 4KB）

在 guest memory 末尾的 4KB guard page 也是 PROT_NONE。碰到 → `is_jit_fault` 仍回 true（因為 hi 包含 guard）→ guest_addr > 4GB → 超出 32-bit 範圍 → 實際上 fault_addr 會落在合法判定中被當作 ExitPageFault。

### 8.3 Chain to Old Handler

```c
static void chain_to_old(int sig, siginfo_t *info, void *uctx) {
    struct sigaction *old = ...;
    if (old->sa_flags & SA_SIGINFO)
        old->sa_sigaction(sig, info, uctx);  // 呼叫舊 handler
    else if (old->sa_handler != SIG_DFL)
        old->sa_handler(sig);
    else {
        // 恢復 SIG_DFL → 下次同樣 signal 會 kill process
        sa.sa_handler = SIG_DFL;
        sigaction(sig, &sa, NULL);
    }
}
```

這確保 Go runtime 自己的 signal handler（用於 GC、stack growth）不被覆蓋。

### 8.4 非 Linux 平台（stub）

```go
//go:build !(linux && amd64 && cgo)

func SetupSignalHandler() {}
func SetFaultWindow(guestBase, codeStart, codeEnd uintptr) {}
func ClearFaultWindow() {}
```

非 linux/amd64 → no-op。recompiler 目前只支援 Linux x86-64。

---

## 10. 完整流程時序圖

```
Go (executeBlockLocked)          Native (JIT code)                Signal Handler (C)
─────────────────────────────────────────────────────────────────────────────────────
LockOSThread()
SetFaultWindow(base, start, end)
callNative(base, nativeAddr, trampoline)
   │
   │    ──── entry trampoline ────►
   │                                 MOV RCX, [R15+guest_addr]
   │                                 (guest_addr is PROT_NONE)
   │                                        │
   │                                        │ ──── SIGSEGV ────►
   │                                        │                     is_jit_fault? ✓
   │                                        │                     guest_addr > 0x10000? ✓
   │                                        │                     store_pvm_regs(gregs)
   │                                        │                     ExitReason = PageFault|addr
   │                                        │                     RIP = ReturnAddr
   │                                        │                     RSP = ReturnStack
   │                                        │  ◄── handler return ──
   │    ◄── (kernel restores ucontext) ─────
   │
ReadExitReason() → ExitPageFault(addr)
ClearFaultWindow()
UnlockOSThread()
return to BlockBasedInvoke loop
```

---

## 相關檔案索引

| 檔案 | 職責 |
|------|------|
| `PVM/recompiler/x86signal/x86_signal_linux.c` | Signal handler 核心實作（C） |
| `PVM/recompiler/x86signal/x86_signal_linux.h` | 常數定義 + 函式宣告 |
| `PVM/recompiler/x86signal/x86_signal.go` | Go wrapper（CGO binding） |
| `PVM/recompiler/x86signal/x86_signal_stub.go` | 非 linux/amd64 的 no-op stub |
| `PVM/recompiler/execute.go` | `ExecuteBlock` / `SetFaultWindow` 呼叫點 |
| `PVM/recompiler/context.go` | Control region offset 定義（header 同步） |
| `PVM/recompiler/trampoline.go` | ReturnAddr / ReturnStack 寫入 |
