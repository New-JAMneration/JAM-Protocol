package safrole_test

import (
	"fmt"
	"testing"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
)

func TestTicketsBodiesController(t *testing.T) {
	// Mock jamTypes.EpochLength
	jamTypes.EpochLength = 5

	// Test initialization
	controller := safrole.NewTicketsBodiesController()
	if controller == nil {
		t.Fatal("Controller should be initialized successfully")
	}
	if len(controller.TicketsBodies) != 0 {
		t.Fatalf("Expected controller to have no tickets initially, got %d", len(controller.TicketsBodies))
	}

	// Test adding ticket
	ticket := safrole.TicketBody{}
	if len(controller.TicketsBodies) < jamTypes.EpochLength {
		controller.AddTicketBody(ticket)
	}
	if len(controller.TicketsBodies) != 1 {
		t.Fatalf("Expected controller to have one ticket after adding, got %d", len(controller.TicketsBodies))
	}

	// Test validation success
	err := controller.Validate()
	if err != nil {
		t.Fatalf("Validation should pass when ticket count is within limits, got error: %v", err)
	}

	// Test exceeding limits
	for i := 0; i <= jamTypes.EpochLength; i++ {
		if len(controller.TicketsBodies) <= jamTypes.EpochLength {
			controller.AddTicketBody(ticket)
		}
	}
	err = controller.Validate()
	if err == nil {
		t.Fatal("Validation should fail when ticket count exceeds limits")
	}
	expectedErr := fmt.Sprintf("TicketsBodiesController must have less than %d entries, got %d", jamTypes.EpochLength, len(controller.TicketsBodies))
	if err.Error() != expectedErr {
		t.Fatalf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}

	// Test AddTicketBody method should not add more tickets after exceeding the limit
	initialLength := len(controller.TicketsBodies)
	if len(controller.TicketsBodies) < jamTypes.EpochLength {
		controller.AddTicketBody(ticket)
	}
	if len(controller.TicketsBodies) != initialLength {
		t.Fatalf("Controller should not allow adding more tickets after exceeding the limit, expected length: %d, got: %d", initialLength, len(controller.TicketsBodies))
	}
}
