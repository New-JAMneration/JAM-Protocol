package PVM

import "os"

// RecompilerDebugMode controls which execution mode the recompiler uses.
type RecompilerDebugMode string

const (
	RecompilerDebugOff        RecompilerDebugMode = ""
	RecompilerDebugSingleStep RecompilerDebugMode = "single-step"
)

// RecompilerDebugModeRuntime is set by config.ApplyPVMRecompilerDebug(...).
// PVM/JIT driver checks this to select production vs debug single-step invoke.
var RecompilerDebugModeRuntime RecompilerDebugMode

func init() {
	if v := os.Getenv("JAM_PVM_RECOMPILER_DEBUG"); v != "" {
		RecompilerDebugModeRuntime = RecompilerDebugMode(v)
	}
}
