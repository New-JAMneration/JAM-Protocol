package safrole

import (
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

// R function return the epoch and slot index
// Equation (6.2)
// !Warning : epoch datatype is undefined in jamtypes and is uncertain
func R(time U32) (epoch U32, slotIndex U32) {
	epoch = time / U32(jamTypes.EpochLength)
	slotIndex = time % U32(jamTypes.EpochLength)
	return epoch, slotIndex
}
