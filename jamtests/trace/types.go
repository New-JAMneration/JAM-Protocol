package jamtests

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type TraceState struct {
	StateRoot types.StateRoot
	KeyVals   types.StateKeyVals
}

type TraceTestCase struct {
	PreState  TraceState
	Block     types.Block
	PostState TraceState
}
