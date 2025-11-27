package types

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

const (
	BadSlot               types.ErrorCode = iota // 0 Timeslot value must be strictly monotonic
	UnexpectedTicket                             // 1 Received a ticket while in epoch's tail
	BadTicketOrder                               // 2 Tickets must be sorted
	BadTicketProof                               // 3 Invalid ticket ring proof
	BadTicketAttempt                             // 4 Invalid ticket attempt value
	Reserved                                     // 5 Reserved
	DuplicateTicket                              // 6 Found a ticket duplicate
	VrfSealInvalid                               // 7 VrfSealInvalid
	VrfEntropyInvalid                            // 8 VrfEntropyInvalid
	InvalidEpochMark                             // 9 InvalidEpochMark
	InvalidTicketsMark                           // 10 InvalidTicketsMark
	InvalidOffenderMarker                        // 11 InvalidOffenderMarker
	UnexpectedAuthor                             // 12 Block author is not the expected one
)

// This map provides human-readable messages following the fuzz-proto examples
var SafroleErrorCodeMessages = map[types.ErrorCode]string{
	BadSlot:               "timeslot value must be strictly monotonic",
	UnexpectedTicket:      "received a ticket while in epoch's tail",
	BadTicketOrder:        "tickets must be sorted",
	BadTicketProof:        "invalid ticket ring proof",
	BadTicketAttempt:      "invalid ticket attempt value",
	Reserved:              "reserved",
	DuplicateTicket:       "found a ticket duplicate",
	VrfSealInvalid:        "BadSealSignature", // matches fuzz-proto example
	VrfEntropyInvalid:     "vrf entropy invalid",
	InvalidEpochMark:      "invalid epoch mark",
	InvalidTicketsMark:    "invalid tickets mark",
	InvalidOffenderMarker: "invalid offender marker",
}
