# PVM Recompiler Architecture（AI Agent Rebuild Reference）

本文件描述 PVM JIT recompiler 的整體架構、模組關係、以及關鍵設計決策。
配合 `SPEC_opcode_emit.md`（翻譯規則）和 `SPEC_pvm_summary.md`（PVM 規格）使用。

**Interpreter Reference**（用於驗證語意正確性）：
- `PVM/instructions.go` — 所有 231 條 opcode 的 reference 實作
- `PVM/decode.go` — operand decode 精確公式
- `PVM/branch.go` — djump / branch 解析邏輯
- `PVM/exit_reason.go` — ExitReason 編碼（type << 56 | payload）
- `PVM/instructions_instrmeta.go` — pre-decode 的 InstrMeta 結構

---

## 1. 模組總覽

```
PVM/
├── asm/                    # x86-64 assembler library
│   ├── assembler.go        # Assembler struct, NewAssembler, Finalize, Reset
│   ├── buffer.go           # CodeBuffer: Emit, BindLabel, UseLabel32, ResolveFixups
│   ├── encoding.go         # rexByte, emitREX, modRM, sib helpers
│   ├── instructions.go     # MovRegToReg, Add, CmpRegImm32, Jcc, etc.
│   └── registers.go        # Register enum (RAX=0..R15=15), ConditionCode enum
│
├── recompiler/
│   ├── register_map.go     # PVMToX86[13], RegGuestBase=R15, RegScratch=RCX
│   ├── context.go          # JITContext: mmap layout, control region offsets, pages map
│   ├── guest_memory.go     # InitFromProgram, mapSegment, SetPageAccess, GuestMemory iface
│   ├── executable.go       # ExecutableMemory: dual-mapping (memfd + RW/RX views)
│   ├── trampoline.go       # EmitEntryTrampoline, EmitExitTrampoline, EmitHostCallExit
│   ├── compiler.go         # Compiler struct, opcodeHandlers[231], CompileBasicBlock
│   ├── gas.go              # emitGasCheck (per-instr v0.7.2), emitBlockGasCheck (v0.8.0)
│   ├── emit_basic.go       # emitTrap, emitFallthrough, emitEcalli, emitLoadImm, etc.
│   ├── emit_branch.go      # emitJump, emitJumpInd, emitBranchImm, emitBranch
│   ├── emit_arith_three.go # 32/64-bit add/sub/mul/div/rem/shift/bitwise
│   ├── emit_arith_imm.go   # arithmetic with immediate operand
│   ├── emit_two_reg.go     # sbrk, move_reg, bit manipulation, sign/zero extend
│   ├── emit_memory.go      # load/store (1/2/4/8 bytes, signed/unsigned)
│   ├── djump_native.go     # djumpSupport, emitDjumpNative, registerDispatch
│   ├── recompiler.go       # Recompiler struct, BlockBasedInvoke, lookupOrCompileBlock
│   ├── execute.go          # executeBlockLocked, callNative, HandleSbrk
│   ├── host.go             # host struct, HostCall dispatch loop (omega integration)
│   ├── invoke_mode.go      # MachineInvoke → BlockBasedInvoke
│   ├── code_cache.go       # CodeCache: PC→CompiledBlock map
│   └── x86signal/          # Signal handler (CGo)
│       ├── x86_signal_linux.c  # C handler: SIGSEGV/SIGBUS/SIGFPE → ExitReason
│       ├── x86_signal_linux.h  # offsets, register indices, macros
│       ├── x86_signal.go       # Go bindings: SetupSignalHandler, SetFaultWindow
│       └── x86_signal_stub.go  # no-op for non-linux/non-amd64
```

---

## 2. Memory Layout（mmap 配置）

### 2.1 Guest Memory Region

```
unix.Mmap(size = 4GB + 4KB control + 4KB guard, MAP_PRIVATE|MAP_ANON|MAP_NORESERVE)

Layout:
┌──────────────────┬──────────────────────────────────┬───────────┐
│  Control Region  │        Guest Memory (4GB)         │ Guard Page│
│    4096 bytes    │                                    │  4096 B   │
└──────────────────┴──────────────────────────────────┴───────────┘
                   ↑
                   R15 = guestBasePtr
```

- `guestBasePtr` = `&rawMem[4096]` → R15 in native code
- Control region = `rawMem[0:4096]`, accessed as `[R15 - offset]`
- Guard page = PROT_NONE, catches off-by-one overflows

