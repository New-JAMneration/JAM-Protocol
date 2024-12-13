package work_report

import (
	"fmt"
	"github.com/New-JAMneration/JAM-Protocol/internal/work_result"
)

type RefineContext struct {
	Anchor           [32]byte // HeaderHash
	StateRoot        [32]byte
	BeefyRoot        [32]byte
	LookupAnchor     [32]byte // HeaderHash
	LookupAnchorSlot uint32   // TimeSlot
	Prerequisites    [][32]byte
}

type WorkPackageSpec struct {
	Hash         [32]byte `json:"hash,omitempty"` // WorkPackageHash
	Length       uint32   `json:"length,omitempty"`
	ErasureRoot  [32]byte `json:"erasure_root,omitempty"`
	ExportsRoot  [32]byte `json:"exports_root,omitempty"`
	ExportsCount uint16   `json:"exports_count,omitempty"`
}

type SegmentRootLookupItem struct {
	WorkPackageHash [32]byte `json:"work_package_hash,omitempty"`
	SegmentTreeRoot [32]byte `json:"segment_tree_root,omitempty"`
}

type SegmentRootLookup []SegmentRootLookupItem

type WorkReport struct {
	PackageSpec       WorkPackageSpec          `json:"package_spec"`
	Context           RefineContext            `json:"context"`
	CoreIndex         uint16                   `json:"core_index,omitempty"`
	AuthorizerHash    [32]byte                 `json:"authorizer_hash,omitempty"`
	AuthOutput        []byte                   `json:"auth_output,omitempty"`
	SegmentRootLookup SegmentRootLookup        `json:"segment_root_lookup,omitempty"`
	Results           []work_result.WorkResult `json:"results,omitempty"`
}

func (w WorkReport) Validate() error {
	if len(w.Results) < 1 || len(w.Results) > 4 {
		return fmt.Errorf("WorkReport Results must have between 1 and 4 items, but got %d", len(w.Results))
	}
	return nil
}
