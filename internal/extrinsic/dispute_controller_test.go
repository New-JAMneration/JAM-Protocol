package extrinsic

import (
	"testing"
)

func TestNewVerdictController(t *testing.T) {
	verdictController := NewVerdictController()

	// Test that controller is not nil
	if verdictController == nil {
		t.Fatal("Expected VerdictController to not be nil")
	}

	// Test that Verdicts slice is initialized
	if verdictController.Verdicts == nil {
		t.Fatal("Expected Verdicts slice to be initialized")
	}

	// Test that Verdicts slice is empty
	if len(verdictController.Verdicts) != 0 {
		t.Errorf("Expected empty verdicts slice, got length %d", len(verdictController.Verdicts))
	}
}

func TestNewFaultController(t *testing.T) {
	faultController := NewFaultController()

	// Test that controller is not nil
	if faultController == nil {
		t.Fatal("Expected FaultController to not be nil")
	}

	// Test that Faults slice is initialized
	if faultController.Faults == nil {
		t.Fatal("Expected Faults slice to be initialized")
	}

	// Test that Faults slice is empty
	if len(faultController.Faults) != 0 {
		t.Errorf("Expected empty faults slice, got length %d", len(faultController.Faults))
	}
}

func TestNewCulpritController(t *testing.T) {
	culpritController := NewCulpritController()

	// Test that controller is not nil
	if culpritController == nil {
		t.Fatal("Expected CulpritController to not be nil")
	}

	// Test that Culprits slice is initialized
	if culpritController.Culprits == nil {
		t.Fatal("Expected Culprits slice to be initialized")
	}

	// Test that Culprits slice is empty
	if len(culpritController.Culprits) != 0 {
		t.Errorf("Expected empty culprits slice, got length %d", len(culpritController.Culprits))
	}
}

func TestNewDisputeController(t *testing.T) {
	// Create sub-controllers
	verdictController := NewVerdictController()
	faultController := NewFaultController()
	culpritController := NewCulpritController()

	// Create the DisputeController
	disputeController := NewDisputeController(verdictController, faultController, culpritController)

	// Test that controller is not nil
	if disputeController == nil {
		t.Fatal("Expected DisputeController to not be nil")
	}

	// Test that all sub-controllers are correctly set
	if disputeController.VerdictController != verdictController {
		t.Error("VerdictController not properly set")
	}

	if disputeController.FaultController != faultController {
		t.Error("FaultController not properly set")
	}

	if disputeController.CulpritController != culpritController {
		t.Error("CulpritController not properly set")
	}
}
