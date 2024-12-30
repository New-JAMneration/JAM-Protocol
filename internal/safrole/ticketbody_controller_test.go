package safrole_test

import (
	"fmt"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestTicketsBodiesController(t *testing.T) {
	// Mock jamTypes.EpochLength
	types.EpochLength = 5

	// Test initialization
	controller := safrole.NewTicketsBodiesController()
	if controller == nil {
		t.Fatal("Controller should be initialized successfully")
	}
	if len(controller.TicketsBodies) != 0 {
		t.Fatalf("Expected controller to have no tickets initially, got %d", len(controller.TicketsBodies))
	}

	// Test adding ticket
	ticket := types.TicketBody{}
	if len(controller.TicketsBodies) < types.EpochLength {
		controller.AddTicketBody(ticket)
	}
	if len(controller.TicketsBodies) != 1 {
		t.Fatalf("Expected controller to have one ticket after adding, got %d", len(controller.TicketsBodies))
	}

	// Test exceeding limits
	for i := 0; i <= types.EpochLength; i++ {
		controller.AddTicketBody(ticket)
	}

	err := controller.AddTicketBody(ticket)
	var expectedErr error = fmt.Errorf("TicketsBodiesController must have less than %d entries, got %d", types.EpochLength, len(controller.TicketsBodies))
	if err.Error() != expectedErr.Error() {
		t.Fatalf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}

}
