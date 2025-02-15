package jamtests

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type HistoryTestCase struct {
	Input     HistoryInput  `json:"input"`
	PreState  HistoryState  `json:"pre_state"`
	Output    HistoryOutput `json:"output"`
	PostState HistoryState  `json:"post_state"`
}

type HistoryInput struct {
	HeaderHash      types.HeaderHash            `json:"header_hash"`
	ParentStateRoot types.StateRoot             `json:"parent_state_root"`
	AccumulateRoot  types.OpaqueHash            `json:"accumulate_root"`
	WorkPackages    []types.ReportedWorkPackage `json:"work_packages"`
}

type HistoryOutput struct { // null
}

type HistoryState struct {
	Beta types.BlocksHistory `json:"beta"`
}
