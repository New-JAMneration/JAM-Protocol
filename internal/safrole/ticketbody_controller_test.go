package safrole_test

import (
	"fmt"
	"testing"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/stretchr/testify/assert"
)

func TestTicketsBodiesController(t *testing.T) {
	// Mock jamTypes.EpochLength
	jamTypes.EpochLength = 5

	// Test initialization
	controller := safrole.NewTicketsBodiesController()
	assert.NotNil(t, controller, "Controller should be initialized successfully")
	assert.Equal(t, 0, len(controller.TicketsBodies), "Controller should have no tickets initially")

	// Test adding ticket
	ticket := safrole.TicketBody{}
	controller.AddTicketBody(ticket)
	assert.Equal(t, 1, len(controller.TicketsBodies), "Controller should have one ticket after adding")

	// Test validation success
	err := controller.Validate()
	assert.NoError(t, err, "Validation should pass when ticket count is within limits")

	// Test exceeding limits
	for i := 0; i < jamTypes.EpochLength; i++ {
		controller.AddTicketBody(ticket)
	}
	err = controller.Validate()
	assert.Error(t, err, "Validation should fail when ticket count exceeds limits")
	assert.EqualError(t, err, fmt.Sprintf("TicketsBodiesController must have less than %d entries, got %d", jamTypes.EpochLength, len(controller.TicketsBodies)))

	// Test AddTicketBody method should not add more tickets after exceeding the limit
	initialLength := len(controller.TicketsBodies)
	controller.AddTicketBody(ticket)
	assert.Equal(t, initialLength, len(controller.TicketsBodies), "Controller should not allow adding more tickets after exceeding the limit")
}
