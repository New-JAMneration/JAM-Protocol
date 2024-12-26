package safrole

import (
	"fmt"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// (6.5) // (6.6)
// TicketsBodiesController is a controller for TicketsBodies
type TicketsBodiesController struct {
	TicketsBodies []TicketBody
}

// NewTicketsBodiesController returns a new TicketsBodiesController
func NewTicketsBodiesController() *TicketsBodiesController {
	return &TicketsBodiesController{
		TicketsBodies: make([]TicketBody, 0),
	}
}

// Validate validates the controller
func (t *TicketsBodiesController) Validate() error {
	if len(t.TicketsBodies) > jamTypes.EpochLength {
		return fmt.Errorf("TicketsBodiesController must have less than %d entries, got %d", jamTypes.EpochLength, len(t.TicketsBodies))
	}
	return nil
}

// (6.5)
// AddTicketBody adds a ticket body to the controller
func (t *TicketsBodiesController) AddTicketBody(ticketBody TicketBody) *TicketsBodiesController {
	t.TicketsBodies = append(t.TicketsBodies, ticketBody)
	return t
}