### 2.2 Executable Memory（Dual Mapping）

```
memfd_create("jit-code")
ftruncate(fd, 16MB)

rwMem = mmap(fd, PROT_READ|PROT_WRITE, MAP_SHARED)  → 寫入 native code
rxMem = mmap(fd, PROT_READ|PROT_EXEC, MAP_SHARED)   → 執行 native code

same physical pages, no mprotect needed
```

### 2.3 Control Region Fields

```
R15 - 8:    ReturnStack   (uintptr)  — Go's RSP, for signal handler restore
R15 - 16:   ReturnAddr    (uintptr)  — Go's return_label address
R15 - 24:   HeapPointer   (uint64)   — current sbrk boundary
R15 - 32:   ExitPC        (uint32)   — PVM PC on exit (+4B padding)
R15 - 40:   ExitReason    (uint64)   — why execution stopped
R15 - 48:   Gas           (int64)    — remaining gas (disp8 reachable!)
R15 - 152:  Registers     ([13]uint64 = 104 bytes)
R15 - 160:  MemAccessAddr (uint32)   — trace: last memory access address
R15 - 168:  MemAccessVal  (uint64)   — trace: last memory access value
R15 - 176:  DjumpTable    (uintptr)  — jump table rodata pointer
R15 - 184:  DjumpBitmask  (uintptr)  — bitmask rodata pointer
R15 - 192:  DjumpDispatch (uintptr)  — PC→native dispatch table pointer
```

---

## 3. Register Mapping

```
PVM Index  Name    x86-64 Register   Rationale
─────────  ────    ───────────────    ─────────────────────────
0          RA      RAX               DIV/MUL 使用，compact encoding
1          SP      RDX               DIV remainder，compact
2          T0      RBX               callee-saved, compact
3          T1      RSI               compact (no REX)
4          T2      RDI               compact (no REX)
5          S0      R8                caller-saved extended
6          S1      R9                caller-saved extended
7          A0      R10               caller-saved extended
8          A1      R11               caller-saved extended
9          A2      R12               callee-saved extended
10         A3      R13               callee-saved extended
11         A4      R14               callee-saved extended
12         A5      RBP               callee-saved, compact

Reserved:
─────────
R15        RegGuestBase              guest memory base + control region
RCX        RegScratch                scratch for DIV(CL), shifts, address calc
RSP        (implicit)                native stack pointer
```

---

## 4. Execution Flow

```
                    ┌──── Go Runtime ────┐        ┌─── Native Code ───┐
                    │                    │        │                    │
 HostCall/Invoke    │  BlockBasedInvoke  │        │   Compiled Blocks  │
      ↓            │       │            │        │        ↑           │
      │            │  lookupOrCompileBlock        │        │           │
      │            │       │            │        │   entry_trampoline │
      │            │  callNative ──────────────→ │        │           │
      │            │       ↑            │        │   [native code]    │
      │            │       │            │        │        │           │
      │            │  (return_label) ←───────── │   exit_trampoline  │
      │            │       │            │        │                    │
      │            │  read ExitReason   │        └────────────────────┘
      │            │       │            │
      │            │  switch:           │
      │            │    CONTINUE → loop │
      │            │    HOST_CALL → dispatch omega
      │            │    HALT/PANIC/OOG → return
      │            │    DjumpCallID → resolveDjump
      │            │    SbrkCallID → HandleSbrk + recompile suffix
      │            └────────────────────┘
```

### 4.1 Entry Path

1. `BlockBasedInvoke(startPC)` — main loop
2. `lookupOrCompileBlock(pc)` — cache hit or compile
3. `executeBlockLocked(block)`:
   - `runtime.LockOSThread()`
   - `SetFaultWindow(guestBase, codeStart, codeEnd)` — TLS for signal handler
   - `callNative(guestBase, blockAddr, trampolineAddr)` — Go→native
4. Entry trampoline runs (see SPEC_opcode_emit.md §8)
5. Native code executes until exit

### 4.2 Exit Path

