package extrinsic

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// TicketController is a struct that contains a slice of Ticket (for controller logic)
type TicketController struct {
	TicketEnvelopes []types.TicketEnvelope
}

// NewTicketController returns a new TicketController
func NewTicketController() *TicketController {
	return &TicketController{
		TicketEnvelopes: make([]types.TicketEnvelope, 0),
	}
}
