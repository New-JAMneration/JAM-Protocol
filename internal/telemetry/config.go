package telemetry

import (
	"fmt"
	"time"
)

// Config controls the telemetry client. Empty Endpoint → New returns
// the disabled (no-op) client and the rest is ignored. Endpoint format
// is "host:port" (anything net.Dial accepts).
type Config struct {
	// Endpoint is the JamTART (or compatible aggregator) "host:port".
	// Empty disables telemetry; New returns the no-op client.
	Endpoint string

	// NodeInfo is the connection-initial message sent on every (re)connect.
	// Encode runs on each connection so an invalid NodeInfo (oversized
	// String<N>, undefined NodeFlags bits, bad UTF-8) becomes a connect
	// error: writer logs and reconnects.
	NodeInfo NodeInfo

	// BufferSize bounds the buffered envelope channel. Producers that
	// can't enqueue record a drop range instead of blocking. 0 = default
	// (4096); negative = error from New.
	BufferSize int

	// ReconnectMin is the initial backoff after a failed dial or dropped
	// connection. Doubles up to ReconnectMax. 0 = default (1s).
	ReconnectMin time.Duration

	// ReconnectMax caps the exponential backoff. 0 = default (30s).
	ReconnectMax time.Duration

	// CloseTimeout bounds Close: drain + flush + FIN. On timeout, dial /
	// writer / reader are force-closed and remaining envelopes logged
	// as discarded. 0 = default (5s).
	CloseTimeout time.Duration

	// TailDropInterval is how often the writer wakes when the channel is
	// empty to flush tail drops (drops with no following event to
	// trigger a flush). 0 = default (1s).
	TailDropInterval time.Duration
}

// Defaults applied by withDefaults when a field is zero.
const (
	defaultBufferSize       = 4096
	defaultReconnectMin     = 1 * time.Second
	defaultReconnectMax     = 30 * time.Second
	defaultCloseTimeout     = 5 * time.Second
	defaultTailDropInterval = 1 * time.Second
)

// validate rejects values that would panic time.NewTicker (negative
// TailDropInterval), hot-loop retries (negative reconnect intervals), or
// skip Close's timeout entirely (negative CloseTimeout). An explicit
// negative is clearly pathological; fail fast at New.
func (cfg Config) validate() error {
	if cfg.BufferSize < 0 {
		return fmt.Errorf("telemetry: BufferSize must be >= 0, got %d", cfg.BufferSize)
	}
	if cfg.ReconnectMin < 0 {
		return fmt.Errorf("telemetry: ReconnectMin must be >= 0, got %v", cfg.ReconnectMin)
	}
	if cfg.ReconnectMax < 0 {
		return fmt.Errorf("telemetry: ReconnectMax must be >= 0, got %v", cfg.ReconnectMax)
	}
	if cfg.ReconnectMin > 0 && cfg.ReconnectMax > 0 && cfg.ReconnectMin > cfg.ReconnectMax {
		return fmt.Errorf(
			"telemetry: ReconnectMin (%v) must be <= ReconnectMax (%v)",
			cfg.ReconnectMin, cfg.ReconnectMax,
		)
	}
	if cfg.CloseTimeout < 0 {
		return fmt.Errorf("telemetry: CloseTimeout must be >= 0, got %v", cfg.CloseTimeout)
	}
	if cfg.TailDropInterval < 0 {
		return fmt.Errorf("telemetry: TailDropInterval must be >= 0, got %v", cfg.TailDropInterval)
	}
	return nil
}

// withDefaults fills in zero-valued fields from the package defaults.
// Endpoint and NodeInfo are not touched.
func (cfg Config) withDefaults() Config {
	if cfg.BufferSize == 0 {
		cfg.BufferSize = defaultBufferSize
	}
	if cfg.ReconnectMin == 0 {
		cfg.ReconnectMin = defaultReconnectMin
	}
	if cfg.ReconnectMax == 0 {
		cfg.ReconnectMax = defaultReconnectMax
	}
	if cfg.CloseTimeout == 0 {
		cfg.CloseTimeout = defaultCloseTimeout
	}
	if cfg.TailDropInterval == 0 {
		cfg.TailDropInterval = defaultTailDropInterval
	}
	return cfg
}
