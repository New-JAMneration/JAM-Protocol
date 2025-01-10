package extrinsic

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"github.com/New-JAMneration/JAM-Protocol/internal/input/jam_types"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"sort"
)

var (
	// U is defined in I.4.4
	U = 5
)

// AvailAssuranceController is a struct that contains a slice of AvailAssurance
type AvailAssuranceController struct {
	AvailAssurances types.AssurancesExtrinsic `json:"assurances"`
	/*
		type AvailAssurance struct {
			Anchor         OpaqueHash       `json:"anchor,omitempty"`
			Bitfield       []byte           `json:"bitfield,omitempty"`
			ValidatorIndex ValidatorIndex   `json:"validator_index,omitempty"`
			Signature      Ed25519Signature `json:"signature,omitempty"`
		}
	*/
}

// NewAvailAssuranceController creates a new AvailAssuranceController
func NewAvailAssuranceController() *AvailAssuranceController {
	return &AvailAssuranceController{
		AvailAssurances: make(types.AssurancesExtrinsic, 0),
	}
}

// ValidateAnchor validates the anchor of the AvailAssurance | Eq. 11.11
func (a *AvailAssuranceController) ValidateAnchor() error {
	headerParent := types.OpaqueHash(store.GetInstance().GetBlock().Header.Parent)

	for _, availAssurance := range a.AvailAssurances {
		if !bytes.Equal(availAssurance.Anchor[:], headerParent[:]) {
			return fmt.Errorf("AvailAssuranceController.ValidateAnchor failed : bad_attestation_parent")
		}
	}

	return nil
}

// SortUnique sorts the AvailAssurance slice and removes duplicates | Eq. 11.12
func (a *AvailAssuranceController) SortUnique() {
	a.Unique()
	a.Sort()
}

// Unique removes duplicates
func (a *AvailAssuranceController) Unique() {
	if len(a.AvailAssurances) == 0 {
		return
	}

	uniqueMap := make(map[types.ValidatorIndex]bool)
	result := make([]types.AvailAssurance, 0)

	for _, availAssurance := range a.AvailAssurances {
		if !uniqueMap[availAssurance.ValidatorIndex] {
			uniqueMap[availAssurance.ValidatorIndex] = true
			result = append(result, availAssurance)
		}
	}
}

// Sort sorts the AvailAssurance slice
func (a *AvailAssuranceController) Sort() {
	sort.Sort(a)
}

func (a *AvailAssuranceController) Len() int {
	return len(a.AvailAssurances)
}

func (a *AvailAssuranceController) Less(i, j int) bool {
	return a.AvailAssurances[i].ValidatorIndex < a.AvailAssurances[j].ValidatorIndex
}

func (a *AvailAssuranceController) Swap(i, j int) {
	a.AvailAssurances[i], a.AvailAssurances[j] = a.AvailAssurances[j], a.AvailAssurances[i]
}

// ValidateSignature validates the signature of the AvailAssurance | Eq. 11.13, 11.14
func (a *AvailAssuranceController) ValidateSignature() error {
	kappaPrime := store.GetInstance().GetPosteriorState().Kappa

	for _, availAssurance := range a.AvailAssurances {
		anchor := utilities.OpaqueHashWrapper{Value: availAssurance.Anchor}.Serialize()
		bitfield := utilities.ByteSequenceWrapper{Value: types.ByteSequence(availAssurance.Bitfield)}.Serialize()
		message := []byte(jam_types.JamAvailable)
		message = append(message, anchor...)
		message = append(message, bitfield...)
		hashed := hash.Blake2bHash(message)
		message = hashed[:]

		publicKey := kappaPrime[availAssurance.ValidatorIndex].Ed25519[:]
		if !ed25519.Verify(publicKey, message, availAssurance.Signature[:]) {
			return fmt.Errorf("invalid_signature")
		}
	}
	return nil
}

// ValidateBitField | Eq. 11.15
func (a *AvailAssuranceController) ValidateBitField() error {
	rhoDagger := store.GetInstance().GetIntermediateStates().GetRhoDagger()

	for i := 0; i < len(a.AvailAssurances); i++ {
		byteIndex, bitIndex := 0, 0
		for j := 0; j < jam_types.CoresCount; j++ {
			// compute bitfield is 1 or 0
			if rhoDagger[j] == nil {
				if a.AvailAssurances[i].Bitfield[byteIndex]&(1<<bitIndex) != 0 {
					return fmt.Errorf("AvailAssuranceController.ValidateBitField failed : core_not_engaged")
				}
			}
			bitIndex++
			if bitIndex == 8 {
				byteIndex++
				bitIndex = 0
			}
		}
	}
	return nil
}

// FilterAvailableReports | Eq. 11.16 & 11.17
func (a *AvailAssuranceController) FilterAvailableReports() {
	rhoDagger := store.GetInstance().GetIntermediateStates().GetRhoDagger()
	availableNumber := jam_types.ValidatorsCount * 2 / 3
	totalAvailable := make([]int, jam_types.CoresCount)

	for i := 0; i < len(a.AvailAssurances); i++ {
		byteIndex, bitIndex := 0, 0
		for j := 0; j < jam_types.CoresCount; j++ {
			if a.AvailAssurances[i].Bitfield[byteIndex]&(1<<bitIndex) == 1 {
				totalAvailable[j]++
			}
			bitIndex++
			if bitIndex == 8 {
				byteIndex++
				bitIndex = 0
			}
		}
	}

	// 11.17 Set available reports or timeout reports to nil
	rhoDoubleDagger := rhoDagger
	headerTimeSlot := store.GetInstance().GetBlock().Header.Slot

	for i := 0; i < jam_types.CoresCount; i++ {
		if totalAvailable[i] < availableNumber || headerTimeSlot >= rhoDagger[i].Timeout+types.TimeSlot(U) {
			rhoDoubleDagger[i] = nil
		}
	}
	store.GetInstance().GetIntermediateStates().SetRhoDoubleDagger(rhoDoubleDagger)

}
