# Host Call 整合

本章說明 JIT recompiler 如何處理 PVM 的 `ecalli` 指令（host call），從 native code exit 到 Go dispatch omega 再回到 native 的完整流程。

前置閱讀：`1_Recompiler_Workflow.md`（Execute 階段的 ExitReason 分支）。

---

## 1. Host Call 在 PVM 中的角色

PVM 是純計算引擎，不能直接存取外部狀態（鏈上資料、service storage 等）。當 guest code 需要跟外界互動時，執行 `ecalli N` 指令，N 是 host call 的 operation ID。

Graypaper 把這些外部操作稱為 **Omega（Ω）**——每種 operation（read、write、lookup、info 等）對應一個 omega function，定義了輸入/輸出語意和 gas 計費。

---

## 2. 雙層架構：host vs Recompiler

```
host（host-call 派發層）
 │
 ├── host.HostCall(pc)           ← 外層 loop
 │     │
 │     ├── recomp.MachineInvoke(pc)  ← 進入 native 執行
 │     │     └── BlockBasedInvoke     ← block 迴圈
 │     │           └── executeBlockLocked → callNative
 │     │
 │     ├── 收到 ExitReason == HOST_CALL
 │     │     └── snapshot regs/gas
 │     │     └── omega(input) → result
 │     │     └── writeback regs/gas
 │     │     └── continue loop
 │     │
 │     └── 收到 HALT/PANIC/OOG/PAGE_FAULT → return
 │
Recompiler（機器層）
```

**類比**：
- `host` ≈ interpreter 的 `PVM.Host`（host-call dispatch）
- `Recompiler` ≈ interpreter 的 `PVM.Interpreter`（machine execution）

---

## 3. ecalli 的 Native Code Emit

`ecalli` 是 block terminator。Compiler emit 的 native code 很簡單：

```go
// emit_basic.go — emitEcalli
func (c *Compiler) emitEcalli(a *asm.Assembler, instr *PVM.InstrMeta) error {
    callID := int(instr.Imm[0])
    nextPC := fallthroughPC(instr)
    exitReason := ExitHostCall | ExitReason(callID)

    MOV RCX, exitReason        // ExitReason = (HOST_CALL << 56) | callID
    MOV [R15-40], RCX          // 寫入 control region
    MOV dword [R15-32], nextPC // ExitPC = ecalli 的下一條指令
    JMP exit_trampoline        // 回到 Go
}
```

Exit trampoline 會回存 13 個 PVM registers → 恢復 Go stack → return。

Go 側 `ReadExitReason()` 拿到 `HOST_CALL | callID`，知道要呼叫哪個 omega。

---

## 4. Go 側 Host Call Dispatch（host.HostCall）

```go
func (h *host) HostCall(pc ProgramCounter) Psi_H_ReturnType {
    for {
        // 1. 進入 native 執行
        exitReason, pcPrime := h.recomp.MachineInvoke(pc)

        // 2. 終止條件直接 return
        switch exitReason.GetReasonType() {
        case HALT, PANIC, OUT_OF_GAS, PAGE_FAULT:
            snapshot()
            return Psi_H_ReturnType{...}
        }

        // 3. HOST_CALL → dispatch omega
        snapshot()  // 讀 regs + gas 到 VMState

        input := OmegaInput{
            Operation: OperationType(exitReason.GetHostCallID()),
            VM:        &vm,
            Addition:  h.Addition,
            HostCalls: h.HostCalls,
        }

        omega := GetOmega(h.HostCalls, input.Operation)
        result := omega(input)

        // 4. Writeback → 繼續
        ctx.WriteRegisters(regs)
        ctx.WriteGas(gas)
        pc = pcPrime    // ecalli 的 fallthrough PC
        continue
    }
}
```

### 4.1 Snapshot（Native → Go）

```go
ctx.ReadRegistersInto(regsBuf)    // control region → Go struct
ctx.ReadGasInto(gasBuf)
vm.Mem = ctx.GuestMemory()        // GuestMemory interface（直接看 mmap）
```

### 4.2 Omega 執行

Omega function 透過 `GuestMemory` interface 讀寫 guest memory：

```go
type GuestMemory interface {
    IsReadable(addr, length uint64) bool
    IsWriteable(addr, length uint64) bool
    Read(addr, length uint64) []byte
    Write(addr uint64, data []byte)
}
```

- `IsReadable` / `IsWriteable` 查 `ctx.pages` map（Layer 1）
- `Read` / `Write` 直接操作 `guestMem[addr:addr+length]`（同一塊 mmap）

因為 omega 在 Go 中執行，不能直接碰 PROT_NONE 頁面（Go 接不住 SIGSEGV），所以先查 Layer 1 再存取。

### 4.3 Writeback（Go → Native）

```go
ctx.WriteRegisters(regs)   // Go struct → control region
ctx.WriteGas(gas)
```

**Guest memory 不需要 writeback**：omega 直接透過 `guestMem` slice 寫入 mmap，native code 讀的是同一塊實體記憶體。零複製。

---

## 5. MachineInvoke 與 block 執行引擎

`MachineInvoke` 是 **Ψ_H 外層 loop 的統一入口**：從 `pc` 跑到 HALT / host call / OOG / panic 等非 CONTINUE 為止。  
本身幾乎不做事，只做 **build-tag 分流**（trace vs production）並轉呼叫底下的 block 引擎。

### 路由（依 backend）

兩個 **outer Ψ_H** backend 都跑 **pre-decoded** basic blocks（`deblob` → `preDecodeBlocks` 產生的 `Instrs` / `BlockMeta` / `GasCost`）。  
函式名稱不同，語意對稱：

