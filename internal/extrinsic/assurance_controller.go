package extrinsic

import (
	"sort"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

// AvailAssuranceController is a struct that contains a slice of AvailAssurance
type AvailAssuranceController struct {
	AvailAssurances []jamTypes.AvailAssurance
	AssuranceCheck  map[jamTypes.ValidatorIndex]struct{} // assurance at most one per validator, use map to check
}

// NewAvailAssuranceController creates a new AvailAssuranceController
func NewAvailAssuranceController() *AvailAssuranceController {
	return &AvailAssuranceController{
		AvailAssurances: make([]jamTypes.AvailAssurance, 0),
		AssuranceCheck:  make(map[jamTypes.ValidatorIndex]struct{}),
	}
}

// Add adds a new AvailAssurance to the AvailAssurance slice
func (a *AvailAssuranceController) Add(newAvailAssurance jamTypes.AvailAssurance) {
	if !a.CheckAssuranceExist(newAvailAssurance) { // return false = validatorIndex has not submit assurance
		a.AssuranceCheck[newAvailAssurance.ValidatorIndex] = struct{}{}
		a.AvailAssurances = append(a.AvailAssurances, newAvailAssurance)
	}
}

// CheckAssuranceExist checks if the AvailAssurance already exists
func (a *AvailAssuranceController) CheckAssuranceExist(availAssurance jamTypes.AvailAssurance) (exist bool) {
	if _, exists := a.AssuranceCheck[availAssurance.ValidatorIndex]; exists {
		return true
	}

	return false
}

// SortAssurances sorts the AvailAssurance slice
func (a *AvailAssuranceController) SortAssurances() {
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

type AvailAssurance jamTypes.AvailAssurance

// SetCoreBit is for setting the validity of the core-index in the assurance bit
func (a AvailAssurance) SetCoreBit(coreIndex uint16, validity bool) {
	// equation 11.15 ~ 11.16
	// 先判定 core 有 pending available 的 report 和 timeout (timeout 在 jam_types line 101 已經處理)
	// 由 workreport get core-index，並驗證 validity
}

// GenerateMessage is for generating the message, which is along with the assurance signature
func GenerateMessage() {
	// equation 11.13,  currently lack of serialization & encoding & signing context

}
