package PVMtrace

import (
	"os"
	"strconv"
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
//
// A negative serviceID means the invocation has no service context (e.g. is_authorized / Psi_I,
// where HostCallArgs.ServiceID is nil) and is logged as "none".
func LogHostCallDispatchEnv(serviceID int64, opName string, regs any, gas int64) {
	if !hostCallDispatchLogEnabled() {
		return
	}
	sid := "none"
	if serviceID >= 0 {
		sid = strconv.FormatInt(serviceID, 10)
	}
	traceLogger.Debugf("serviceID: %s, host-call: %s, regs: %v, gas: %d", sid, opName, regs, gas)
}
