package PVMtrace

import (
	"os"
	"strconv"
)

// TraceConfig holds runtime configuration for trace output.
type TraceConfig struct {
	Dir        string // JAM_PVM_TRACE_DIR
	RunID      string // JAM_PVM_TRACE_RUN_ID
	Streams    string // JAM_PVM_TRACE_STREAMS ("all" or comma-separated)
	BufferMB   int    // JAM_PVM_TRACE_BUFFER_MB
	MaxSteps   int64  // JAM_PVM_TRACE_MAX_STEPS (0 = unlimited)
	TotalMB    int64  // JAM_PVM_TRACE_TOTAL_MB (0 = unlimited)
	GzipLevel  int    // JAM_PVM_TRACE_GZIP_LEVEL (1-9, default 6)
}

const (
	envTraceDir      = "JAM_PVM_TRACE_DIR"
	envTraceRunID    = "JAM_PVM_TRACE_RUN_ID"
	envTraceStreams  = "JAM_PVM_TRACE_STREAMS"
	envTraceBufferMB = "JAM_PVM_TRACE_BUFFER_MB"
	envTraceMaxSteps = "JAM_PVM_TRACE_MAX_STEPS"
	envTraceTotalMB  = "JAM_PVM_TRACE_TOTAL_MB"
	envTraceGzipLvl  = "JAM_PVM_TRACE_GZIP_LEVEL"
)

// LoadConfigFromEnv reads trace configuration from environment variables.
func LoadConfigFromEnv() TraceConfig {
	cfg := TraceConfig{
		Dir:       os.Getenv(envTraceDir),
		RunID:     os.Getenv(envTraceRunID),
		Streams:   os.Getenv(envTraceStreams),
		GzipLevel: 6,
	}
	if cfg.Streams == "" {
		cfg.Streams = "all"
	}
	if v := os.Getenv(envTraceBufferMB); v != "" {
		cfg.BufferMB, _ = strconv.Atoi(v)
	}
	if cfg.BufferMB <= 0 {
		cfg.BufferMB = 4
	}
	if v := os.Getenv(envTraceMaxSteps); v != "" {
		cfg.MaxSteps, _ = strconv.ParseInt(v, 10, 64)
	}
	if v := os.Getenv(envTraceTotalMB); v != "" {
		cfg.TotalMB, _ = strconv.ParseInt(v, 10, 64)
	}
	if v := os.Getenv(envTraceGzipLvl); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 9 {
			cfg.GzipLevel = n
		}
	}
	return cfg
}
