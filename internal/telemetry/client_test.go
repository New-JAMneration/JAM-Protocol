package telemetry

import (
	"strings"
	"testing"
	"time"
)

// New(Config{Endpoint: ""}) must return the disabled (no-op) Client. All
// subsequent Emit / EmitLazy / EmitFollowup* calls return InvalidID and
// Close is a no-op. Callers can rely on this so they never need a nil
// check on the Client interface returned by New.
func TestNew_EmptyEndpointReturnsDisabled(t *testing.T) {
	c, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, ok := c.(disabledClient); !ok {
		t.Fatalf("expected disabledClient, got %T", c)
	}
	if c.Enabled() {
		t.Errorf("disabled client must not be Enabled")
	}
	if id := c.Emit(1, []byte{1}); id != InvalidID {
		t.Errorf("disabled Emit: got id %d, want InvalidID", id)
	}
	if err := c.Close(); err != nil {
		t.Errorf("disabled Close: %v", err)
	}
}

// Negative BufferSize is a misconfiguration — fail fast at New rather than
// silently allocating a non-buffered channel that would behave unlike the
// documented BufferSize=0 default.
func TestNew_NegativeBufferSize(t *testing.T) {
	_, err := New(Config{Endpoint: "127.0.0.1:0", BufferSize: -1})
	if err == nil {
		t.Fatalf("expected error for negative BufferSize")
	}
}

// New rejects misconfigured durations: negative TailDropInterval would
// panic time.NewTicker; negative reconnect intervals would hot-loop;
// inverted Reconnect bounds would never reach the cap. Fail fast at New
// rather than letting any of those reach the running goroutines.
func TestNew_RejectsInvalidDurations(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "negative-ReconnectMin",
			cfg:  Config{Endpoint: "x:1", ReconnectMin: -1 * time.Second},
			want: "ReconnectMin",
		},
		{
			name: "negative-ReconnectMax",
			cfg:  Config{Endpoint: "x:1", ReconnectMax: -1 * time.Second},
			want: "ReconnectMax",
		},
		{
			name: "negative-CloseTimeout",
			cfg:  Config{Endpoint: "x:1", CloseTimeout: -1 * time.Second},
			want: "CloseTimeout",
		},
		{
			name: "negative-TailDropInterval",
			cfg:  Config{Endpoint: "x:1", TailDropInterval: -1 * time.Second},
			want: "TailDropInterval",
		},
		{
			name: "min-greater-than-max",
			cfg:  Config{Endpoint: "x:1", ReconnectMin: 10 * time.Second, ReconnectMax: 1 * time.Second},
			want: "ReconnectMin",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.cfg)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err %q does not mention %q", err.Error(), tc.want)
			}
		})
	}
}
