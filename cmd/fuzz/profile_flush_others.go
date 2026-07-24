//go:build !(linux && amd64 && cgo)

package main

// flushPVMProfile is a no-op: the recompiler backend (and its JIT profiler) only
// builds on linux/amd64.
func flushPVMProfile() {}
