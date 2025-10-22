package store

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

var (
	headerPrefix         = []byte("h")
	headerHashPrefix     = []byte("hh")
	headerTimeSlotPrefix = []byte("ht")
	extrinsicPrefix      = []byte("e")
)

func headerKey(slot types.TimeSlot, hash types.HeaderHash) []byte {
	encoded, _ := types.NewEncoder().EncodeMany(&slot, &hash)
	return append(headerPrefix, encoded...)
}

func headerHashKey(slot types.TimeSlot) []byte {
	encoded, _ := types.NewEncoder().Encode(&slot)
	return append(headerHashPrefix, encoded...)
}

func headerTimeSlotKey(hash types.HeaderHash) []byte {
	return append(headerTimeSlotPrefix, hash[:]...)
}

func extrinsicKey(slot types.TimeSlot, hash types.HeaderHash) []byte {
	encoded, _ := types.NewEncoder().EncodeMany(&slot, &hash)
	return append(extrinsicPrefix, encoded...)
}
