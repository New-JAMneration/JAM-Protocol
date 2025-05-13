package jamtests

import (
	"encoding/json"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// UnmarshalJSON unmarshals a JSON-encoded TraceState.
func (s *TraceState) UnmarshalJSON(data []byte) error {
	var raw struct {
		StateRoot types.StateRoot    `json:"state_root"`
		KeyVals   types.StateKeyVals `json:"keyvals"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	s.StateRoot = raw.StateRoot
	s.KeyVals = raw.KeyVals

	return nil
}

// UnmarshalJSON unmarshals a JSON-encoded TraceTestCase.
func (t *TraceTestCase) UnmarshalJSON(data []byte) error {
	var raw struct {
		PreState  TraceState  `json:"pre_state"`
		Block     types.Block `json:"block"`
		PostState TraceState  `json:"post_state"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	t.PreState = raw.PreState
	t.Block = raw.Block
	t.PostState = raw.PostState

	return nil
}
