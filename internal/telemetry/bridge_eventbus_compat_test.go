package telemetry

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
)

// Compile-time guards: the values returned by Bridge / BridgeLazy /
// BridgeFollowup must be assignable to quic.Handler without explicit
// conversion. This requires telemetry.Handler to be a type alias for
// the function literal (so it stays unnamed in the type-system sense)
// — a named type with the same underlying signature would force every
// callsite to write quic.Handler(bridge.Handler(...)).
//
// If quic.EventBus's signature drifts (e.g. handler return type
// changes), one of these variable declarations stops compiling and CI
// catches it before any callsite tries to subscribe.
var (
	_ quic.Handler = Bridge(
		NewDisabled(),
		0,
		func(any) []byte { return nil },
	)
	_ quic.Handler = BridgeLazy(
		NewDisabled(),
		0,
		func(any) any { return nil },
		func(any) []byte { return nil },
	)
	_ quic.Handler = BridgeFollowup(
		NewDisabled(),
		0,
		func(any) (uint64, []byte) { return 0, nil },
	)
)
