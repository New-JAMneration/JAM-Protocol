package jamtests

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
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
	ValsCurrStats  types.ValidatorsStatistics `json:"vals_curr_stats"`
	ValsLastStats  types.ValidatorsStatistics `json:"vals_last_stats"`
	Slot           types.TimeSlot             `json:"slot"`
	CurrValidators types.ValidatorsData       `json:"curr_validators"`
}

// StatisticsState UnmarshalJSON
func (s *StatisticsState) UnmarshalJSON(data []byte) error {
	var err error

	var state struct {
		ValsCurrStats  types.ValidatorsStatistics `json:"vals_curr_stats"`
		ValsLastStats  types.ValidatorsStatistics `json:"vals_last_stats"`
		Slot           types.TimeSlot             `json:"slot"`
		CurrValidators types.ValidatorsData       `json:"curr_validators"`
	}

	if err = json.Unmarshal(data, &state); err != nil {
		return err
	}

	s.ValsCurrStats = state.ValsCurrStats
	s.ValsLastStats = state.ValsLastStats
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

	if err = s.ValsCurrStats.Decode(d); err != nil {
		return err
	}

	if err = s.ValsLastStats.Decode(d); err != nil {
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

	if err = s.ValsCurrStats.Encode(e); err != nil {
		return err
	}

	if err = s.ValsLastStats.Encode(e); err != nil {
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

func (s *StatisticsTestCase) Dump() error {
	cs := blockchain.GetInstance()

	// Input
	block := types.Block{
		Header: types.Header{
			Slot:        s.Input.Slot,
			AuthorIndex: s.Input.AuthorIndex,
		},
		Extrinsic: s.Input.Extrinsic,
	}
	cs.AddBlock(block)

	// w (present work reports) from guarantee extrinsic
	reports := []types.WorkReport{}
	for _, guarantee := range s.Input.Extrinsic.Guarantees {
		reports = append(reports, guarantee.Report)
	}
	cs.GetIntermediateStates().SetPresentWorkReports(reports)

	// PreState
	cs.GetPriorStates().SetTau(s.PreState.Slot)
	cs.GetPriorStates().SetPiCurrent(s.PreState.ValsCurrStats)
	cs.GetPriorStates().SetPiLast(s.PreState.ValsLastStats)

	// PostState
	cs.GetPosteriorStates().SetTau(s.Input.Slot)
	cs.GetPosteriorStates().SetKappa(s.PreState.CurrValidators)
	return nil
}

func (s *StatisticsTestCase) GetPostState() interface{} {
	return s.PostState
}

func (s *StatisticsTestCase) GetOutput() interface{} {
	return s.Output
}

func (s *StatisticsTestCase) ExpectError() error {
	return nil
}

func (s *StatisticsTestCase) Validate() error {
	cs := blockchain.GetInstance()

	statistics := cs.GetPosteriorStates().GetPi()

	if !reflect.DeepEqual(statistics.ValsCurr, s.PostState.ValsCurrStats) {
		return fmt.Errorf("statistics.ValsCurrent failed: expected %v, got %v", s.PostState.ValsCurrStats, statistics.ValsCurr)
	}

	if !reflect.DeepEqual(statistics.ValsLast, s.PostState.ValsLastStats) {
		return fmt.Errorf("statistics.ValsLast failed: expected %v, got %v", s.PostState.ValsLastStats, statistics.ValsLast)
	}

	expectedCurrentValidators := s.PostState.CurrValidators
	actualCurrentValidators := cs.GetPosteriorStates().GetKappa()
	if !reflect.DeepEqual(actualCurrentValidators, expectedCurrentValidators) {
		return fmt.Errorf("CurrentValidators failed: expected %v, got %v", expectedCurrentValidators, actualCurrentValidators)
	}

	// https://github.com/davxy/jam-test-vectors/issues/91
	// Davxy:
	// Slot should remain unchanged in the state.
	// Statistics is not supposed to be (at least in our vectors proposal) the STF subsystem that changes the slot in the state.

	return nil
}
