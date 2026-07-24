//go:build linux && amd64 && cgo

package main

import "github.com/New-JAMneration/JAM-Protocol/PVM/recompiler"

// flushPVMProfile emits a final [JIT-PROFILE] line on shutdown so short runs
// (which exit before the 10s ticker fires) still produce output. No-op unless
// JIT_PROFILE=1.
func flushPVMProfile() { recompiler.FlushProfile() }
