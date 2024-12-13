package work

// refine context (11.4)
type RefineContext struct {
	Anchor           [32]byte   `json:"anchor"`             // anchor header hash
	StateRoot        [32]byte   `json:"state_root"`         // posterior state root
	BeefyRoot        [32]byte   `json:"beefy_root"`         // posterior beefy root
	LookupAnchor     [32]byte   `json:"lookup_anchor"`      // lookup anchor header hash
	LookupAnchorSlot uint32     `json:"lookup_anchor_slot"` // lookup anchor time slot
	Prerequisites    [][32]byte `json:"prerequisites"`      // hash of prerequisite work packages
}
