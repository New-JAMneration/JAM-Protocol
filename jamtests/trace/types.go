package jamtests

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// RawState
type TraceState struct {
	StateRoot types.StateRoot
	KeyVals   types.StateKeyVals
}

// TraceStep
type TraceTestCase struct {
	PreState  TraceState
	Block     types.Block
	PostState TraceState
}

type Genesis struct {
	Header types.Header
	State  TraceState
}
