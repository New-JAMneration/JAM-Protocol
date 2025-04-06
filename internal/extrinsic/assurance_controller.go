package extrinsic

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
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
			return errors.New("bad_attestation_parent")
		}
	}

	return nil
}

func (a *AvailAssuranceController) CheckValidatorIndex() error {
	for _, availAssurance := range a.AvailAssurances {
		if int(availAssurance.ValidatorIndex) >= types.ValidatorsCount {
			return errors.New("bad_validator_index")
		}
	}
	return nil
}

// SortUnique sorts the AvailAssurance slice and removes duplicates | Eq. 11.12
func (a *AvailAssuranceController) SortUnique() error {
	err := a.Unique()
	a.Sort()
	if err != nil {
		return err
	}

	return nil
}

// Unique removes duplicates
func (a *AvailAssuranceController) Unique() error {
	if len(a.AvailAssurances) == 0 {
		return nil
	}

	uniqueMap := make(map[types.ValidatorIndex]bool)
	result := make([]types.AvailAssurance, 0)

	for i, availAssurance := range a.AvailAssurances {
		if !uniqueMap[availAssurance.ValidatorIndex] && int(availAssurance.ValidatorIndex) == i {
			uniqueMap[availAssurance.ValidatorIndex] = true
			result = append(result, availAssurance)
		} else {
			return errors.New("not_sorted_or_unique_assurers")
		}
	}

	a.AvailAssurances = result

	return nil
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
	kappa := store.GetInstance().GetPriorStates().GetKappa()

	for _, availAssurance := range a.AvailAssurances {
		anchor := utilities.OpaqueHashWrapper{Value: availAssurance.Anchor}.Serialize()
		bitfield := utilities.ByteSequenceWrapper{Value: types.ByteSequence(availAssurance.Bitfield)}.Serialize()
		hased := hash.Blake2bHash(append(anchor, bitfield...))
		message := []byte(types.JamAvailable)
		message = append(message, hased[:]...)

		publicKey := kappa[availAssurance.ValidatorIndex].Ed25519
		if !ed25519.Verify(publicKey[:], message, availAssurance.Signature[:]) {
			return errors.New("invalid_signature")
		}
	}

	return nil
}

// ValidateBitField | Eq. 11.15
func (a *AvailAssuranceController) ValidateBitField() error {
	rhoDagger := store.GetInstance().GetIntermediateStates().GetRhoDagger()

	for i := 0; i < len(a.AvailAssurances); i++ {
		for j := 0; j < types.CoresCount; j++ {
			// rhoDagger[j] nil : core j has no report to be process
			// assurers can not set nil core
			if a.AvailAssurances[i].Bitfield[j] == byte(1) && rhoDagger[j] == nil {
				return errors.New("core_engaged")
			}
		}
	}
	return nil
}

// BitfieldOctetSequenceToBinarySequence transform the input octet bitfield to a binary sequence
func (a *AvailAssuranceController) BitfieldOctetSequenceToBinarySequence() {
	// input (bitfield) : octet sequence ,  output	(binaryBitfield) : binary sequence
	if len(a.AvailAssurances) <= 0 {
		return
	}
	/*
		length := len(a.AvailAssurances[0].Bitfield)
		bitLength := length * 8
	*/
	bitLength := types.CoresCount
	for assuranceIndex := 0; assuranceIndex < len(a.AvailAssurances); assuranceIndex++ {
		binaryBitfield := make([]byte, bitLength)

		for i := 0; i < len(a.AvailAssurances[0].Bitfield); i++ {
			for j := 0; j < 8; j++ {
				if i*8+j >= bitLength {
					break
				}
				if a.AvailAssurances[assuranceIndex].Bitfield[i]&(1<<j) != 0 {
					binaryBitfield[i*8+j] = 1
				} else {
					binaryBitfield[i*8+j] = 0
				}
			}
		}

		a.AvailAssurances[assuranceIndex].Bitfield = make([]byte, bitLength)
		copy(a.AvailAssurances[assuranceIndex].Bitfield, binaryBitfield)
	}
}

// Filter newly available work reports | Eq. 11.16
func (a *AvailAssuranceController) UpdateNewlyAvailableWorkReports(rhoDagger types.AvailabilityAssignments) []types.WorkReport {
	// Filter newly available work reports from rhoDagger
	availableNumber := types.ValidatorsCount * 2 / 3
	totalAvailable := make([]int, types.CoresCount)

	// compute total availability of a report | at this moment of the workflow, the bitfield is transformed into a binary sequence.
	for i := 0; i < len(a.AvailAssurances); i++ {
		for j := 0; j < types.CoresCount; j++ {
			if a.AvailAssurances[i].Bitfield[j] == 1 {
				totalAvailable[j]++
			}
		}
	}

	availableWorkReports := []types.WorkReport{}
	for i := 0; i < types.CoresCount; i++ {
		// If the votes for this core are greater than the available number, add the work report to the available work reports
		if totalAvailable[i] > availableNumber {
			// Get work reports from rhoDagger
			if rhoDagger[i] == nil {
				continue
			}

			// Append the work report to the available work reports
			availableWorkReports = append(availableWorkReports, rhoDagger[i].Report)
		}
	}

	// Set the available work reports to the available work reports
	store := store.GetInstance()
	store.GetAvailableWorkReportsPointer().SetAvailableWorkReports(availableWorkReports)

	return availableWorkReports
}

// Create a work report map for checking a work report existence
func (a *AvailAssuranceController) CreateWorkReportMap(workReports []types.WorkReport) map[types.CoreIndex]bool {
	workReportMap := make(map[types.CoreIndex]bool)

	for _, workReport := range workReports {
		workReportMap[workReport.CoreIndex] = true
	}

	return workReportMap
}

// FilterAvailableReports | Eq. 11.16 & 11.17
func (a *AvailAssuranceController) FilterAvailableReports() error {
	if len(a.AvailAssurances) == 0 {
		return nil
	}

	store := store.GetInstance()

	rhoDagger := store.GetIntermediateStates().GetRhoDagger()
	rho := store.GetPriorStates().GetRho()

	// 11.17 Set available reports or timeout reports to nil
	rhoDoubleDagger := rhoDagger
	headerTimeSlot := store.GetBlock().Header.Slot

	// (11.16) Filter newly available work reports
	availableWorkReports := a.UpdateNewlyAvailableWorkReports(rhoDagger)

	// Create a map of available work reports for faster lookup
	availableWorkReportsMap := a.CreateWorkReportMap(availableWorkReports)

	for coreIndex := 0; coreIndex < types.CoresCount; coreIndex++ {
		if rho[coreIndex] == nil {
			continue
		}

		reportIsAvailable := availableWorkReportsMap[rho[coreIndex].Report.CoreIndex]
		reportIsTimeout := headerTimeSlot >= rhoDagger[coreIndex].Timeout+types.TimeSlot(types.WorkReportTimeout)

		if reportIsAvailable || reportIsTimeout {
			rhoDoubleDagger[coreIndex] = nil
		}
	}

	store.GetIntermediateStates().SetRhoDoubleDagger(rhoDoubleDagger)

	return nil
}
