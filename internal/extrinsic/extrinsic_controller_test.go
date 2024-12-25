package extrinsic

import (
	"testing"
)

func TestNewExtrinsicController(t *testing.T) {
	// Create mock controllers
	ticketController := NewTicketController()
	preimageController := NewPreimageController()
	guaranteeController := NewGuaranteeController()
	disputeController := NewDisputeController(NewVerdictController(), NewFaultController(), NewCulpritController())
	availAssuranceController := NewAvailAssuranceController()

	// Create the ExtrinsicController
	extrinsicController := NewExtrinsicController(
		ticketController,
		preimageController,
		guaranteeController,
		disputeController,
		availAssuranceController,
	)

	// Test that controller is not nil
	if extrinsicController == nil {
		t.Fatal("Expected extrinsicController to not be nil")
	}

	// Test that all fields are correctly set
	if extrinsicController.TicketController != ticketController {
		t.Error("TicketController not properly set")
	}

	if extrinsicController.PreimageController != preimageController {
		t.Error("PreimageController not properly set")
	}

	if extrinsicController.GuaranteeController != guaranteeController {
		t.Error("GuaranteeController not properly set")
	}

	if extrinsicController.DisputeController != disputeController {
		t.Error("DisputeController not properly set")
	}

	if extrinsicController.AvailAssuranceController != availAssuranceController {
		t.Error("AvailAssuranceController not properly set")
	}
}
