package safrole

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// (6.5) // (6.6)
// TicketsBodiesController is a controller for TicketsBodies
type TicketsBodiesController struct {
	TicketsBodies []types.TicketBody
}

// NewTicketsBodiesController returns a new TicketsBodiesController
func NewTicketsBodiesController() *TicketsBodiesController {
	return &TicketsBodiesController{
		TicketsBodies: make([]types.TicketBody, 0),
	}
}

// Validate validates the controller
func (tbc *TicketsBodiesController) Validate() error {
	if len(tbc.TicketsBodies) > types.EpochLength {
		return fmt.Errorf("TicketsBodiesController must have less than %d entries, got %d", types.EpochLength, len(tbc.TicketsBodies))
	}
	return nil
}

// (6.5)
// AddTicketBody adds a ticket body to the controller
func (tbc *TicketsBodiesController) AddTicketBody(ticketBody types.TicketBody) error {
	if len(tbc.TicketsBodies) < types.EpochLength {
		tbc.TicketsBodies = append(tbc.TicketsBodies, ticketBody)
		return nil
	}
	return fmt.Errorf("TicketsBodiesController must have less than %d entries, got %d", types.EpochLength, len(tbc.TicketsBodies))
}
