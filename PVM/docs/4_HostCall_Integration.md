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

## 5. MachineInvoke 與 BlockBasedInvoke 的關係

```
host.HostCall (外層 loop)
  │
  └── MachineInvoke(pc)
        │
        └── BlockBasedInvoke(pc)
              │
              for {
                  block = lookupOrCompileBlock(pc)
                  executeBlockLocked(block)
                  switch exitReason:
                    CONTINUE → pc = exitPC, continue
                    sbrk     → resolveSbrk, continue
                    djump    → resolveDjump, continue
                    其他     → return (HOST_CALL / HALT / ...)
              }
```

**BlockBasedInvoke 內部消化的 exit**：
- `CONTINUE`：block fallthrough，繼續下一個 block
- `sbrk`（`0xFF`）：HandleSbrk（Go mprotect），不出 MachineInvoke
- `djump miss`（`0xFE`）：compile target，不出 MachineInvoke

**上報給 host 的 exit**：
- `HOST_CALL`（真正的 ecalli）：需要 omega dispatch
- `HALT` / `PANIC` / `OOG` / `PAGE_FAULT`：程式結束

---

## 6. sbrk 的特殊處理

sbrk 在語意上也是「回到 Go 做事」，但它**不是真的 host call**——它不走 omega dispatch，而是在 `BlockBasedInvoke` 內部用特殊的 `SbrkCallID = 0xFF` 標記。

為什麼不走 omega：
- sbrk 需要 `mprotect`（kernel syscall），只有 Go 能安全呼叫
- 但它不需要讀寫 service state，不需要 `Addition` / `HostCalls`
- 處理完就能繼續跑，不需要離開 `MachineInvoke` 的 LockOSThread 區間

```go
// recompiler.go — BlockBasedInvoke
if IsSbrkExit(exitReason) {
    exitReason, pc = r.resolveSbrk(instr)
    continue  // 不出 MachineInvoke
}
```

---

## 7. 完整時序圖

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

## 8. ExitReason 編碼

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
| sbrk | 0xFF | HandleSbrk（mprotect） |
| djump | 0xFE | indirect jump resolve |

---

## 相關檔案索引

| 檔案 | 職責 |
|------|------|
| `PVM/recompiler/host.go` | host-call dispatch 層（外層 loop + omega 呼叫） |
| `PVM/recompiler/recompiler.go` | `BlockBasedInvoke`（inner loop、sbrk/djump resolve） |
| `PVM/recompiler/invoke_mode.go` | `MachineInvoke` → `BlockBasedInvoke` routing |
| `PVM/recompiler/emit_basic.go` | `emitEcalli`（native code emit） |
| `PVM/recompiler/execute.go` | `HandleSbrk`、`SbrkCallID`、`DjumpCallID` |
| `PVM/recompiler/guest_memory.go` | `GuestMemory` interface 實作（Layer 1 check） |
| `PVM/recompiler/trampoline.go` | exit trampoline（回存 regs → return Go） |