Exit reasons（control region 使用 PVM package 的 `ExitReason` 格式：`type<<56 | payload`）：
- **Gas exhaustion**: `emitGasCheck` → JS oog_label → ExitReason=ExitOOG
- **Block end / branch**: `emitExitToPC` → ExitReason=ExitContinue(0), ExitPC=target
- **Host call**: `emitEcalli` → ExitReason=ExitHostCall|callID
- **Halt**: jump_ind to 0xFFFF0000 → ExitReason=ExitHalt
- **Panic**: trap / invalid target → ExitReason=ExitPanic
- **Signal**: SIGSEGV → signal_handler → ExitReason=PAGE_FAULT|faultAddr or ExitPanic
- **sbrk runtime**: → ExitReason=ExitHostCall|SbrkCallID(0xFF)（內部 sentinel，不是真的 host call）
- **djump miss**: → ExitReason=ExitHostCall|DjumpCallID(0xFE)（內部 sentinel）

All paths store registers → `exit_trampoline` → restore RSP → JMP return_label → back to Go.

### 4.3 Internal Sentinel Exit IDs

```go
SbrkCallID  = 0xFF  // sbrk 跨頁需要 mprotect → exit to Go
DjumpCallID = 0xFE  // djump dispatch miss → exit to Go for compile + retry
```

Go 側 `BlockBasedInvoke` 先檢查 sentinel，不傳給外部 host。

### 4.4 Signal Handler

- C handler registered via `sigaction(SIGSEGV/SIGBUS/SIGFPE)`
- TLS checks: `jit_code_start <= RIP < jit_code_end` → JIT fault
- If JIT fault:
  - `store_pvm_regs` → write 13 registers from ucontext to control region
  - Set `ExitReason` (PAGE_FAULT for SIGSEGV/SIGBUS, PANIC for SIGFPE)
  - Modify `ucontext.RIP = ReturnAddr`, `ucontext.RSP = ReturnStack`
  - Return from signal → resumes at Go's `return_label`
- If not JIT fault: chain to old handler

---

## 5. Block Linking（Static）

Compiler pre-compiles successor blocks **before** emitting current block:

```go
linkFallthrough, _ = compileForLink(fallthroughPC, depth+1)
linkTaken, _       = compileForLink(branchTarget, depth+1)
```

If successor is compiled, terminator emits direct `JMP native_addr` instead of exit-to-Go:

```go
func emitLinkOrExit(a, link *CompiledBlock, targetPC):
    if link != nil:
        JMP link.NativeAddr    // direct native→native
    else:
        emitExitToPC(targetPC)  // exit to Go dispatcher
```

Cycle detection: `c.linking[pc]` map prevents infinite recursion on back-edges.

---

## 6. Djump (Indirect Jump) Native Dispatch

For `jump_ind` / `load_imm_jump_ind`:

```
1. target = uint32(Reg[rA] + offset)
2. if target == 0xFFFF0000 → HALT
3. Validate: target != 0, target % 2 == 0, target/2 <= tableSize
4. index = target/2 - 1
5. Load destPC from jump table rodata
6. Check bitmask[destPC] == 0x03 (basic block start)
7. Load dispatch[destPC] (native address, set at compile time)
8. If dispatch[destPC] != 0 → JMP native_addr (HIT, stay in native)
9. If dispatch[destPC] == 0 → exit to Go with DjumpCallID (MISS)
```

Data structures:
- `djumpTable` (rodata): jump table entries in ExecutableMemory
- `djumpBitmask` (rodata): per-PC basic-block-start markers
- `djumpDispatch` (Go heap): `[]uintptr` updated by `registerDispatch` when blocks compile

---

## 7. Host Call (ecalli) Flow

```
Native:
    MOV [R15-40], (HOST_CALL | callID<<8)   // ExitReason
    MOV [R15-32], nextPC                     // ExitPC
    JMP exit_trampoline

Go (host.HostCall):
    loop:
        exitReason = MachineInvoke(pc, gas)
        if exitReason == HOST_CALL:
            callID = exitReason >> 8
            snapshot registers, gas
            result = omega(callID, registers, gas, guestMemory)
            writeback registers, gas
            pc = ExitPC (next instruction)
            continue
        else:
            return exitReason
```

### GuestMemory Interface

```go
type GuestMemory interface {
    Read(addr uint32, size int) ([]byte, error)
    Write(addr uint32, data []byte) error
    ReadUint32(addr uint32) (uint32, error)
    // ... etc
}
```

Zero-copy: returns slices into the mmap'd guest memory directly.
Layer 1 check: validates against `ctx.pages` before pointer arithmetic.

---

## 8. sbrk Handling

