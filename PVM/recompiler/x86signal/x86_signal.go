//go:build linux && amd64 && cgo

package x86_signal_linux

// #cgo CFLAGS: -Wall -Wno-unused-variable
// #include "x86_signal_linux.h"
import "C"

import (
	"sync"
	"unsafe"
)

var signalOnce sync.Once

// SetupSignalHandler installs the SIGSEGV/SIGBUS/SIGFPE handler that converts
// hardware faults in JIT code into ExitPanic/ExitPageFault.
// Safe to call multiple times; only the first call installs the handler.
func SetupSignalHandler() {
	signalOnce.Do(func() {
		C.setup_signal_handler()
	})
}

// SetVmCtx sets the per-OS-thread TLS pointer so the signal handler
// knows which guest memory region a fault belongs to.
// Must be called after runtime.LockOSThread() and before entering JIT code.
func SetVmCtx(guestBase uintptr) {
	C.set_vmctx(unsafe.Pointer(guestBase))
}

// SetFaultWindow records the guest base and JIT code bounds for the current
// OS thread so the signal handler can distinguish JIT faults from other crashes
// and handle SIGFPE from JIT IDIV instructions.
func SetFaultWindow(guestBase, codeStart, codeEnd uintptr) {
	C.set_fault_window(
		unsafe.Pointer(guestBase),
		unsafe.Pointer(codeStart),
		unsafe.Pointer(codeEnd),
	)
}

// ClearFaultWindow resets the per-thread fault window after leaving JIT code.
func ClearFaultWindow() {
	C.clear_fault_window()
}
