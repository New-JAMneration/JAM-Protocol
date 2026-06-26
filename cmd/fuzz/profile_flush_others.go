//go:build !(linux && amd64)

package main

// flushPVMProfile is a no-op: the recompiler backend (and its JIT profiler) only
// builds on linux/amd64.
func flushPVMProfile() {}
