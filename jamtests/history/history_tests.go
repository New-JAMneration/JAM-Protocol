package jamtests

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// ANSI color codes
var (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Gray    = "\033[37m"
	White   = "\033[97m"
)

var debugMode = false

// var debugMode = true

func cLog(color string, string string) {
	if debugMode {
		fmt.Printf("%s%s%s\n", color, string, Reset)
	}
}

type HistoryTestCase struct {
	Input     HistoryInput  `json:"input"`
	PreState  HistoryState  `json:"pre_state"`
	Output    HistoryOutput `json:"output"`
	PostState HistoryState  `json:"post_state"`
}

type HistoryInput struct {
	HeaderHash      types.HeaderHash            `json:"header_hash"`
	ParentStateRoot types.StateRoot             `json:"parent_state_root"`
	AccumulateRoot  types.OpaqueHash            `json:"accumulate_root"`
	WorkPackages    []types.ReportedWorkPackage `json:"work_packages"`
}

type HistoryOutput struct { // null
}

type HistoryState struct {
	Beta types.BlocksHistory `json:"beta"`
}

// HistoryInput
func (h *HistoryInput) Decode(d *types.Decoder) error {
	var err error

	if err = h.HeaderHash.Decode(d); err != nil {
		return err
	}

	if err = h.ParentStateRoot.Decode(d); err != nil {
		return err
	}

	if err = h.AccumulateRoot.Decode(d); err != nil {
		return err
	}

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length != 0 {
		h.WorkPackages = make([]types.ReportedWorkPackage, length)
		for i := range h.WorkPackages {
			if err = h.WorkPackages[i].Decode(d); err != nil {
				return err
			}
		}
	}

	return nil
}

// HistoryOutput
func (h *HistoryOutput) Decode(d *types.Decoder) error {
	return nil
}

// HistoryState
func (h *HistoryState) Decode(d *types.Decoder) error {
	cLog(Blue, "Decoding HistoryState")

	var err error

	if err = h.Beta.Decode(d); err != nil {
		return err
	}

	return nil
}

// HistoryTestCase Decode
func (h *HistoryTestCase) Decode(d *types.Decoder) error {
	var err error

	if err = h.Input.Decode(d); err != nil {
		return err
	}

	if err = h.PreState.Decode(d); err != nil {
		return err
	}

	if err = h.Output.Decode(d); err != nil {
		return err
	}

	if err = h.PostState.Decode(d); err != nil {
		return err
	}

	return nil
}

// Encode
type Encodable interface {
	Encode(e *types.Encoder) error
}

// HistoryInput
func (h *HistoryInput) Encode(e *types.Encoder) error {
	var err error

	if err = h.HeaderHash.Encode(e); err != nil {
		return err
	}

	if err = h.ParentStateRoot.Encode(e); err != nil {
		return err
	}

	if err = h.AccumulateRoot.Encode(e); err != nil {
		return err
	}

	if err = e.EncodeLength(uint64(len(h.WorkPackages))); err != nil {
		return err
	}

	for i := range h.WorkPackages {
		if err = h.WorkPackages[i].Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// HistoryOutput
func (h *HistoryOutput) Encode(e *types.Encoder) error {
	return nil
}

// HistoryState
func (h *HistoryState) Encode(e *types.Encoder) error {
	var err error

	if err = h.Beta.Encode(e); err != nil {
		return err
	}

	return nil
}

// HistoryTestCase
func (h *HistoryTestCase) Encode(e *types.Encoder) error {
	var err error

	if err = h.Input.Encode(e); err != nil {
		return err
	}

	if err = h.PreState.Encode(e); err != nil {
		return err
	}

	if err = h.Output.Encode(e); err != nil {
		return err
	}

	if err = h.PostState.Encode(e); err != nil {
		return err
	}

	return nil
}
