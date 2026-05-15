package telemetry

import "sync"

// BridgeContract describes a registered bridge handler so the contract
// test suite (L5 nightly) can validate it satisfies the three Bridge
// invariants without each domain PR re-writing the boilerplate.
//
// Domain PRs register their bridges via init():
//
//	func init() {
//	    telemetry.Register(telemetry.BridgeContract{
//	        Name:    "status.event10",
//	        Disc:    10,
//	        Handler: statusEvent10Bridge(tel),
//	        Sample:  func() any { return &StatusEvent10{...} },
//	    })
//	}
//
// The L5 driver iterates RegisteredBridges and asserts that each
// Handler swallows panics, returns nil within a bounded time, and
// doesn't block the caller.
type BridgeContract struct {
	// Name is a stable identifier for diagnostics (e.g. "status.event10").
	Name string

	// Disc is the JIP-3 discriminator the bridge emits.
	Disc uint8

	// Handler is the Bridge / BridgeLazy / BridgeFollowup output.
	Handler Handler

	// Sample produces a representative event the contract driver feeds
	// to Handler. A factory (not a value) so the driver can rerun the
	// handler against a fresh event without prior runs' mutations.
	Sample func() any
}

var (
	registryMu sync.Mutex
	registry   []BridgeContract
)

// Register adds a bridge to the contract registry. Safe to call from
// init() concurrently across packages.
func Register(c BridgeContract) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = append(registry, c)
}

// RegisteredBridges returns a copy of the registry. The copy keeps the
// test driver from racing with concurrent Register calls (Register is
// init-time only in production, but tests may add entries dynamically).
func RegisteredBridges() []BridgeContract {
	registryMu.Lock()
	defer registryMu.Unlock()
	out := make([]BridgeContract, len(registry))
	copy(out, registry)
	return out
}
