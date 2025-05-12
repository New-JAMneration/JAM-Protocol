package jamtests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type TraceKeyVal struct {
	Key   types.StateKey     `json:"key"`
	Value types.ByteSequence `json:"value"`
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

// Decode StateKeyVals
func (s *StateKeyVals) Decode(d *types.Decoder) error {
	var err error

	// Decode the length of the array
	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// Allocate space for the array
	*s = make(StateKeyVals, length)

	// Decode each element in the array
	for i := uint64(0); i < length; i++ {
		if err = (*s)[i].Key.Decode(d); err != nil {
			return err
		}

		if err = (*s)[i].Value.Decode(d); err != nil {
			return err
		}
	}

	return nil
}

// Decode TraceState
func (s *TraceState) Decode(d *types.Decoder) error {
	var err error

	if err = s.StateRoot.Decode(d); err != nil {
		return err
	}

	if err = s.KeyVals.Decode(d); err != nil {
		return err
	}

	return nil
}

// Decode TraceTestCase
func (t *TraceTestCase) Decode(d *types.Decoder) error {
	var err error

	if err = t.PreState.Decode(d); err != nil {
		return err
	}

	if err = t.Block.Decode(d); err != nil {
		return err
	}

	if err = t.PostState.Decode(d); err != nil {
		return err
	}

	return nil
}
