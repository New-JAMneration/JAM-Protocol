package header

import (
	"errors"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// TestGetParentHeader tests the GetParentHeader function.
// If the header is the genesis header, it should return an error.
// Otherwise, it should return the parent header.
func TestGetParentHeader(t *testing.T) {
	testCases := []struct {
		header         types.Header
		expectedErr    error
		expectedHeader types.Header
	}{
		{
			types.Header{
				Slot: 0,
			},
			errors.New("This is the genesis header."),
			types.Header{},
		},
		{
			types.Header{
				Slot: 1,
			},
			nil,
			types.Header{
				Slot: 0,
			},
		},
	}

	// New HeaderController
	hc := NewHeaderController()

	// Get current header
	s := store.GetInstance()

	// Add headers
	header1 := types.Header{
		Slot: 0,
	}

	header2 := types.Header{
		Slot: 1,
	}

	s.AddAncestorHeader(header1)
	s.AddAncestorHeader(header2)

	for _, tc := range testCases {
		parentHeader, err := hc.GetParentHeader(tc.header)
		if err != nil {
			if err.Error() != tc.expectedErr.Error() {
				t.Errorf("Expected error: %s, got: %s", tc.expectedErr.Error(), err.Error())
			}
		} else {
			if parentHeader.Slot != tc.expectedHeader.Slot {
				t.Errorf("Expected header slot: %v, got: %v", tc.expectedHeader.Slot, parentHeader.Slot)
			}
		}
	}
}

func TestValidateHeaderTimeslot(t *testing.T) {
	testCases := []struct {
		header      types.Header
		expectedErr error
		isValid     bool
	}{
		{
			types.Header{
				Slot: 0,
			},
			errors.New("This is the genesis header."),
			false,
		},
		{
			types.Header{
				Slot: 1,
			},
			nil,
			true,
		},
		{
			types.Header{
				Slot: 2,
			},
			nil,
			true,
		},
	}

	// Get current header
	s := store.GetInstance()

	s.AddAncestorHeader(types.Header{
		Slot: 0,
	})
	s.AddAncestorHeader(types.Header{
		Slot: 1,
	})
	s.AddAncestorHeader(types.Header{
		Slot: 2,
	})

	// New HeaderController
	hc := NewHeaderController()

	for _, tc := range testCases {
		headerIsValid, err := hc.ValidateHeaderTimeslot(tc.header)

		if err != nil {
			if err.Error() != tc.expectedErr.Error() {
				t.Errorf("Expected error: %s, got: %s", tc.expectedErr.Error(), err.Error())
			}
		} else {
			if headerIsValid != tc.isValid {
				t.Errorf("Expected headerIsValid: %v, got: %v", tc.isValid, headerIsValid)
			}
		}
	}
}
