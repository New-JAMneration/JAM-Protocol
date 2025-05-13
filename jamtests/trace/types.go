package jamtests

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type TraceKeyVal struct {
	Key   types.StateKey
	Value types.ByteSequence
}

type StateKeyVals []TraceKeyVal

type TraceState struct {
	StateRoot types.StateRoot
	KeyVals   StateKeyVals
}

type TraceTestCase struct {
	PreState  TraceState
	Block     types.Block
	PostState TraceState
}
