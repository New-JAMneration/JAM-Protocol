//go:build linux && amd64 && cgo

package recompiler

import (
	"fmt"
	"os"
	"sync"
)

// JIT symbol map for Linux perf. Set JIT_PERFMAP=1 to enable.
//
// perf samples instruction addresses across the whole process, including the
// JIT-emitted machine code, but by default cannot symbolise the dynamic region.
// The standard fix is a /tmp/perf-<PID>.map file: one "START SIZE symbol" line
// per emitted region. With it, `perf report` / `perf annotate` attribute native
// samples to per-PVM-block symbols (pvm_block_<PC>) — the per-block hotness the
// subtraction profiler and pprof cannot give. See PVM/docs/JIT_PROFILE_ANALYSIS.md §6.3.
//
// Independent of JIT_PROFILE. Note: per-block native time is only observable
// where perf's PMU works (bare-metal Linux); WSL2 commonly restricts perf_events,
// so use pprof there. This only WRITES the map; running perf is out-of-band.
var perfMap = os.Getenv("JIT_PERFMAP") == "1"

var (
	// perfMapMu guards concurrent writes: distinct CodeHashes have separate
	// CompiledProgram.mu and may compile (and append) in parallel.
	perfMapMu   sync.Mutex
	perfMapFile *os.File
)

func init() {
	if !perfMap {
		return
	}
	path := fmt.Sprintf("/tmp/perf-%d.map", os.Getpid())
	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[JIT-PERFMAP] create %s: %v (disabled)\n", path, err)
		perfMap = false
		return
	}
	perfMapFile = f
	fmt.Fprintf(os.Stderr, "[JIT-PERFMAP] writing JIT symbol map to %s\n", path)
}

// recordPerfMapEntry appends one perf-map line for a freshly compiled block.
// Cheap bool check when disabled. Symbol is pvm_block_<PC>; names may repeat
// across CodeHashes (perf maps by address, so this only affects display).
func recordPerfMapEntry(b *CompiledBlock) {
	if !perfMap || perfMapFile == nil {
		return
	}
	perfMapMu.Lock()
	// Format perf expects: <hex start addr> <hex size> <symbol name>
	fmt.Fprintf(perfMapFile, "%x %x pvm_block_%d\n", b.NativeAddr, b.NativeSize, b.PVMStartPC)
	perfMapMu.Unlock()
}
