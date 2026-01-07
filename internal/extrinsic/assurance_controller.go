package extrinsic

import (
	"bytes"
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	AssuranceErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/assurances"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/hdevalence/ed25519consensus"
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
func (a *AvailAssuranceController) ValidateAnchor() *types.ErrorCode {
	headerParent := types.OpaqueHash(blockchain.GetInstance().GetLatestBlock().Header.Parent)

	for _, availAssurance := range a.AvailAssurances {
		if !bytes.Equal(availAssurance.Anchor[:], headerParent[:]) {
			errCode := AssuranceErrorCode.BadAttestationParent
			return &errCode
		}
	}

	return nil
}

func (a *AvailAssuranceController) CheckValidatorIndex() *types.ErrorCode {
	for _, availAssurance := range a.AvailAssurances {
		if int(availAssurance.ValidatorIndex) >= types.ValidatorsCount {
			errCode := AssuranceErrorCode.BadValidatorIndex
			return &errCode
		}
	}
	return nil
}

// SortUnique sorts the AvailAssurance slice and removes duplicates | Eq. 11.12
func (a *AvailAssuranceController) SortUnique() *types.ErrorCode {
	err := a.CheckUniqueAndSort()
	if err != nil {
		return err
	}

	return nil
}

// CheckUniqueAndSort checks if the AvailAssurance slice is sorted and unique
func (a *AvailAssuranceController) CheckUniqueAndSort() *types.ErrorCode {
	if len(a.AvailAssurances) == 0 {
		return nil
	}

	uniqueMap := make(map[types.ValidatorIndex]bool)
	result := make([]types.AvailAssurance, 0)
	var last types.ValidatorIndex = 0
	for _, availAssurance := range a.AvailAssurances {
		if availAssurance.ValidatorIndex < last {
			errCode := AssuranceErrorCode.NotSortedOrUniqueAssurers
			return &errCode
		}

		if uniqueMap[availAssurance.ValidatorIndex] {
			errCode := AssuranceErrorCode.NotSortedOrUniqueAssurers
			return &errCode
		}

		uniqueMap[availAssurance.ValidatorIndex] = true
		result = append(result, availAssurance)
		last = availAssurance.ValidatorIndex
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
func (a *AvailAssuranceController) ValidateSignature() *types.ErrorCode {
	kappa := blockchain.GetInstance().GetPriorStates().GetKappa()

	for _, availAssurance := range a.AvailAssurances {
		anchor := utilities.OpaqueHashWrapper{Value: availAssurance.Anchor}.Serialize()
		bitfield := utilities.ByteSequenceWrapper{Value: types.ByteSequence(availAssurance.Bitfield.ToOctetSlice())}.Serialize()
		hased := hash.Blake2bHash(append(anchor, bitfield...))
		message := []byte(types.JamAvailable)
		message = append(message, hased[:]...)

		publicKey := kappa[availAssurance.ValidatorIndex].Ed25519
		if !ed25519consensus.Verify(publicKey[:], message, availAssurance.Signature[:]) {
			errCode := AssuranceErrorCode.BadSignature
			return &errCode
		}
	}

	return nil
}

// ValidateBitField | Eq. 11.15
func (a *AvailAssuranceController) ValidateBitField() *types.ErrorCode {
	rhoDagger := blockchain.GetInstance().GetIntermediateStates().GetRhoDagger()

	for i := 0; i < len(a.AvailAssurances); i++ {
		for j := 0; j < types.CoresCount; j++ {
			// rhoDagger[j] nil : core j has no report to be process
			// assurers can not set nil core
			if a.AvailAssurances[i].Bitfield.GetBit(j) == 1 && rhoDagger[j] == nil {
				errCode := AssuranceErrorCode.CoreNotEngaged
				return &errCode
			}
		}
	}
	return nil
}

// Filter newly available work reports | Eq. 11.16
func (a *AvailAssuranceController) UpdateNewlyAvailableWorkReports(rhoDagger types.AvailabilityAssignments) []types.WorkReport {
	// Filter newly available work reports from rhoDagger
	totalAvailable := make([]int, types.CoresCount)

	// compute total availability of a report | at this moment of the workflow, the bitfield is transformed into a binary sequence.
	for i := 0; i < len(a.AvailAssurances); i++ {
		for j := 0; j < types.CoresCount; j++ {
			if a.AvailAssurances[i].Bitfield.GetBit(j) == 1 {
				totalAvailable[j]++
			}
		}
	}

	availableWorkReports := []types.WorkReport{}
	for i := 0; i < types.CoresCount; i++ {
		// If the votes for this core are greater than the available number, add the work report to the available work reports
		if totalAvailable[i] >= types.ValidatorsSuperMajority {
			// Get work reports from rhoDagger
			if rhoDagger[i] == nil {
				continue
			}

			// Append the work report to the available work reports
			availableWorkReports = append(availableWorkReports, rhoDagger[i].Report)
		}
	}

	// Set the available work reports to the available work reports
	blockchain.GetInstance().GetIntermediateStates().SetAvailableWorkReports(availableWorkReports)

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
func (a *AvailAssuranceController) FilterAvailableReports() *types.ErrorCode {
	cs := blockchain.GetInstance()

	rhoDagger := cs.GetIntermediateStates().GetRhoDagger()
	rhoDoubleDagger := cs.GetIntermediateStates().GetRhoDoubleDagger()
	rho := cs.GetPriorStates().GetRho()

	// 11.17 Set available reports or timeout reports to nil
	// Make a copy to avoid aliasing with rhoDagger
	copy(rhoDoubleDagger, rhoDagger)
	headerTimeSlot := cs.GetLatestBlock().Header.Slot

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

	cs.GetIntermediateStates().SetRhoDoubleDagger(rhoDoubleDagger)

	return nil
}
