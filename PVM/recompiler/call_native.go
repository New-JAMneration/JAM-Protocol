//go:build linux && amd64

package recompiler

// callNative is implemented in call_native_amd64.s.
// Go internal ABI: RAX=guestBase, RBX=blockAddr, RCX=trampolineAddr.
func callNative(guestBase uintptr, blockAddr uintptr, trampolineAddr uintptr)
