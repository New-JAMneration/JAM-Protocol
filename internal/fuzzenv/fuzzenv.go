package fuzzenv

import (
	"os"
	"strings"
)

const EnvKey = "JAM_FUZZ"

// FuzzPersistentRetainBlocks is the maximum number of blocks retained on disk
// when running as a fuzz target. Independent of protocol L (MaxLookupAge) to
// prevent disk exhaustion in constrained containers regardless of spec.
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
