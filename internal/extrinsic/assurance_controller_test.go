package extrinsic

import (
	"testing"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

func TestAvailAssuranceController(t *testing.T) {
	AvailAssurances := NewAvailAssuranceController()

	if len(AvailAssurances.AvailAssurances) != 0 {
		t.Errorf("Expected %d assurances, got %d", 0, len(AvailAssurances.AvailAssurances))
	}
}

func TestAddAssurance(t *testing.T) { // include test for CheckAssuranceExist
	AvailAssurances := NewAvailAssuranceController()
	AvailAssurances.Add(jamTypes.AvailAssurance{ValidatorIndex: 0})
	AvailAssurances.Add(jamTypes.AvailAssurance{ValidatorIndex: 1})
	AvailAssurances.Add(jamTypes.AvailAssurance{ValidatorIndex: 1})

	Expected := []jamTypes.AvailAssurance{
		{ValidatorIndex: 0},
		{ValidatorIndex: 1},
	}

	if len(AvailAssurances.AvailAssurances) != len(Expected) {
		t.Errorf("Expected %d assurances, got %d", len(Expected), len(AvailAssurances.AvailAssurances))
	} else {
		for i := 0; i < len(AvailAssurances.AvailAssurances); i++ {
			if AvailAssurances.AvailAssurances[i].ValidatorIndex != Expected[i].ValidatorIndex {
				t.Errorf("Expected assurance %d to have validator index %d, got %d", i, Expected[i].ValidatorIndex, AvailAssurances.AvailAssurances[i].ValidatorIndex)
			}
		}
	}
}

func TestSortAssurances(t *testing.T) {
	AvailAssurances := NewAvailAssuranceController()
	AvailAssurances.AvailAssurances = []jamTypes.AvailAssurance{
		{ValidatorIndex: 1},
		{ValidatorIndex: 10},
		{ValidatorIndex: 6},
		{ValidatorIndex: 0},
	}

	AvailAssurances.SortAssurances()
	Expected := []jamTypes.AvailAssurance{
		{ValidatorIndex: 0},
		{ValidatorIndex: 1},
		{ValidatorIndex: 6},
		{ValidatorIndex: 10},
	}

	if len(AvailAssurances.AvailAssurances) != len(Expected) {
		t.Errorf("Expected %d assurances, got %d", len(Expected), len(AvailAssurances.AvailAssurances))
	} else {
		for i := 0; i < len(AvailAssurances.AvailAssurances); i++ {
			if AvailAssurances.AvailAssurances[i].ValidatorIndex != Expected[i].ValidatorIndex {
				t.Errorf("Expected assurance %d to have validator index %d, got %d", i, Expected[i].ValidatorIndex, AvailAssurances.AvailAssurances[i].ValidatorIndex)
			}
		}
	}
}
