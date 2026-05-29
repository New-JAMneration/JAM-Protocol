package fuzzenv

import (
	"os"
	"strings"
)

const EnvKey = "JAM_FUZZ"

// FuzzPersistentRetainBlocks caps in-memory block/state history for the fuzz
// target (unfinalized chain + restore window). Independent of protocol L
// (MaxLookupAge) to prevent OOM in constrained containers.
const FuzzPersistentRetainBlocks = 24

// Enabled reports whether the process is running as a JAM fuzz target.
func Enabled() bool {
	v := strings.TrimSpace(os.Getenv(EnvKey))
	if v == "" {
		return false
	}
	switch strings.ToLower(v) {
	case "0", "false", "no", "off":
		return false
	}
	return true
}
