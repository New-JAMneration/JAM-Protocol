package extrinsic

import (
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// TicketController is a struct that contains a slice of Ticket (for controller logic)
type TicketController struct {
	TicketEnvelopes []jamTypes.TicketEnvelope
}

// NewTicketController returns a new TicketController
func NewTicketController() *TicketController {
	return &TicketController{
		TicketEnvelopes: make([]jamTypes.TicketEnvelope, 0),
	}
}
