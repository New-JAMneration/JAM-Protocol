//go:build linux && amd64

package recompiler

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

// Hot-path profiling for the JIT recompiler. Set JIT_PROFILE=1 to enable.
// When off, each hook is only a bool check — normal runs are unaffected.
var jitProfile = os.Getenv("JIT_PROFILE") == "1"

// How often to print the summary to stderr.
var jitProfileInterval = 10 * time.Second

// jm holds process-wide counters. One invoke = setup + deblob + run;
// run = compile + native exec + host calls.
//
// Parallel goroutines may overlap in wall time — compare ratios, not absolute ms.
var jm struct {
	invokes     atomic.Int64 // Psi_M invoke count
	invokeNanos atomic.Int64 // total time per invoke
	setupNanos  atomic.Int64 // JIT context and executable memory init
	deblobNanos atomic.Int64 // program decode (DeBlobProgramCode)
	runNanos    atomic.Int64 // compile + native exec + host calls

	compileCalls atomic.Int64 // basic block compile count
	compileNanos atomic.Int64 // compile time (top-level only)

	execBlocks atomic.Int64 // native block runs (trampoline round-trips)
	lockCalls  atomic.Int64 // LockOSThread count

	hostCalls atomic.Int64 // host call (omega) count
	hostNanos atomic.Int64 // time spent in omega

	mprotectCalls atomic.Int64 // unused (legacy; dual-mapping removed mprotect)
	mprotectNanos atomic.Int64

	djumpResolves atomic.Int64 // djump resolved on Go side
}

func init() {
	if !jitProfile {
		return
	}
	go func() {
		for range time.Tick(jitProfileInterval) {
			dumpJITProfile()
		}
	}()
}

// dumpJITProfile prints the cumulative breakdown to stderr.
func dumpJITProfile() {
	ms := func(n int64) int64 { return n / 1_000_000 }
	run := jm.runNanos.Load()
	comp := jm.compileNanos.Load()
	host := jm.hostNanos.Load()
	execEst := run - comp - host
	if execEst < 0 {
		execEst = 0
	}
	fmt.Fprintf(os.Stderr,
		"[JIT-PROFILE] invokes=%d | sums(ms): invoke=%d setup=%d deblob=%d run=%d | "+
			"run-split(ms): compile=%d exec=%d(est) host=%d | "+
			"counts: compile=%d execBlocks=%d host=%d lock=%d djump=%d mprotect=%d(%dms)\n",
		jm.invokes.Load(),
		ms(jm.invokeNanos.Load()), ms(jm.setupNanos.Load()), ms(jm.deblobNanos.Load()), ms(run),
		ms(comp), ms(execEst), ms(host),
		jm.compileCalls.Load(), jm.execBlocks.Load(), jm.hostCalls.Load(),
		jm.lockCalls.Load(), jm.djumpResolves.Load(),
		jm.mprotectCalls.Load(), ms(jm.mprotectNanos.Load()),
	)
}
