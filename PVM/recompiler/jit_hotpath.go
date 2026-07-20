//go:build linux && amd64 && cgo

package recompiler

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

// Hot-path profiling for the JIT recompiler. Set JIT_PROFILE=1 to enable.
// When off, each hook is only a bool check — normal runs are unaffected.
//
// Scope: this is a phase/cost-attribution + counter profiler, NOT a hot-block
// detector. The directly-measured time buckets (setup/deblob/compile/host/run)
// are kept for cheap online monitoring and per-call averages. The old
// subtracted "exec(est) = run - compile - host" bucket has been retired: it
// conflated native execution with dispatch glue (LockOSThread, SetFaultWindow,
// snapshot, cache lookup, djump/sbrk resolution) and so could not tell a slow
// JIT from slow glue. For native-vs-glue attribution use pprof; for per-PVM-block
// hotness use perf + a JIT symbol map. See PVM/docs/JIT_PROFILE_ANALYSIS.md.
//
// Output: the periodic ticker prints the cumulative running total, skipping ticks
// where nothing changed (so idle windows / finished runs don't spam duplicates).
// FlushProfile prints the final cumulative TOTAL once on shutdown — that last line
// is the authoritative "whole run" number.
var jitProfile = os.Getenv("JIT_PROFILE") == "1"

// How often to print the summary to stderr.
var jitProfileInterval = 10 * time.Second

// jm holds process-wide counters and directly-measured time buckets.
// Parallel goroutines may overlap in wall time — compare ratios, not absolute ms.
var jm struct {
	invokes     atomic.Int64 // Psi_M invoke count
	invokeNanos atomic.Int64 // total time per invoke
	setupNanos  atomic.Int64 // JIT context + executable memory + guest memory init
	deblobNanos atomic.Int64 // program decode (GetOrDeblobProgram)
	runNanos    atomic.Int64 // whole host.HostCall loop (compile + native exec + glue + host)

	compileCalls atomic.Int64 // basic block compile count
	compileNanos atomic.Int64 // compile time (top-level only)

	roundTrips atomic.Int64 // Go→native trampoline entries (one per executeBlockLocked)
	lockCalls  atomic.Int64 // MachineInvoke / LockOSThread count

	hostCalls atomic.Int64 // host call (omega) count
	hostNanos atomic.Int64 // time spent in omega

	cacheHits   atomic.Int64 // block cache hit in lookupOrCompileBlock
	cacheMisses atomic.Int64 // block cache miss (compile needed)

	djumpResolves atomic.Int64 // djump resolved on Go side
}

// jmSnapshot is a plain-value copy of jm at one instant, used to compute
// per-interval deltas without re-loading atomics mid-format.
type jmSnapshot struct {
	invokes, invokeNanos, setupNanos, deblobNanos, runNanos int64
	compileCalls, compileNanos                              int64
	roundTrips, lockCalls                                   int64
	hostCalls, hostNanos                                    int64
	cacheHits, cacheMisses                                  int64
	djumpResolves                                           int64
}

func loadJMSnapshot() jmSnapshot {
	return jmSnapshot{
		invokes: jm.invokes.Load(), invokeNanos: jm.invokeNanos.Load(),
		setupNanos: jm.setupNanos.Load(), deblobNanos: jm.deblobNanos.Load(),
		runNanos:     jm.runNanos.Load(),
		compileCalls: jm.compileCalls.Load(), compileNanos: jm.compileNanos.Load(),
		roundTrips: jm.roundTrips.Load(), lockCalls: jm.lockCalls.Load(),
		hostCalls: jm.hostCalls.Load(), hostNanos: jm.hostNanos.Load(),
		cacheHits: jm.cacheHits.Load(), cacheMisses: jm.cacheMisses.Load(),
		djumpResolves: jm.djumpResolves.Load(),
	}
}

func init() {
	if !jitProfile {
		return
	}
	go func() {
		var last jmSnapshot
		for range time.Tick(jitProfileInterval) {
			cur := loadJMSnapshot()
			if cur == last {
				continue // nothing changed since last dump — skip duplicate
			}
			last = cur
			printJITProfile("cum", cur)
		}
	}()
}

// FlushProfile prints a final cumulative [JIT-PROFILE TOTAL] line. Callers should
// defer it from main (or call on shutdown) so short-lived runs — which exit before
// the ticker fires — still emit output. No-op when profiling is disabled.
func FlushProfile() {
	if !jitProfile {
		return
	}
	printJITProfile("TOTAL", loadJMSnapshot())
}

// printJITProfile writes one breakdown line for the given snapshot (cumulative
// total or per-interval delta). Time buckets are directly measured (no
// subtraction); exec time is intentionally NOT derived — use pprof to split
// native vs dispatch glue.
func printJITProfile(tag string, s jmSnapshot) {
	ms := func(n int64) int64 { return n / 1_000_000 }
	avg := func(tot, n int64) int64 {
		if n == 0 {
			return 0
		}
		return tot / n
	}
	ratio := func(a, b int64) float64 {
		if b == 0 {
			return 0
		}
		return float64(a) / float64(b)
	}

	fmt.Fprintf(os.Stderr,
		"[JIT-PROFILE %s] invokes=%d | time(ms): invoke=%d setup=%d deblob=%d run=%d compile=%d host=%d | "+
			"avg(ns): compile=%d host=%d | "+
			"counts: roundTrips=%d lock=%d host=%d compile=%d djump=%d cache(hit/miss)=%d/%d | "+
			"ratios: roundTrips/lock=%.2f\n",
		tag,
		s.invokes,
		ms(s.invokeNanos), ms(s.setupNanos), ms(s.deblobNanos), ms(s.runNanos), ms(s.compileNanos), ms(s.hostNanos),
		avg(s.compileNanos, s.compileCalls), avg(s.hostNanos, s.hostCalls),
		s.roundTrips, s.lockCalls, s.hostCalls, s.compileCalls, s.djumpResolves,
		s.cacheHits, s.cacheMisses,
		ratio(s.roundTrips, s.lockCalls),
	)
}
