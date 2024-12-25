package safrole

import (
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// R function return the epoch and slot index
// Equation (6.2)
// !Warning : epoch datatype is undefined in jamtypes and is uncertain
func R(time jamTypes.U32) (epoch jamTypes.U32, slotIndex jamTypes.U32) {
	epoch = time / jamTypes.U32(jamTypes.EpochLength)
	slotIndex = time % jamTypes.U32(jamTypes.EpochLength)
	return epoch, slotIndex
}
