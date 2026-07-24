//go:build !(linux && amd64 && cgo)

package x86_signal_linux

// SetupSignalHandler is a no-op on unsupported platforms.
func SetupSignalHandler() {}

// SetVmCtx is a no-op on unsupported platforms.
func SetVmCtx(guestBase uintptr) {}

// SetFaultWindow is a no-op on unsupported platforms.
func SetFaultWindow(guestBase, codeStart, codeEnd uintptr) {}

// ClearFaultWindow is a no-op on unsupported platforms.
func ClearFaultWindow() {}
