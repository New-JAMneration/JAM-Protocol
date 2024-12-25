package extrinsic

import (
	"testing"
)

func TestNewTicketController(t *testing.T) {
	tickcontroller := NewTicketController()

	// Test that controller is not nil
	if tickcontroller == nil {
		t.Fatal("Expected tickcontroller to not be nil")
	}

	// Test that TicketEnvelopes slice is initialized
	if tickcontroller.TicketEnvelopes == nil {
		t.Fatal("Expected ticket envelope to be initialized")
	}

	// Test that TicketEnvelopes slice is empty
	if len(tickcontroller.TicketEnvelopes) != 0 {
		t.Errorf("Expected empty ticket envelope slice, got length %d", len(tickcontroller.TicketEnvelopes))
	}
}
