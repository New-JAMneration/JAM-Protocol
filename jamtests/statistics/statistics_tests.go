package jamtests

import (
	"encoding/json"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type StatisticsTestCase struct {
	Input     StatisticsInput  `json:"input"`
	PreState  StatisticsState  `json:"pre_state"`
	Output    StatisticsOutput `json:"output"`
	PostState StatisticsState  `json:"post_state"`
}

type StatisticsInput struct {
	Slot        types.TimeSlot       `json:"slot"`
	AuthorIndex types.ValidatorIndex `json:"author_index"`
	Extrinsic   types.Extrinsic      `json:"extrinsic"`
}

type StatisticsOutput struct { // null
}

type StatisticsState struct {
	Statistics     types.Statistics     `json:"statistics"`
	Slot           types.TimeSlot       `json:"slot"`
	CurrValidators types.ValidatorsData `json:"curr_validators"`
}

// StatisticsState UnmarshalJSON
func (s *StatisticsState) UnmarshalJSON(data []byte) error {
	var err error

	var state struct {
		Statistics     types.Statistics     `json:"statistics"`
		Slot           types.TimeSlot       `json:"slot"`
		CurrValidators types.ValidatorsData `json:"curr_validators"`
	}

	if err = json.Unmarshal(data, &state); err != nil {
		return err
	}

	s.Statistics = state.Statistics
	s.Slot = state.Slot
	s.CurrValidators = state.CurrValidators

	return nil
}

func (s *StatisticsTestCase) UnmarshalJSON(data []byte) error {
	var err error

	var testCase struct {
		Input     StatisticsInput  `json:"input"`
		PreState  StatisticsState  `json:"pre_state"`
		Output    StatisticsOutput `json:"output"`
		PostState StatisticsState  `json:"post_state"`
	}

	if err = json.Unmarshal(data, &testCase); err != nil {
		return err
	}

	s.Input = testCase.Input
	s.PreState = testCase.PreState
	s.Output = testCase.Output
	s.PostState = testCase.PostState

	return nil
}

// StatisticsTestCase
func (t *StatisticsTestCase) Decode(d *types.Decoder) error {
	var err error

	if err = t.Input.Decode(d); err != nil {
		return err
	}

	if err = t.PreState.Decode(d); err != nil {
		return err
	}

	if err = t.Output.Decode(d); err != nil {
		return err
	}

	if err = t.PostState.Decode(d); err != nil {
		return err
	}

	return nil
}

// StatisticsInput
func (i *StatisticsInput) Decode(d *types.Decoder) error {
	var err error

	if err = i.Slot.Decode(d); err != nil {
		return err
	}

	if err = i.AuthorIndex.Decode(d); err != nil {
		return err
	}

	if err = i.Extrinsic.Decode(d); err != nil {
		return err
	}

	return nil
}

// StatisticsOutput
func (o *StatisticsOutput) Decode(d *types.Decoder) error {
	return nil
}

// StatisticsState
func (s *StatisticsState) Decode(d *types.Decoder) error {
	var err error

	if err = s.Statistics.Decode(d); err != nil {
		return err
	}

	if err = s.Slot.Decode(d); err != nil {
		return err
	}

	if err = s.CurrValidators.Decode(d); err != nil {
		return err
	}

	return nil
}

type Decodable interface {
	Decode(d *types.Decoder) error
}

// Encode
type Encodable interface {
	Encode(e *types.Encoder) error
}

// StatisticsInput
func (i *StatisticsInput) Encode(e *types.Encoder) error {
	var err error

	if err = i.Slot.Encode(e); err != nil {
		return err
	}

	if err = i.AuthorIndex.Encode(e); err != nil {
		return err
	}

	if err = i.Extrinsic.Encode(e); err != nil {
		return err
	}

	return nil
}

// StatisticsOutput
func (o *StatisticsOutput) Encode(e *types.Encoder) error {
	return nil
}

// StatisitcsState
func (s *StatisticsState) Encode(e *types.Encoder) error {
	var err error

	if err = s.Statistics.Encode(e); err != nil {
		return err
	}

	if err = s.Slot.Encode(e); err != nil {
		return err
	}

	if err = s.CurrValidators.Encode(e); err != nil {
		return err
	}

	return nil
}

// StatisticsTestCase
func (t *StatisticsTestCase) Encode(e *types.Encoder) error {
	var err error

	if err = t.Input.Encode(e); err != nil {
		return err
	}

	if err = t.PreState.Encode(e); err != nil {
		return err
	}

	if err = t.Output.Encode(e); err != nil {
		return err
	}

	if err = t.PostState.Encode(e); err != nil {
		return err
	}

	return nil
}

// TODO: Implement Dump method
func (s *StatisticsTestCase) Dump() error {
	return nil
}
