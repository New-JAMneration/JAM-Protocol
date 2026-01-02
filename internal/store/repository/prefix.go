package repository

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

var (
	separator = []byte(":")

	headerPrefix              = []byte("h:")
	headerHashPrefix          = []byte("hh:")
	headerTimeSlotPrefix      = []byte("ht:")
	finalizedHeaderHashPrefix = []byte("fh:")

	extrinsicPrefix = []byte("e:")

	stateRootPrefix = []byte("sr:")
	stateDataPrefix = []byte("sd:")
)

func headerKeyPrefix(encoder *types.Encoder, slot types.TimeSlot) []byte {
	timeSlotEncoded, _ := encoder.Encode(&slot)
	return append(append(headerPrefix, timeSlotEncoded...), separator...)
}

func headerKey(encoder *types.Encoder, slot types.TimeSlot, hash types.HeaderHash) []byte {
	return append(headerKeyPrefix(encoder, slot), hash[:]...)
}

func canocicalHeaderHashKey(encoder *types.Encoder, slot types.TimeSlot) []byte {
	encoded, _ := encoder.Encode(&slot)
	return append(headerHashPrefix, encoded...)
}

func headerTimeSlotKey(hash types.HeaderHash) []byte {
	return append(headerTimeSlotPrefix, hash[:]...)
}

func extrinsicKeyPrefix(encoder *types.Encoder, slot types.TimeSlot) []byte {
	timeSlotEncoded, _ := encoder.Encode(&slot)
	return append(append(extrinsicPrefix, timeSlotEncoded...), separator...)
}

func extrinsicKey(encoder *types.Encoder, slot types.TimeSlot, hash types.HeaderHash) []byte {
	return append(extrinsicKeyPrefix(encoder, slot), hash[:]...)
}

func stateRootKey(headerHash types.HeaderHash) []byte {
	return append(stateRootPrefix, headerHash[:]...)
}

func stateDataKey(stateRoot types.StateRoot) []byte {
	return append(stateDataPrefix, stateRoot[:]...)
}