| Backend | 檔案 | Production | 執行方式 | Trace |
|---------|------|------------|----------|-------|
| Interpreter | `interpreter/invoke_mode.go` | `BlockBasedInvokeDecodedBlocks` | Go 直譯 pre-decoded blocks | `DebugSingleStepInvoke` |
| Recompiler | `recompiler/invoke_mode.go` | `BlockBasedInvoke` | native JIT 編譯 + 執行 **同一套** pre-decoded blocks | `DebugSingleStepInvoke` |

Recompiler 的 `lookupOrCompileBlock` → `CompileBasicBlock` 讀 `program.BlockContaining(pc)` 與 `Instrs[InstrStart:InstrEnd]`，**不是** interpreter 的 runtime `DecodeInstructionBlock`。

Refine inner VM（Ω_K `invoke`）同樣走 **`BlockBasedInvokeDecodedBlocks`**（interpreter only；無 recompiler / 無 `MachineInvoke` 包裝）。`machine` 註冊時預 decode 存 `IntegratedPVMType.Program`；`invoke` 僅重驗 `𝔳_inst`。

### 呼叫點

| 呼叫者 | 檔案 | 被叫 |
|--------|------|------|
| Ψ_H 外層 loop | `interpreter/host.go` | `h.MachineInvoke(pc)` |
| Ψ_H 外層 loop | `recompiler/host.go` | `h.recomp.MachineInvoke(pc)` |
| refine inner VM（Ω_K invoke） | `host_call_refine.go` | `tempInterp.BlockBasedInvokeDecodedBlocks` |

Inner 不經 `MachineInvoke` / recompiler：每次 `invoke` 建立 ephemeral `tempInterp`，gas 由 outer `M_K + g_R` 帳務退還。

### Recompiler：`BlockBasedInvoke` 內部

```
host.HostCall
  └── MachineInvoke(pc)
        └── BlockBasedInvoke(pc)
              for {
                  block = lookupOrCompileBlock(pc)
                  executeBlockLocked(block)
                  switch exitReason:
                    CONTINUE → 下一 block
                    djump miss → 內部消化，不出 MachineInvoke
                    其他     → 回傳 host（HOST_CALL / HALT / …）
              }
```

**內部消化（不上報 host）**：`CONTINUE` fallthrough、`djump miss`（0xFE）。

**上報 host**：`HOST_CALL`（含 `grow_heap`，omega ID=1）、`HALT`、`PANIC`、`OOG`、`PAGE_FAULT`。

### 為何保留 `MachineInvoke` 這層？

- Trace / production 分流在 `invoke_mode*.go`，不污染 `host.go`
- 兩 backend 共用同一呼叫慣例
- JIT profile 以 `MachineInvoke` 計 `lockCalls`

刪掉改直接呼叫 block 引擎幾乎無效能收益，還會打散 build-tag 結構。

---

## 6. 完整時序圖

```
host.HostCall                MachineInvoke/BlockBased         Native Code
────────────────────────────────────────────────────────────────────────────
MachineInvoke(pc=0) ──►
                              lookupOrCompileBlock(0)
                              executeBlockLocked(block0)  ──►
                                                              block0: 算術...
                                                              ecalli 5
                                                              MOV ExitReason = HOST_CALL|5
                                                              JMP exit_trampoline
                                                          ◄── return ExitReason
                         ◄── return (HOST_CALL|5, pc=next)

snapshot regs/gas
omega = GetOmega(5)           // e.g. "info"
result = omega(input)
  └── vm.Mem.Read(...)        // 直接讀 mmap
  └── vm.Mem.Write(...)       // 直接寫 mmap
WriteRegisters(result.regs)
WriteGas(result.gas)

MachineInvoke(pc=next) ──►
                              lookupOrCompileBlock(next)
                              executeBlockLocked(blockN) ──►
                                                              blockN: ...
                                                              HALT
                                                          ◄── return ExitReason
                         ◄── return (HALT, pc=final)
◄── return Psi_H_ReturnType{HALT, ...}
```

---

## 7. ExitReason 編碼

```
ExitReason = uint64
  bits 63..56 = ExitType (1 byte)
  bits 55..0  = payload (7 bytes)

HOST_CALL: type=0x05, payload=callID (omega operation ID)
```

| ExitType | 值 | payload 意義 |
|----------|---|------------|
| HALT | 0x01 | 無 |
| PANIC | 0x02 | 無 |
| OOG | 0x03 | 無 |
| PAGE_FAULT | 0x04 | fault guest address |
| HOST_CALL | 0x05 | omega operation ID |

特殊 sentinel（不出 MachineInvoke）：

| sentinel | callID | 用途 |
|----------|--------|------|
| djump | 0xFE | indirect jump resolve |

---

## 相關檔案索引

| 檔案 | 職責 |
|------|------|
| `PVM/interpreter/host.go` | interpreter Ψ_H 外層 loop → `MachineInvoke` |
| `PVM/interpreter/invoke_mode.go` | `MachineInvoke` → `BlockBasedInvokeDecodedBlocks` |
| `PVM/interpreter/invoke_mode_trace.go` | trace 分流 → `DebugSingleStepInvoke` |
| `PVM/recompiler/host.go` | recompiler Ψ_H 外層 loop → `MachineInvoke` |
| `PVM/recompiler/recompiler.go` | `BlockBasedInvoke`（pre-decoded blocks → native JIT） |
| `PVM/recompiler/invoke_mode.go` | `MachineInvoke` → `BlockBasedInvoke` |
| `PVM/recompiler/invoke_mode_trace.go` | trace 分流 |
| `PVM/invocation.go` | `BlockBasedInvoke*` / `DebugSingleStepInvoke` 實作 |
| `PVM/host_call_refine.go` | Ω_K inner VM → `BlockBasedInvoke`（非 MachineInvoke；#15） |
