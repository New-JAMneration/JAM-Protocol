package safrole

import (
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// GetEpochIndex returns the epoch index of the most recent block't timeslot
// \tau : The most recent block't timeslot
// (6.2)
func GetEpochIndex(t jamTypes.TimeSlot) jamTypes.TimeSlot {
	return t / jamTypes.TimeSlot(jamTypes.EpochLength)
}

// GetSlotIndex returns the slot index of the most recent block't timeslot
// \tau : The most recent block't timeslot
// (6.2)
func GetSlotIndex(t jamTypes.TimeSlot) jamTypes.TimeSlot {
	return t % jamTypes.TimeSlot(jamTypes.EpochLength)
}
