package jamtests

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type TraceState struct {
	StateRoot types.StateRoot
	// TODO: Add other fields as needed
}

type TraceTestCase struct {
	PreState  TraceState
	Block     types.Block
	PostState TraceState
}
