package validator

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

// MergeValidators deduplicates validators from multiple epoch sets by Ed25519 public key,
// preserving first-seen order. Used during epoch transitions to keep transport peer lists stable.
func MergeValidators(sets ...types.ValidatorsData) []types.Validator {
	seen := make(map[types.Ed25519Public]struct{})
	out := make([]types.Validator, 0)
	for _, set := range sets {
		for _, v := range set {
			if _, ok := seen[v.Ed25519]; ok {
				continue
			}
			seen[v.Ed25519] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}

// TransportTargets returns validators this node must maintain QUIC transport connectivity to
// per JAMNP required connectivity (all unique validators in previous, current, and next epochs).
func TransportTargets(grid *GridMapper, selfKey types.Ed25519Public) []types.Validator {
	if grid == nil {
		return nil
	}
	merged := MergeValidators(grid.Previous, grid.Current, grid.Next)
	out := make([]types.Validator, 0, len(merged))
	for _, v := range merged {
		if v.Ed25519 == selfKey {
			continue
		}
		out = append(out, v)
	}
	return out
}

// GossipPeerKeys returns Ed25519 keys eligible for grid gossip streams (UP 0 / preimage).
// This is a subset of transport connectivity: only grid-gossip peers, not the full merged set.
func GossipPeerKeys(vm *ValidatorManager) []types.Ed25519Public {
	if vm == nil || vm.Grid == nil || vm.SelfIndex < 0 {
		return nil
	}
	neighbors := vm.GetNeighbors()
	keys := make([]types.Ed25519Public, 0, len(neighbors))
	for _, v := range neighbors {
		if v.Ed25519 == vm.SelfKey {
			continue
		}
		keys = append(keys, v.Ed25519)
	}
	return keys
}
