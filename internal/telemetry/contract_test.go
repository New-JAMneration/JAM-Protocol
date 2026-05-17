package telemetry

import (
	"context"
	"testing"
	"time"
)

func TestContract_RegisterAddsEntry(t *testing.T) {
	t.Cleanup(clearRegistry)
	clearRegistry()

	Register(BridgeContract{
		Name:    "test.event10",
		Disc:    10,
		Handler: func(context.Context, any) error { return nil },
		Sample:  func() any { return 42 },
	})

	got := RegisteredBridges()
	if len(got) != 1 {
		t.Fatalf("got %d entries, want 1", len(got))
	}
	if got[0].Name != "test.event10" || got[0].Disc != 10 {
		t.Errorf("unexpected entry: %+v", got[0])
	}
}

func TestContract_RegisteredBridgesReturnsCopy(t *testing.T) {
	t.Cleanup(clearRegistry)
	clearRegistry()

	Register(BridgeContract{Name: "a"})
	Register(BridgeContract{Name: "b"})

	snap1 := RegisteredBridges()
	// Mutating the returned slice must not affect the registry.
	snap1[0].Name = "mutated"

	snap2 := RegisteredBridges()
	if snap2[0].Name != "a" {
		t.Errorf("registry mutated through returned slice: got %q, want %q",
			snap2[0].Name, "a")
	}
}

func TestContract_ClearRegistryEmpties(t *testing.T) {
	t.Cleanup(clearRegistry)
	clearRegistry()

	Register(BridgeContract{Name: "x"})
	if got := len(RegisteredBridges()); got != 1 {
		t.Fatalf("setup: got %d, want 1", got)
	}

	clearRegistry()
	if got := len(RegisteredBridges()); got != 0 {
		t.Errorf("after clearRegistry: got %d, want 0", got)
	}
}

// Driver-style: walk every registered bridge and feed Sample() through
// Handler. The handler MUST return nil and MUST NOT block — exactly the
// three Bridge invariants. Domain PRs add Sample factories so their
// bridges get exercised here without each PR duplicating the test
// boilerplate.
//
// Each handler runs in its own goroutine bounded by handlerTimeout so a
// blocking bridge fails its sub-test (locatable) rather than hanging
// the whole test binary.
func TestContract_AllRegisteredBridgesUpholdInvariants(t *testing.T) {
	t.Cleanup(clearRegistry)
	clearRegistry()

	// Synthetic registrations so the driver has something to chew on.
	// Real usage: domain PRs register their actual bridges in init().
	fc := newFakeClient(true)
	Register(BridgeContract{
		Name:    "synthetic.eager",
		Disc:    1,
		Handler: Bridge(fc, 1, func(ev any) []byte { return []byte{0x01} }),
		Sample:  func() any { return 0 },
	})
	Register(BridgeContract{
		Name: "synthetic.lazy",
		Disc: 2,
		// snapshot copies by value. `return ev` happens to be safe here
		// because Sample returns an int (value type, no aliasing), but
		// real domain bridges where ev is a pointer MUST copy fields
		// out explicitly — see the bad/good examples in bridge.go's
		// BridgeLazy doc.
		Handler: BridgeLazy(fc, 2,
			func(ev any) any { v, _ := ev.(int); return v },
			func(snap any) []byte { return []byte{0x02} },
		),
		Sample: func() any { return 0 },
	})

	const handlerTimeout = 500 * time.Millisecond
	for _, c := range RegisteredBridges() {
		t.Run(c.Name, func(t *testing.T) {
			ev := c.Sample()
			done := make(chan error, 1)
			go func() {
				done <- c.Handler(context.Background(), ev)
			}()
			select {
			case err := <-done:
				if err != nil {
					t.Errorf("handler returned err: %v (must always be nil)", err)
				}
			case <-time.After(handlerTimeout):
				t.Fatalf("handler did not return within %s (non-blocking invariant violated)", handlerTimeout)
			}
		})
	}
}
