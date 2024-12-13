package work

// refine context (11.4)
type RefineContext struct {
	Anchor           [32]byte   // anchor header hash
	StateRoot        [32]byte   // posterior state root
	BeefyRoot        [32]byte   // posterior beefy root
	LookupAnchor     [32]byte   // lookup anchor header hash
	LookupAnchorSlot uint32     // lookup anchor time slot
	Prerequisites    [][32]byte // hash of prerequisite work packages
}
