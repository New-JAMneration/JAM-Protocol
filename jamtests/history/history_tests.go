package jamtests

import (
	"fmt"
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/recent_history"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/google/go-cmp/cmp"
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
	Beta types.RecentBlocks `json:"beta"`
}

type HistoryErrorCode types.ErrorCode

func (h *HistoryErrorCode) Error() string {
	if h == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", *h)
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

func (h *HistoryTestCase) Dump() error {
	blockchain.ResetInstance()
	storeInstance := blockchain.GetInstance()

	storeInstance.GetPriorStates().SetBeta(h.PreState.Beta)

	mockGuarantessExtrinsic := types.GuaranteesExtrinsic{}
	for _, workPackage := range h.Input.WorkPackages {
		mockGuarantessExtrinsic = append(mockGuarantessExtrinsic, types.ReportGuarantee{
			Report: types.WorkReport{
				PackageSpec: types.WorkPackageSpec{
					Hash:        types.WorkPackageHash(workPackage.Hash),
					ExportsRoot: workPackage.ExportsRoot,
				},
			},
		})
	}

	// Set extrinsic
	block := types.Block{
		Header: types.Header{
			// We set this block's HeaderHash to parent
			// for the sake of using this value in STF
			Parent:          h.Input.HeaderHash,
			ParentStateRoot: h.Input.ParentStateRoot,
		},
		Extrinsic: types.Extrinsic{
			Guarantees: mockGuarantessExtrinsic,
		},
	}
	storeInstance.AddBlock(block)

	// We do part of STF here, due to the .asn file giving the intermediate value
	beefyBelt := h.PreState.Beta.Mmr
	merkleRoot := h.Input.AccumulateRoot
	beefyBeltPrime, commitment := recent_history.AppendAndCommitMmr(beefyBelt, merkleRoot)
	storeInstance.GetIntermediateStates().SetMmrCommitment(commitment)
	storeInstance.GetPosteriorStates().SetBetaB(beefyBeltPrime)

	return nil
}

func (h *HistoryTestCase) GetPostState() interface{} {
	return h.PostState
}

func (h *HistoryTestCase) GetOutput() interface{} {
	return h.Output
}

func (h *HistoryTestCase) ExpectError() error {
	// HistoryTestCase does not expect error
	return nil
}

func (h *HistoryTestCase) Validate() error {
	s := blockchain.GetInstance()
	// === (4.6) ===
	// Get result of BetaDagger from store
	HistoryDagger := s.GetIntermediateStates().GetBetaHDagger()

	// length of BetaDagger should not exceed maxBlocksHistory
	if err := HistoryDagger.Validate(); err != nil {
		return err
	}

	// === (4.7) ===
	// Get result of (7.4), beta^prime, from store
	betaPrime := s.GetPosteriorStates().GetBeta()
	if err := betaPrime.History.Validate(); err != nil {
		return err
	} else if len(betaPrime.History) < 1 {
		return fmt.Errorf("BetaPrime should not be nil, got %d", len(betaPrime.History))
	}

	// Validate output state
	if !cmp.Equal(betaPrime.History, h.PostState.Beta.History) {
		diff := cmp.Diff(h.PostState.Beta.History, betaPrime.History)
		log.Printf("BetaPrime.History: %+#v", betaPrime.History[len(betaPrime.History)-1].BeefyRoot)
		log.Printf("PostState.Beta.History: %+#v", h.PostState.Beta.History[len(h.PostState.Beta.History)-1].BeefyRoot)
		return fmt.Errorf("BetaPrime.History should equal to PostState.Beta.History\n%s", diff)
	} else if !cmp.Equal(betaPrime.Mmr.Peaks, h.PostState.Beta.Mmr.Peaks) {
		diff := cmp.Diff(h.PostState.Beta.Mmr.Peaks, betaPrime.Mmr.Peaks)
		return fmt.Errorf("BetaPrime.Mmr.Peaks should equal to PostState.Beta.Mmr.Peaks\n%s", diff)
	}
	return nil
}
