# Recompiler 優化 TODO

以下是已確認值得實作的優化項目。前提：正確性已通過 conformance test。

---

## 1. Block Chaining（jump / branch linking）

**現狀**：fallthrough block linking 已實作（`block_link.go`）。當 block A 的 sequential successor（block B）已編譯完，A 的 epilogue 直接 `JMP` 到 B 的 native address，不回 Go dispatcher。

**尚未做**：static jump 和 branch 的 linking。目前 `jump` 和 `branch_eq` 等指令的目標即使已編譯，仍走 `emitExitToPC` → 回 Go → `lookupOrCompileBlock` → 重新進入 native。

**目標**：當 branch/jump 的靜態目標已編譯，直接 emit `JMP targetBlock.NativeAddr`，消除 Go dispatcher round-trip。

**設計選項**：

| 策略 | 優點 | 缺點 |
|------|------|------|
| (a) 先遞迴 compile 目標再 link | 無 runtime patch、實作直覺 | 可能 compile 未執行的 cold block |
| (b) stub + runtime patch | 只 compile 熱路徑 | dual mapping 下 patch 簡單，但需記錄 patch site |
| (c) PC → native dispatch table | 漸進填表、可與 djump 收斂 | 多一次間接跳 |

目前 fallthrough 用策略 (a)（`compileForLink` 遞迴 compile），jump/branch 可延續相同策略。

**預期效益**：消除 per-block trampoline 開銷（實測約佔 execute 時間 ~40%）。配合 dual-mapping（已完成），Go dispatcher 只在 host call / sbrk / djump miss / OOG / halt 才觸發。

**相關檔案**：
- `block_link.go`：`emitFallthroughEpilogue`、`emitLinkOrExit`、`compileForLink`
- `compiler.go`：`linkFallthrough` / `linkTaken`、block epilogue emit

---

## 2. 常數折疊（Constant Folding）

**現狀**：single-pass compiler 逐條指令 emit，不做跨指令分析。每條 `load_imm` 都 emit `MOV r64, imm64`（10 bytes），後續使用該 register 的指令再 emit 完整的 reg-reg 操作。

**機會**：如果 `load_imm` 的目標 register 緊接著被一條算術指令消耗（且中間無 branch 進入），可以直接用 immediate 編碼：

```
// 現在 emit 的：
MOV R10, 42          // 10 bytes (movabs)
ADD RAX, R10         // 3 bytes

// 優化後：
ADD RAX, 42          // 7 bytes (ADD r64, imm32) 或 4 bytes (ADD r64, imm8)
```

**實作思路**：

1. 在 compile loop 中維護一個 per-register 的 `knownImm` map
2. `load_imm` 時不立即 emit，而是記錄 `knownImm[dst] = imm`
3. 下一條指令使用該 register 時，如果 imm 在 imm8/imm32 範圍內，直接用 immediate encoding
4. 如果下一條指令是 branch target（有其他路徑跳入）或 register 被其他指令修改前未消耗 → flush `knownImm`，正常 emit `MOV r64, imm`

**注意**：
- Single-pass 限制：只能看「前一條」，不能回頭改
- Branch target 會 invalidate 所有 knownImm（因為其他路徑可能帶不同值進來）
- 這是 peephole optimization，不是完整的 constant propagation

**預期效益**：減少 code size（少 emit movabs）+ 可能減少 register pressure（scratch 少用一次）。在 `load_imm + op` 密集的 code 中效益明顯。

---

## 3. Memory Access 合併 Bounds Check

**現狀**：目前 PVM 的 memory access **不做 software bounds check**，完全依賴硬體 MMU（PROT_NONE page → SIGSEGV → signal handler → ExitPageFault）。這是零成本的。

**但有另一種 overhead**：每個 load/store emit 都是獨立的 `MOV ECX, addr_reg` + `MOV dst, [R15+RCX]`，即使連續存取已知在同一頁的位址。

**機會**：連續 load/store 且 offset 差小於 page size（4KB）時，可以省掉重複的 address setup：

```
// 現在（兩次獨立的 address calculation）：
MOV ECX, R10         // addr for load_u32 [r7]
MOV EAX, [R15+RCX]
MOV ECX, R10         // addr for load_u32 [r7+4]  ← 重複！
ADD ECX, 4
MOV EDX, [R15+RCX]

// 優化後（reuse base）：
MOV ECX, R10
MOV EAX, [R15+RCX]
MOV EDX, [R15+RCX+4]  // 直接用 disp 偏移
```

**實作思路**：

1. 在 compile loop 追蹤「上一次 memory access 的 base register + 已知 address expression」
2. 如果下一條 memory op 的 base 相同且 offset 差在 disp8/disp32 範圍內：
   - 直接用 `[R15 + RCX + disp]` 定址（SIB + displacement）
   - 省掉重複的 `MOV ECX, reg` + `ADD ECX, offset`
3. 如果中間有可能改變 address register 的指令 → invalidate tracking

**注意**：
- PVM 地址是 32-bit wrap-around，offset 計算要用 32-bit 算術
- 如果兩次存取跨頁（addr 在 page boundary 附近），第二次可能 fault 但第一次沒有 → 語意正確（signal handler 照樣接住）
- 這個優化只省 code size 和 instruction count，不影響正確性

**預期效益**：在 struct field access 密集的 code（連續 load 同一 base + 不同 offset）中，每組省 2-3 條 x86 指令。

---

## 優先序建議

```
1. Block Chaining (jump/branch)  ← 效益最大（消除 ~40% execute overhead）
2. 常數折疊                      ← 中等效益、低風險
3. Memory Access 合併            ← 小效益、需 careful tracking
```

Block chaining 的前置條件（dual-mapping）已完成，可以直接做。
