package main

import (
	"net/http"
	_ "net/http/pprof" // registers /debug/pprof handlers on DefaultServeMux
	"os"
	"runtime/pprof"

	"github.com/New-JAMneration/JAM-Protocol/logger"
)

// startProfiling wires the "Go-side attribution" layer from
// PVM/docs/JIT_PROFILE_ANALYSIS.md §6.2. Both are off unless their env var is set:
//
//   - JIT_CPUPROFILE=<path>: write a Go CPU profile for the whole run.
//     Best for batch runs; analyse with `go tool pprof <path>`.
//   - JIT_PPROF_ADDR=host:port: serve net/http/pprof. Best for the long-lived
//     fuzz server; grab with `go tool pprof http://host:port/debug/pprof/profile`.
//
// Returns a stop func to defer (nil when JIT_CPUPROFILE is unset).
func startProfiling() func() {
	if addr := os.Getenv("JIT_PPROF_ADDR"); addr != "" {
		logger.Infof("JIT_PPROF_ADDR: serving pprof on http://%s/debug/pprof/", addr)
		go func() {
			if err := http.ListenAndServe(addr, nil); err != nil {
				logger.Errorf("JIT_PPROF_ADDR: %v", err)
			}
		}()
	}

	path := os.Getenv("JIT_CPUPROFILE")
	if path == "" {
		return nil
	}
	f, err := os.Create(path)
	if err != nil {
		logger.Errorf("JIT_CPUPROFILE: create %s: %v", path, err)
		return nil
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		logger.Errorf("JIT_CPUPROFILE: start: %v", err)
		_ = f.Close()
		return nil
	}
	logger.Infof("JIT_CPUPROFILE: writing CPU profile to %s", path)
	return func() {
		pprof.StopCPUProfile()
		_ = f.Close()
	}
}
