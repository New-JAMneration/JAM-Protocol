# JIT_PROFILE — Usage Guide

This guide covers **how to use it** and **how it works**.

---

## 0. 30-second background

- **PVM** is the virtual machine JAM uses to run service programs (guest code).
- It has two backends:
  - **interpreter** — executes one instruction at a time in Go.
  - **recompiler (JIT)** — compiles PVM basic blocks to **native x86-64** and runs
    that, which should be faster.
- **`JIT_PROFILE`** is the built-in tool for seeing *where the recompiler backend
  spends its time and how much work it does*.
- Only active on **linux/amd64** builds and only with the **recompiler** backend.

> Important: profiling must be enabled on the **process that actually runs the
> PVM**. In the fuzz setup that is the **target (server)**, not the fuzzer (client).

---

## 1. How to use (quick start)

### 1.1 Most common: the `[JIT-PROFILE]` summary line

Set `JIT_PROFILE=1`; a cumulative summary line is printed to stderr. Fuzzing uses
two terminals:

```bash
# Terminal 1 — target (the server that runs the PVM), with profiling on
JIT_PROFILE=1 make run-target PVM_BACKEND=recompiler

# Terminal 2 — fuzzer (feeds data to the server)
go run ./cmd/fuzz/ test_folder ./.jam_fuzz_docker_run/fuzz.sock \
    ./pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces/
```

While running, the target prints a cumulative line every 10s (skipped if nothing
changed). **When done, go back to Terminal 1 and press `Ctrl-C`** — it prints a
final `[JIT-PROFILE TOTAL]` line, which is the whole-run summary.

### 1.2 Going deeper: pprof — "where does the time actually go?"

`[JIT-PROFILE]` gives only coarse time + counts. To see the real split between
"native execution" and "the glue around it", use pprof. Add one env var:

```bash
# Add JIT_CPUPROFILE to write a CPU profile for the whole run
JIT_PROFILE=1 JIT_CPUPROFILE=/tmp/jit_cpu.out make run-target PVM_BACKEND=recompiler
# ...run the fuzzer, then Ctrl-C the target (the file is written on shutdown)...

# Look at PVM internals only (everything else filtered out)
go tool pprof -tagfocus='phase:pvm' -top /tmp/jit_cpu.out
go tool pprof -tagfocus='phase:pvm' /tmp/jit_cpu.out   # interactive: top / list / web
```

For a long-lived service you can grab a profile live instead:

```bash
JIT_PROFILE=1 JIT_PPROF_ADDR=localhost:6060 make run-target PVM_BACKEND=recompiler
# in another terminal, while the fuzzer runs:
go tool pprof -tagfocus='phase:pvm' "http://localhost:6060/debug/pprof/profile?seconds=30"
```

### 1.3 Expert: perf — "which PVM block is hottest?"

```bash
JIT_PROFILE=1 JIT_PERFMAP=1 perf record -g -- <program that runs the PVM>
perf report          # see which pvm_block_<PC> is hottest
```

> ⚠️ Needs **bare-metal Linux**. WSL2 usually restricts the PMU, so perf won't
> sample — use pprof there instead.

### 1.4 Environment variables

| Variable | Effect | Default |
|----------|--------|---------|
| `JIT_PROFILE=1` | Counters + `[JIT-PROFILE]` summary line (also enables the pprof `phase=pvm` label) | off |
| `JIT_CPUPROFILE=<path>` | Write a Go CPU profile to a file (whole run) | off |
| `JIT_PPROF_ADDR=host:port` | Serve `net/http/pprof` (grab anytime) | off |
| `JIT_PERFMAP=1` | Write `/tmp/perf-<pid>.map` for Linux perf | off |

All off by default; when off there is no performance impact (just a bool check).

---

## 2. Reading the `[JIT-PROFILE]` line

Example:

```
[JIT-PROFILE TOTAL] invokes=645 | time(ms): invoke=1835 setup=166 deblob=21 run=1614 compile=141 host=123 | avg(ns): compile=31822 host=5418 | counts: roundTrips=3311026 lock=23199 host=22717 compile=4452 djump=0 cache(hit/miss)=3306574/4452 | ratios: roundTrips/lock=142.72
```

Two concepts first. One **invoke** = running a guest program once (one `Psi_M`
call). It breaks down into phases:

```
invoke ≈ setup + deblob + run
run    = compile + exec(native) + host
```

| Field | Meaning |
|-------|---------|
| `invokes` | How many guest programs were run |
| `time(ms) invoke` | Total time across all invokes |
| `time(ms) setup` | Building the JIT context / memory |
| `time(ms) deblob` | Decoding the program |
| `time(ms) run` | The whole execution loop (compile + native + host calls) |
| `time(ms) compile` | Compiling PVM blocks to x86 |
| `time(ms) host` | Guest calling back into the host (omega) |
| `avg(ns) compile/host` | Average time per compile / per host call |
| `counts roundTrips` | **Go↔native round-trips** (one per entry into native) |
| `counts lock` | `MachineInvoke` count (times we enter the native run loop) |
| `counts host` | Number of host calls |
| `counts compile` | Blocks compiled (= cache misses) |
| `counts djump` | Indirect jumps resolved on the Go side |
| `counts cache(hit/miss)` | Block compile-cache hits / misses |
| `ratios roundTrips/lock` | Avg native round-trips per `MachineInvoke` |