Two paths:
- **Inline (no page crossing)**: update heapPointer + set rD in native code
- **Runtime exit (page crossing)**: exit with `SbrkCallID=0xFF`, Go calls `HandleSbrk`:
  - `mprotect(newPages, PROT_READ|PROT_WRITE)`
  - Update `ctx.pages` (Layer 1 sync)
  - Write back heapPointer and rD to control region
  - Recompile block suffix starting from next instruction

---

## 9. Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| 13 PVM regs → 13 x86 regs (zero spill) | x86-64 has 16 GPRs; reserve 3 (R15, RCX, RSP) → 13 remaining |
| R15 = guest base + control region | Single register for all guest memory + VM state access |
| RCX = scratch | Required by x86 DIV (CL) and shift instructions |
| Dual mapping (no mprotect toggle) | Eliminates 49% overhead from W^X switching |
| Signal handler instead of bounds checks | Zero overhead on valid accesses; hardware MMU does the work |
| Per-instruction gas (v0.7.2) | Exact PC on OOG; v0.8.0 switches to per-block |
| Block linking (depth-limited) | Eliminates Go dispatcher round-trip for sequential/branch targets |
| pvmRegSlot reordering | RA/SP at slots 10,11 → disp8 offsets for frequent DIV spill paths |
| MAP_NORESERVE | 4GB virtual space without committing physical memory |
| INT3 fill on reset | Catches accidental execution of stale/uninitialized code regions |

---

## 10. Build Constraints

```go
//go:build linux && amd64
```

- Linux-only (signal handler + memfd_create + mmap semantics)
- amd64-only (register map, instruction encoding)
- CGo required (signal handler in C for TLS + ucontext access)
- Go 1.17+ (internal ABI register-based calling convention)

---

## 11. Pre-Decode Phase（Program 結構）

Recompiler 不直接解碼 raw bytecode，而是使用 `PVM.Program`（由 `DeBlobProgramCode` 產生）：

```go
type Program struct {
    InstructionData ProgramCode   // raw opcode stream
    Bitmasks        Bitmask       // per-PC basic-block-start marker
    JumpTable       JumpTable     // indirect jump targets

    Instrs     []InstrMeta        // pre-decoded flat array
    BlockAt    []*BlockMeta       // PC-indexed: block start markers
    InstrIdxAt []int32            // PC-indexed: instruction index lookup
}

type InstrMeta struct {
    PC      ProgramCounter        // instruction 的 PVM PC
    Opcode  byte                  // opcode number (0-230)
    SkipLen uint8                 // operand bytes (instruction length = SkipLen + 1)
    Dst     uint8                 // destination register index (0xFF = none)
    Src     [2]uint8              // source register indices (0xFF = unused)
    Imm     [2]uint64             // pre-decoded immediates / branch target PC
}

type BlockMeta struct {
    StartPC    ProgramCounter
    EndPC      ProgramCounter     // last instruction PC (inclusive)
    InstrStart int                // index into Program.Instrs[]
    InstrEnd   int                // exclusive upper bound
    GasCost    Gas                // = InstrCount (v0.7.2)
}
```

Recompiler 的 `CompileBasicBlock(startPC)` 取得 `BlockMeta` → iterate `Instrs[InstrStart:InstrEnd]`。
每條 `InstrMeta` 的 `Imm[]` 已 pre-decoded（sign-extended），register indices 在 `Dst` / `Src[]`。

---

## 12. File Dependencies（Build Order）

```
asm/registers.go          → no deps
asm/encoding.go           → registers.go
asm/buffer.go             → no deps
asm/instructions.go       → encoding.go, buffer.go, registers.go
asm/assembler.go          → buffer.go

recompiler/register_map.go     → asm/
recompiler/context.go          → asm/, PVM/
recompiler/executable.go       → unix (syscalls)
recompiler/guest_memory.go     → context.go, unix
recompiler/trampoline.go       → asm/, context.go, register_map.go
recompiler/gas.go              → asm/, context.go
recompiler/emit_*.go           → asm/, context.go, register_map.go, PVM/
recompiler/djump_native.go     → asm/, context.go, executable.go
recompiler/compiler.go         → all emit_*.go, asm/, PVM/
recompiler/code_cache.go       → executable.go
recompiler/execute.go          → context.go, code_cache.go, x86signal/
recompiler/recompiler.go       → compiler.go, execute.go, code_cache.go
recompiler/host.go             → recompiler.go, context.go, PVM/
recompiler/x86signal/          → C (CGo), context.go offsets
```
