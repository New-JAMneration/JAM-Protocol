package jamtests

import (
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
	Pi    types.Statistics     `json:"pi"`
	Tau   types.TimeSlot       `json:"tau"`
	Kappa types.ValidatorsData `json:"kappa_prime"`
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

	if err = s.Pi.Decode(d); err != nil {
		return err
	}

	if err = s.Tau.Decode(d); err != nil {
		return err
	}

	if err = s.Kappa.Decode(d); err != nil {
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

	if err = s.Pi.Encode(e); err != nil {
		return err
	}

	if err = s.Tau.Encode(e); err != nil {
		return err
	}

	if err = s.Kappa.Encode(e); err != nil {
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