**Note: there is no `exec` time field.** "Pure native execution" and "the glue of
entering/leaving native" are mixed together, so a single number would mislead — to
split them, use pprof (see §3).

Two more caveats:
- `invoke ≈ setup + deblob + run` is an approximation, not an identity: per-invoke
  work outside the sub-buckets (artifact acquire/bind, register writes, trace init)
  lands in the gap.
- Numbers from a **`-tags trace` build are not comparable** to a standard build:
  single-step mode enters native once per *instruction* (roundTrips explodes) and
  compiles outside the counted path.

---

## 3. How it works

`JIT_PROFILE` is really **three layers**, each answering a different question:

| Layer | Tool | Question it answers | How to enable |
|-------|------|---------------------|---------------|
| Counters + coarse timing | the `[JIT-PROFILE]` line | how many runs, rough time per phase, how many round-trips | `JIT_PROFILE=1` |
| Precise CPU attribution | **pprof** | where the time *really* goes per Go function (native vs glue) | `JIT_CPUPROFILE` / `JIT_PPROF_ADDR` |
| Native-code detail | **perf** | which PVM block / x86 instruction is hottest | `JIT_PERFMAP=1` + perf |

### 3.1 Counter layer (`jit_hotpath.go`)

A handful of atomic counters/timers sit at key points in the recompiler (entering
`Psi_M`, compiling a block, entering/leaving native, host calls…). With
`JIT_PROFILE=1`:

- a background goroutine prints the cumulative totals every 10s (skipped if
  unchanged, so it doesn't spam);
- on shutdown it prints a final `[JIT-PROFILE TOTAL]` (the authoritative summary).

The counters are **process-wide and cumulative** — they describe how much work the
**PVM internals** did, *not* the fraction of the whole system.

### 3.2 pprof layer

`JIT_CPUPROFILE` / `JIT_PPROF_ADDR` enable Go's built-in sampling CPU profiler,
which attributes CPU time to real functions. Each `Psi_M_recompiler` call is tagged
with a `phase=pvm` label, so `-tagfocus='phase:pvm'` shows **only the PVM** and
filters out the rest of the system (VRF, storage, etc.).

Two limits to know:
- pprof **cannot see inside the JIT-generated machine code** (it's not in the symbol
  table). Native execution shows up as one opaque lump — `runtime._ExternalCode` /
  `callNative`. You learn *how much* is native, but not *which block*.
- pprof reports **CPU sample percentages**, not the cumulative ms of `[JIT-PROFILE]`
  — don't subtract one from the other.

### 3.3 perf layer

To find *which PVM block* is hottest you need perf. `JIT_PERFMAP=1` writes each
compiled block's "address, size, name (`pvm_block_<PC>`)" to `/tmp/perf-<pid>.map`;
perf reads that file to map native addresses back to block names. Bare-metal Linux
only.

> Cache eviction caveat: evicted artifacts (`JIT_CACHE_MAX_PROGRAMS`) unmap their
> arenas and the address range can be reused by later blocks, leaving stale map
> lines. When profiling with perf, raise the cap so nothing evicts during the run
> (proper fix would be jitdump + `perf inject --jit`).

---

## 4. How to interpret the numbers (important)

1. **These numbers are "PVM internal", not "share of the whole system".**
   On real workloads the PVM is often only a few percent of total node CPU (most goes
   to crypto verification, state storage, etc.). So "some PVM phase is heavy" does not
   mean it's the system bottleneck.

2. **To compare recompiler vs interpreter, compare PVM-only time, not wall time.**
   If PVM is only ~1.5% of the system, even making PVM twice as fast moves wall time
   by under 1% (a ceiling effect). Do it right: run the same data on each backend and
   compare `Psi_M` time:
   ```bash
   go tool pprof -focus='PVM\.Psi_M' -top /tmp/interp_cpu.out   # interpreter
   go tool pprof -tagfocus='phase:pvm' -top /tmp/jit_cpu.out    # recompiler
   ```

3. **The JIT only shines on compute-heavy workloads.**
   If the guest constantly makes host calls / grows the heap (`roundTrips/lock` large,
   native share tiny), the JIT's strength — running hot loops as native code — never
   kicks in. You need compute-dense guests to see the gap.

4. **Enough samples matter.**
   A short run (e.g. one small dataset) may give pprof only a few dozen samples, which
   is statistically weak. The larger/longer the run, the more stable the numbers.

5. **On WSL2 use pprof, not perf.** perf needs a hardware PMU, which WSL2 usually
   restricts.

---

## 5. Where the code lives

- Counters and the `[JIT-PROFILE]` line: `PVM/recompiler/jit_hotpath.go`
- perf symbol map: `PVM/recompiler/jit_perfmap.go`
- pprof wiring (env vars): `cmd/fuzz/profile.go`, `cmd/node/profile.go`
- The `phase=pvm` pprof label: `PVM/recompiler/psi_m_recompiler.go`