package PVMtrace

import (
	"os"
	"strings"

	"github.com/New-JAMneration/JAM-Protocol/logger"
)

var traceLogger = logger.GetLogger("pvmtrace")

const envHostCallLog = "JAM_PVM_HOSTCALL_LOG"

func hostCallDispatchLogEnabled() bool {
	switch strings.TrimSpace(os.Getenv(envHostCallLog)) {
	case "1", "true", "TRUE", "yes", "YES", "on", "ON":
		return true
	default:
		return false
	}
}

// LogHostCallDispatchEnv logs one host-call dispatch (registers / gas before omega) when the
// environment variable JAM_PVM_HOSTCALL_LOG is set to 1, true, yes, or on. Default is off so trace
// and normal runs do not flood stdout.
func LogHostCallDispatchEnv(serviceID uint32, opName string, regs any, gas int64) {
	if !hostCallDispatchLogEnabled() {
		return
	}
	traceLogger.Debugf("serviceID: %d, host-call: %s, regs: %v, gas: %d", serviceID, opName, regs, gas)
}
