package utilities

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestEmptyOrPair(t *testing.T) {
	/*
			type TicketsMark []TicketBody

			type TicketBody struct {
			ID      TicketID      = OpaqueHash = ByteArray32 = [32]byte 	`json:"id,omitempty"`
			Attempt TicketAttempt = U64										`json:"attempt,omitempty"`
		}
	*/
	// test 1 empty
	ticketsMark := types.TicketsMark{}
	result, _ := EmptyOrPair(ticketsMark)
	expected := 0
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}

	// test 2 empty Ptr
	ticketsMarkPtr := &ticketsMark
	result, _ = EmptyOrPair(ticketsMarkPtr)
	expected = 0
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}

	// test 3 not empty
	ticketsMark = []types.TicketBody{
		{ID: types.TicketID{0x01}, Attempt: types.TicketAttempt(1)},
		{ID: types.TicketID{0x02}, Attempt: types.TicketAttempt(2)},
	}
	result, _ = EmptyOrPair(ticketsMark)
	expected = 1

	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}

	// test 4 not empty Ptr
	ticketsMarkPtr = &ticketsMark
	result, _ = EmptyOrPair(ticketsMarkPtr)
	expected = 1
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}
