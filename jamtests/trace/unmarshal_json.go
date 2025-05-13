package jamtests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// TraceState UnmarshalJSON unmarshals a JSON-encoded StateKeyVals.
func (s *StateKeyVals) UnmarshalJSON(data []byte) error {
	var raw []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*s = make(StateKeyVals, len(raw))
	for i, kv := range raw {
		// key
		decodedKey, err := hex.DecodeString(kv.Key[2:])
		if err != nil {
			return fmt.Errorf("failed to decode hex string: %w", err)
		}
		(*s)[i].Key = types.StateKey(decodedKey)

		// value
		decodedValue, err := hex.DecodeString(kv.Value[2:])
		if err != nil {
			return fmt.Errorf("failed to decode hex string: %w", err)
		}
		(*s)[i].Value = decodedValue
	}

	return nil
}

// UnmarshalJSON unmarshals a JSON-encoded TraceState.
func (s *TraceState) UnmarshalJSON(data []byte) error {
	var raw struct {
		StateRoot types.StateRoot `json:"state_root"`
		KeyVals   StateKeyVals    `json:"keyvals"`
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
