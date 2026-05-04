package telemetry

import (
	"testing"
	"time"
)

func TestConfig_withDefaults(t *testing.T) {
	t.Run("ZeroValueGetsDefaults", func(t *testing.T) {
		got := Config{}.withDefaults()
		if got.BufferSize != defaultBufferSize {
			t.Errorf("BufferSize: got %d want %d", got.BufferSize, defaultBufferSize)
		}
		if got.ReconnectMin != defaultReconnectMin {
			t.Errorf("ReconnectMin: got %v want %v", got.ReconnectMin, defaultReconnectMin)
		}
		if got.ReconnectMax != defaultReconnectMax {
			t.Errorf("ReconnectMax: got %v want %v", got.ReconnectMax, defaultReconnectMax)
		}
		if got.CloseTimeout != defaultCloseTimeout {
			t.Errorf("CloseTimeout: got %v want %v", got.CloseTimeout, defaultCloseTimeout)
		}
		if got.TailDropInterval != defaultTailDropInterval {
			t.Errorf("TailDropInterval: got %v want %v", got.TailDropInterval, defaultTailDropInterval)
		}
	})

	t.Run("NonZeroValuesPreserved", func(t *testing.T) {
		in := Config{
			BufferSize:       16,
			ReconnectMin:     250 * time.Millisecond,
			ReconnectMax:     5 * time.Second,
			CloseTimeout:     2 * time.Second,
			TailDropInterval: 100 * time.Millisecond,
		}
		got := in.withDefaults()
		if got.BufferSize != 16 {
			t.Errorf("BufferSize overridden: got %d", got.BufferSize)
		}
		if got.ReconnectMin != 250*time.Millisecond {
			t.Errorf("ReconnectMin overridden: got %v", got.ReconnectMin)
		}
		if got.ReconnectMax != 5*time.Second {
			t.Errorf("ReconnectMax overridden: got %v", got.ReconnectMax)
		}
		if got.CloseTimeout != 2*time.Second {
			t.Errorf("CloseTimeout overridden: got %v", got.CloseTimeout)
		}
		if got.TailDropInterval != 100*time.Millisecond {
			t.Errorf("TailDropInterval overridden: got %v", got.TailDropInterval)
		}
	})

	t.Run("EndpointAndNodeInfoNotTouched", func(t *testing.T) {
		in := Config{Endpoint: "host:123", NodeInfo: NodeInfo{ImplName: "x"}}
		got := in.withDefaults()
		if got.Endpoint != "host:123" {
			t.Errorf("Endpoint changed: %q", got.Endpoint)
		}
		if got.NodeInfo.ImplName != "x" {
			t.Errorf("NodeInfo.ImplName changed: %q", got.NodeInfo.ImplName)
		}
	})
}
