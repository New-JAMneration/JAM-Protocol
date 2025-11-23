package repository

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

var (
	headerPrefix              = []byte("h:")
	headerHashPrefix          = []byte("hh:")
	headerTimeSlotPrefix      = []byte("ht:")
	finalizedHeaderHashPrefix = []byte("fh:")

	extrinsicPrefix = []byte("e:")

	stateDataPrefix = []byte("sd:")
)

func headerKey(encoder *types.Encoder, slot types.TimeSlot, hash types.HeaderHash) []byte {
	timeSlotEncoded, _ := encoder.Encode(&slot)
	return append(append(headerPrefix, timeSlotEncoded...), hash[:]...)
}

func canocicalHeaderHashKey(encoder *types.Encoder, slot types.TimeSlot) []byte {
	encoded, _ := encoder.Encode(&slot)
	return append(headerHashPrefix, encoded...)
}

func headerTimeSlotKey(hash types.HeaderHash) []byte {
	return append(headerTimeSlotPrefix, hash[:]...)
}

func extrinsicKey(encoder *types.Encoder, slot types.TimeSlot, hash types.HeaderHash) []byte {
	encoded, _ := encoder.EncodeMany(&slot, &hash)
	return append(extrinsicPrefix, encoded...)
}

func stateRootKey(headerHash types.HeaderHash) []byte {
	return append([]byte("sr:"), headerHash[:]...)
}

func stateDataKey(stateRoot types.StateRoot) []byte {
	return append(stateDataPrefix, stateRoot[:]...)
}
