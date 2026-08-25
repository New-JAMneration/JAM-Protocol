package PVMtrace

import (
	"os"
	"strconv"
	"strings"

	"github.com/New-JAMneration/JAM-Protocol/logger"
)

var traceLogger = logger.GetLogger("pvmtrace")

const envHostCallLog = "JAM_PVM_HOSTCALL_LOG"

// Snapshotted at process start, matching JIT_PROFILE. Runtime os.Setenv changes
// are ignored so the host-call glue path never calls os.Getenv.
var hostCallDispatchLogEnabled = parseHostCallDispatchLogEnv(os.Getenv(envHostCallLog))

func parseHostCallDispatchLogEnv(v string) bool {
	switch strings.TrimSpace(v) {
	case "1", "true", "TRUE", "yes", "YES", "on", "ON":
		return true
	default:
		return false
	}
}

// HostCallDispatchLogEnabled reports whether JAM_PVM_HOSTCALL_LOG was enabled
// when this process started. Callers must skip HostCallName (and the log)
// when this is false.
func HostCallDispatchLogEnabled() bool {
	return hostCallDispatchLogEnabled
}

// LogHostCallDispatchEnv logs one host-call dispatch (registers / gas before omega)
// when HostCallDispatchLogEnabled is true. Default is off so trace and normal
// runs do not flood stdout.
//
// A negative serviceID means the invocation has no service context (e.g. is_authorized / Psi_I,
// where HostCallArgs.ServiceID is nil) and is logged as "none".
func LogHostCallDispatchEnv(serviceID int64, opName string, regs any, gas int64) {
	if !hostCallDispatchLogEnabled {
		return
	}
	sid := "none"
	if serviceID >= 0 {
		sid = strconv.FormatInt(serviceID, 10)
	}
	traceLogger.Debugf("serviceID: %s, host-call: %s, regs: %v, gas: %d", sid, opName, regs, gas)
}
